package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

func TestWebHandler_getUserFromContext(t *testing.T) {
	store := newFakeUserStore()
	store.byID["u-chef"] = &models.User{
		ID:       "u-chef",
		Email:    "chef@example.com",
		Username: "chef",
	}

	h := NewWebHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	claims := &appmiddleware.Claims{UserID: "u-chef", Email: "chef@example.com"}
	ctx := context.WithValue(req.Context(), appmiddleware.UserClaimsKey, claims)
	req = req.WithContext(ctx)

	user := h.getUserFromContext(req)
	if user == nil {
		t.Fatal("expected user from context")
	}
	if user.ID != "u-chef" {
		t.Fatalf("got user ID %q, want u-chef", user.ID)
	}
}

func TestWebHandler_getUserFromContext_MissingOrInvalid(t *testing.T) {
	store := newFakeUserStore()
	h := NewWebHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := h.getUserFromContext(req); got != nil {
		t.Fatal("expected nil user without claims")
	}

	claims := &appmiddleware.Claims{UserID: "missing-id"}
	ctx := context.WithValue(req.Context(), appmiddleware.UserClaimsKey, claims)
	req = req.WithContext(ctx)
	if got := h.getUserFromContext(req); got != nil {
		t.Fatal("expected nil user for unresolved ID")
	}
}

func TestWebHandler_GuardsProtectedPagesWithoutSession(t *testing.T) {
	h := NewWebHandler(newFakeUserStore())

	tests := []struct {
		name   string
		path   string
		method string
		call   func(http.ResponseWriter, *http.Request)
	}{
		{
			name:   "new recipe guard",
			path:   "/recipes/new",
			method: http.MethodGet,
			call:   h.HandleNewRecipe,
		},
		{
			name:   "edit recipe guard",
			path:   "/recipes/1/edit",
			method: http.MethodGet,
			call:   h.HandleEditRecipe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			tt.call(w, req)

			if w.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
			}
			if loc := w.Header().Get("Location"); loc != "/" {
				t.Fatalf("Location = %q, want /", loc)
			}
		})
	}
}
