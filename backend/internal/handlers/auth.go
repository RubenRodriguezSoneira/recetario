package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	appmiddleware "recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// tokenExpiresInSeconds mirrors the AuthService token expiry (24h) reported to clients.
const tokenExpiresInSeconds = 86400

func isUniqueUserConflict(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "unique constraint failed: users.email") ||
		strings.Contains(errMsg, "unique constraint failed: users.username")
}

func isValidEmailFormat(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func buildAuthCookie(r *http.Request, token string, maxAge int) *http.Cookie {
	secure := false
	if r.TLS != nil {
		secure = true
	}
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); strings.EqualFold(forwardedProto, "https") {
		secure = true
	}

	return &http.Cookie{
		Name:     appmiddleware.AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}

func clearAuthCookie(r *http.Request) *http.Cookie {
	cookie := buildAuthCookie(r, "", -1)
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

// UserStore describes the user data-access methods the auth and user handlers
// depend on. The concrete *repositories.UserRepository satisfies it; tests inject
// a fake.
type UserStore interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
	UpdateUser(user *models.User) error
}

type AuthHandler struct {
	authService *appmiddleware.AuthService
	users       UserStore
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresIn int64  `json:"expires_in"`
}

// User is the public representation of a user (never includes the password hash).
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func toPublicUser(u *models.User) User {
	return User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

func NewAuthHandler(authService *appmiddleware.AuthService, users UserStore) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		users:       users,
	}
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Info("Register failed: invalid request body", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_INVALID_BODY", "Invalid request payload", "No pudimos procesar la solicitud. Revisa los datos e inténtalo de nuevo.")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if req.Email == "" {
		log.Info("Register failed: email required")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid registration data", "El email es obligatorio.")
		return
	}
	if !isValidEmailFormat(req.Email) {
		log.Info("Register failed: invalid email format")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid registration data", "El formato del email no es válido.")
		return
	}
	if req.Username == "" {
		log.Info("Register failed: username required")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid registration data", "El nombre de usuario es obligatorio.")
		return
	}
	if req.Password == "" {
		log.Info("Register failed: password required")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid registration data", "La contraseña es obligatoria.")
		return
	}
	if len(req.Password) < 8 {
		log.Info("Register failed: password too short")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid registration data", "La contraseña debe tener al menos 8 caracteres.")
		return
	}

	emailTaken, err := h.users.EmailExists(req.Email)
	if err != nil {
		log.Error("Failed to check email existence", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}
	usernameTaken, err := h.users.UsernameExists(req.Username)
	if err != nil {
		log.Error("Failed to check username existence", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}
	if emailTaken || usernameTaken {
		log.Info("Register conflict: email or username already exists", "email", req.Email, "username", req.Username)
		appmiddleware.WriteJSONError(w, http.StatusConflict, "USER_ALREADY_EXISTS", "Registration conflict", "El email o nombre de usuario ya está en uso.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Password hashing failed", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}

	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  string(hash),
	}
	if err := h.users.CreateUser(user); err != nil {
		if isUniqueUserConflict(err) {
			log.Info("Register conflict: unique constraint", "email", req.Email, "username", req.Username, "error", err)
			appmiddleware.WriteJSONError(w, http.StatusConflict, "USER_ALREADY_EXISTS", "Registration conflict", "El email o nombre de usuario ya está en uso.")
			return
		}
		log.Error("Failed to create user", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}

	http.SetCookie(w, buildAuthCookie(r, token, tokenExpiresInSeconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		User:      toPublicUser(user),
		ExpiresIn: tokenExpiresInSeconds,
	}); err != nil {
		log.Error("Failed to encode register response", "error", fmt.Errorf("failed to encode register response: %w", err))
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Info("Login failed: invalid request body", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_INVALID_BODY", "Invalid request payload", "No pudimos procesar la solicitud. Revisa los datos e inténtalo de nuevo.")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		log.Info("Login failed: email required")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid login data", "El email es obligatorio.")
		return
	}
	if !isValidEmailFormat(req.Email) {
		log.Info("Login failed: invalid email format")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid login data", "El formato del email no es válido.")
		return
	}
	if req.Password == "" {
		log.Info("Login failed: password required")
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "AUTH_VALIDATION_FAILED", "Invalid login data", "La contraseña es obligatoria.")
		return
	}

	user, err := h.users.GetUserByEmail(req.Email)
	if err != nil {
		// Do not reveal whether the email exists; log server-side only.
		log.Info("Login failed: user lookup", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Authentication failed", "Credenciales inválidas.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Info("Login failed: invalid credentials")
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Authentication failed", "Credenciales inválidas.")
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}

	http.SetCookie(w, buildAuthCookie(r, token, tokenExpiresInSeconds))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		User:      toPublicUser(user),
		ExpiresIn: tokenExpiresInSeconds,
	}); err != nil {
		log.Error("Failed to encode login response", "error", fmt.Errorf("failed to encode login response: %w", err))
	}
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	logger.FromContext(r.Context()).Info("User logout request")
	http.SetCookie(w, clearAuthCookie(r))

	if strings.Contains(strings.ToLower(r.Header.Get("Accept")), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Logout successful",
		})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "AUTH_HEADER_REQUIRED", "Authentication required", "Debes iniciar sesión para continuar.")
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Invalid authorization format", "Tu sesión no es válida. Inicia sesión nuevamente.")
		return
	}

	claims, err := h.authService.ValidateToken(tokenParts[1])
	if err != nil {
		log.Info("Token validation failed", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token", "Tu sesión expiró o no es válida. Inicia sesión nuevamente.")
		return
	}

	newToken, err := h.authService.GenerateToken(claims.UserID, claims.Email, claims.IsAdmin)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Ocurrió un error interno. Inténtalo más tarde.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      newToken,
		"expires_in": tokenExpiresInSeconds,
	})
}
