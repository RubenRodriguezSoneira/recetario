package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

func withUser(req *http.Request, userID string) *http.Request {
	if userID == "" {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), appmiddleware.UserIDKey, userID))
}

func TestUserHandler_Profile(t *testing.T) {
	store := newFakeUserStore()
	store.byID["u-1"] = &models.User{ID: "u-1", Email: "me@example.com", Username: "me", Password: "secret-hash"}
	handler := NewUserHandler(store)

	t.Run("authenticated", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodGet, "/api/users/profile", nil), "u-1")
		w := httptest.NewRecorder()

		handler.HandleProfile(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if strings.Contains(w.Body.String(), "secret-hash") {
			t.Error("profile response leaked the password hash")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
		w := httptest.NewRecorder()

		handler.HandleProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})

	t.Run("missing user", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodGet, "/api/users/profile", nil), "ghost")
		w := httptest.NewRecorder()

		handler.HandleProfile(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	store := newFakeUserStore()
	store.byID["u-1"] = &models.User{ID: "u-1", Email: "me@example.com", Username: "me", Password: "secret-hash"}
	handler := NewUserHandler(store)

	t.Run("success", func(t *testing.T) {
		body := `{"first_name":"Ada","last_name":"Lovelace","avatar_url":"/a.png"}`
		req := withUser(httptest.NewRequest(http.MethodPut, "/api/users/profile", strings.NewReader(body)), "u-1")
		w := httptest.NewRecorder()

		handler.HandleUpdateProfile(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
		}
		if store.byID["u-1"].FirstName != "Ada" {
			t.Errorf("first name not persisted: %+v", store.byID["u-1"])
		}
		if strings.Contains(w.Body.String(), "secret-hash") {
			t.Error("update response leaked the password hash")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPut, "/api/users/profile", strings.NewReader("nope")), "u-1")
		w := httptest.NewRecorder()

		handler.HandleUpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/users/profile", strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		handler.HandleUpdateProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})
}
