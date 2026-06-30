package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

// fakeUserStore is an in-memory UserStore for auth/user handler tests.
type fakeUserStore struct {
	byEmail   map[string]*models.User
	byID      map[string]*models.User
	usernames map[string]bool
	createErr error
	updateErr error
	created   []*models.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		byEmail:   map[string]*models.User{},
		byID:      map[string]*models.User{},
		usernames: map[string]bool{},
	}
}

func (f *fakeUserStore) CreateUser(user *models.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	if user.ID == "" {
		user.ID = "u-" + user.Username
	}
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	f.usernames[user.Username] = true
	f.created = append(f.created, user)
	return nil
}

func (f *fakeUserStore) GetUserByID(id string) (*models.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (f *fakeUserStore) GetUserByEmail(email string) (*models.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (f *fakeUserStore) EmailExists(email string) (bool, error) {
	_, ok := f.byEmail[email]
	return ok, nil
}

func (f *fakeUserStore) UsernameExists(username string) (bool, error) {
	return f.usernames[username], nil
}

func (f *fakeUserStore) UpdateUser(user *models.User) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.byID[user.ID] = user
	f.byEmail[user.Email] = user
	return nil
}

func seedUser(t *testing.T, store *fakeUserStore, email, username, password string) *models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	u := &models.User{
		ID:       "u-" + username,
		Email:    email,
		Username: username,
		Password: string(hash),
	}
	store.byEmail[email] = u
	store.byID[u.ID] = u
	store.usernames[username] = true
	return u
}

func newAuthHandler(store UserStore) *AuthHandler {
	return NewAuthHandler(appmiddleware.NewAuthService("test-secret"), store)
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		seedEmail  string
		wantStatus int
	}{
		{"valid", `{"email":"new@example.com","username":"newbie","password":"supersecret"}`, "", http.StatusCreated},
		{"duplicate email", `{"email":"dup@example.com","username":"other","password":"supersecret"}`, "dup@example.com", http.StatusConflict},
		{"short password", `{"email":"x@example.com","username":"x","password":"short"}`, "", http.StatusBadRequest},
		{"missing email", `{"username":"x","password":"supersecret"}`, "", http.StatusBadRequest},
		{"invalid json", `not json`, "", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeUserStore()
			if tc.seedEmail != "" {
				seedUser(t, store, tc.seedEmail, "dupuser", "whatever1")
			}
			handler := newAuthHandler(store)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			handler.HandleRegister(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var resp AuthResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if resp.Token == "" {
					t.Error("expected a token")
				}
				if resp.User.Email != "new@example.com" {
					t.Errorf("email = %q, want new@example.com", resp.User.Email)
				}
				// The stored password must be a bcrypt hash, never the plaintext.
				if len(store.created) != 1 || store.created[0].Password == "supersecret" {
					t.Error("password was not hashed before storage")
				}

				resCookies := w.Result().Cookies()
				foundAuthCookie := false
				for _, c := range resCookies {
					if c.Name == appmiddleware.AuthCookieName {
						foundAuthCookie = true
						if !c.HttpOnly {
							t.Error("expected auth cookie to be HttpOnly")
						}
						if c.SameSite != http.SameSiteLaxMode {
							t.Errorf("expected SameSite Lax, got %v", c.SameSite)
						}
					}
				}
				if !foundAuthCookie {
					t.Error("expected auth cookie in register response")
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid", `{"email":"chef@example.com","password":"password123"}`, http.StatusOK},
		{"wrong password", `{"email":"chef@example.com","password":"wrongpass"}`, http.StatusUnauthorized},
		{"unknown email", `{"email":"ghost@example.com","password":"password123"}`, http.StatusUnauthorized},
		{"missing fields", `{"email":"chef@example.com"}`, http.StatusBadRequest},
		{"invalid json", `nope`, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeUserStore()
			seedUser(t, store, "chef@example.com", "chef", "password123")
			handler := newAuthHandler(store)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			handler.HandleLogin(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				var resp AuthResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if resp.Token == "" {
					t.Error("expected a token")
				}

				resCookies := w.Result().Cookies()
				foundAuthCookie := false
				for _, c := range resCookies {
					if c.Name == appmiddleware.AuthCookieName {
						foundAuthCookie = true
						if c.Value == "" {
							t.Error("expected non-empty auth cookie value")
						}
					}
				}
				if !foundAuthCookie {
					t.Error("expected auth cookie in login response")
				}
			}
		})
	}
}

func TestAuthHandler_Login_NoBackdoor(t *testing.T) {
	// The removed backdoor used to log in any request whose password contained
	// the substring "password". Verify that is no longer the case.
	store := newFakeUserStore()
	seedUser(t, store, "chef@example.com", "chef", "the-real-secret")
	handler := newAuthHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"chef@example.com","password":"password"}`))
	w := httptest.NewRecorder()

	handler.HandleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (backdoor must be gone)", w.Code)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	handler := newAuthHandler(newFakeUserStore())
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	handler.HandleLogout(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	if location := w.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}

	resCookies := w.Result().Cookies()
	foundAuthCookie := false
	for _, c := range resCookies {
		if c.Name == appmiddleware.AuthCookieName {
			foundAuthCookie = true
			if c.MaxAge >= 0 {
				t.Fatalf("expected cookie MaxAge < 0 to clear, got %d", c.MaxAge)
			}
		}
	}
	if !foundAuthCookie {
		t.Fatal("expected auth cookie to be cleared on logout")
	}
}

func TestAuthHandler_Register_UniqueConstraintConflict(t *testing.T) {
	store := newFakeUserStore()
	store.createErr = fmt.Errorf("failed to create user: constraint failed: UNIQUE constraint failed: users.email (2067)")
	handler := newAuthHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"email":"dup@example.com","username":"dup","password":"password123"}`))
	w := httptest.NewRecorder()

	handler.HandleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}
