package appmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthService_GenerateToken(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)

	tests := []struct {
		name    string
		userID  string
		email   string
		isAdmin bool
		wantErr bool
	}{
		{
			name:    "Valid token generation",
			userID:  "1",
			email:   "test@example.com",
			isAdmin: false,
			wantErr: false,
		},
		{
			name:    "Admin token generation",
			userID:  "2",
			email:   "admin@example.com",
			isAdmin: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.userID, tt.email, tt.isAdmin)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			if !tt.wantErr {
				claims, err := auth.ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token validation failed: %v", err)
				}

				if claims.UserID != tt.userID {
					t.Errorf("Expected UserID %s, got %s", tt.userID, claims.UserID)
				}

				if claims.Email != tt.email {
					t.Errorf("Expected Email %s, got %s", tt.email, claims.Email)
				}

				if claims.IsAdmin != tt.isAdmin {
					t.Errorf("Expected IsAdmin %v, got %v", tt.isAdmin, claims.IsAdmin)
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)

	validToken, _ := auth.GenerateToken("1", "test@example.com", false)
	adminToken, _ := auth.GenerateToken("2", "admin@example.com", true)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantAdmin bool
		wantID    string
	}{
		{
			name:      "Valid user token",
			token:     validToken,
			wantErr:   false,
			wantAdmin: false,
			wantID:    "1",
		},
		{
			name:      "Valid admin token",
			token:     adminToken,
			wantErr:   false,
			wantAdmin: true,
			wantID:    "2",
		},
		{
			name:    "Invalid token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims.UserID != tt.wantID {
					t.Errorf("Expected UserID %s, got %s", tt.wantID, claims.UserID)
				}

				if claims.IsAdmin != tt.wantAdmin {
					t.Errorf("Expected IsAdmin %v, got %v", tt.wantAdmin, claims.IsAdmin)
				}
			}
		})
	}
}

func TestAuthService_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)
	auth.TokenExpiry = -1 * time.Hour // Expired 1 hour ago

	expiredToken, err := auth.GenerateToken("1", "test@example.com", false)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	_, err = auth.ValidateToken(expiredToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)

	validToken, _ := auth.GenerateToken("1", "test@example.com", false)
	otherToken, _ := auth.GenerateToken("2", "other@example.com", false)

	tests := []struct {
		name           string
		authHeader     string
		cookieToken    string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "Valid authorization header",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: "1",
		},
		{
			name:           "Valid auth cookie",
			cookieToken:    validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: "1",
		},
		{
			name:           "Header has priority over cookie",
			authHeader:     "Bearer " + otherToken,
			cookieToken:    validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: "2",
		},
		{
			name:           "Missing credentials",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid header format does not fallback to cookie",
			authHeader:     "InvalidTokenFormat",
			cookieToken:    validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid cookie token",
			cookieToken:    "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: tt.cookieToken})
			}

			w := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := GetUserID(r.Context())
				if tt.expectedUserID == "" && ok {
					t.Error("Unexpected user ID in context")
				}
				if tt.expectedUserID != "" {
					if !ok {
						t.Error("Expected user ID in context")
					}
					if userID != tt.expectedUserID {
						t.Errorf("Expected user ID %s, got %s", tt.expectedUserID, userID)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			auth.AuthMiddleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)

	token, _ := auth.GenerateToken("1", "test@example.com", false)
	otherToken, _ := auth.GenerateToken("2", "other@example.com", false)

	tests := []struct {
		name           string
		authHeader     string
		cookieToken    string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "Valid authorization header",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedUserID: "1",
		},
		{
			name:           "Cookie fallback when no header",
			cookieToken:    token,
			expectedStatus: http.StatusOK,
			expectedUserID: "1",
		},
		{
			name:           "Header priority over cookie",
			authHeader:     "Bearer " + otherToken,
			cookieToken:    token,
			expectedStatus: http.StatusOK,
			expectedUserID: "2",
		},
		{
			name:           "Missing credentials",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token should still pass",
			cookieToken:    "invalid-token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: tt.cookieToken})
			}

			w := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := GetUserID(r.Context())
				if tt.expectedUserID == "" && ok {
					t.Error("Unexpected user ID in context")
				}
				if tt.expectedUserID != "" {
					if !ok {
						t.Error("Expected user ID in context")
					}
					if userID != tt.expectedUserID {
						t.Errorf("Expected user ID %s, got %s", tt.expectedUserID, userID)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			auth.OptionalAuthMiddleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	secret := "test-secret-key"
	auth := NewAuthService(secret)

	userToken, _ := auth.GenerateToken("1", "user@example.com", false)
	adminToken, _ := auth.GenerateToken("2", "admin@example.com", true)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Admin user",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular user",
			authHeader:     "Bearer " + userToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := auth.AuthMiddleware(RequireAdmin(next))
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetUserClaims(t *testing.T) {
	ctx := context.Background()

	claims := &Claims{
		UserID:  "1",
		Email:   "test@example.com",
		IsAdmin: false,
	}

	withClaims := context.WithValue(ctx, UserClaimsKey, claims)

	tests := []struct {
		name     string
		ctx      context.Context
		want     *Claims
		wantBool bool
	}{
		{
			name:     "Context with claims",
			ctx:      withClaims,
			want:     claims,
			wantBool: true,
		},
		{
			name:     "Context without claims",
			ctx:      ctx,
			want:     nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetUserClaims(tt.ctx)
			if ok != tt.wantBool {
				t.Errorf("GetUserClaims() ok = %v, want %v", ok, tt.wantBool)
				return
			}
			if !tt.wantBool && got != nil {
				t.Errorf("GetUserClaims() got = %v, want %v", got, tt.want)
			}
			if tt.wantBool && got != tt.want {
				t.Errorf("GetUserClaims() got = %v, want %v", got, tt.want)
			}
		})
	}
}
