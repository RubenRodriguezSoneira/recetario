package appmiddleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"recipe-app/internal/logger"
)

type Claims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

type AuthService struct {
	JWTSecret     []byte
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{
		JWTSecret:     []byte(secret),
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
}

func (a *AuthService) GenerateToken(userID string, email string, isAdmin bool) (string, error) {
	claims := &Claims{
		UserID:  userID,
		Email:   email,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "recipe-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.JWTSecret)
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
	UserIDKey     contextKey = "user_id"
)

func (a *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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

		claims, err := a.ValidateToken(tokenParts[1])
		if err != nil {
			logger.LogError(ctx, err, "Token validation failed")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, UserClaimsKey, claims)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *AuthService) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				if claims, err := a.ValidateToken(tokenParts[1]); err == nil {
					ctx = context.WithValue(ctx, UserClaimsKey, claims)
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetUserClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*Claims)
	return claims, ok
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, ok := GetUserClaims(ctx)
		if !ok || !claims.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
