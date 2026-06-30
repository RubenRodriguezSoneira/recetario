package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	appmiddleware "recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// tokenExpiresInSeconds mirrors the AuthService token expiry (24h) reported to clients.
const tokenExpiresInSeconds = 86400

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Username == "" || len(req.Password) < 8 {
		http.Error(w, "email, username and password (min 8 chars) are required", http.StatusBadRequest)
		return
	}

	emailTaken, err := h.users.EmailExists(req.Email)
	if err != nil {
		log.Error("Failed to check email existence", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	usernameTaken, err := h.users.UsernameExists(req.Username)
	if err != nil {
		log.Error("Failed to check username existence", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if emailTaken || usernameTaken {
		http.Error(w, "email or username already in use", http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Password hashing failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		log.Error("Failed to create user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, buildAuthCookie(r, token, tokenExpiresInSeconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		User:      toPublicUser(user),
		ExpiresIn: tokenExpiresInSeconds,
	})
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetUserByEmail(req.Email)
	if err != nil {
		// Do not reveal whether the email exists; log server-side only.
		log.Info("Login failed: user lookup", "error", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, false)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, buildAuthCookie(r, token, tokenExpiresInSeconds))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     token,
		User:      toPublicUser(user),
		ExpiresIn: tokenExpiresInSeconds,
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, clearAuthCookie(r))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}

	claims, err := h.authService.ValidateToken(tokenParts[1])
	if err != nil {
		log.Info("Token validation failed", "error", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	newToken, err := h.authService.GenerateToken(claims.UserID, claims.Email, claims.IsAdmin)
	if err != nil {
		log.Error("Token generation failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      newToken,
		"expires_in": tokenExpiresInSeconds,
	})
}
