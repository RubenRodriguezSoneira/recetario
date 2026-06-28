package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

// fakeCollectionStore is an in-memory CollectionStore for handler tests.
type fakeCollectionStore struct {
	byID      map[string]*models.RecipeCollection
	createErr error
	updateErr error
	deleteErr error
	addErr    error
	removeErr error
	deleted   []string
	added     [][2]string
	removed   [][2]string
}

func (f *fakeCollectionStore) CreateCollection(c *models.RecipeCollection) error {
	if f.createErr != nil {
		return f.createErr
	}
	c.ID = "col-new"
	return nil
}

func (f *fakeCollectionStore) GetCollection(id string) (*models.RecipeCollection, error) {
	if c, ok := f.byID[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("collection not found")
}

func (f *fakeCollectionStore) GetCollectionsByUser(userID string) ([]*models.RecipeCollection, error) {
	var out []*models.RecipeCollection
	for _, c := range f.byID {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeCollectionStore) UpdateCollection(c *models.RecipeCollection) error {
	return f.updateErr
}

func (f *fakeCollectionStore) DeleteCollection(id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeCollectionStore) AddRecipe(collectionID, recipeID string) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.added = append(f.added, [2]string{collectionID, recipeID})
	return nil
}

func (f *fakeCollectionStore) RemoveRecipe(collectionID, recipeID string) error {
	if f.removeErr != nil {
		return f.removeErr
	}
	f.removed = append(f.removed, [2]string{collectionID, recipeID})
	return nil
}

func newTestCollectionStore() *fakeCollectionStore {
	byID := map[string]*models.RecipeCollection{
		"c1": {ID: "c1", UserID: testOwner, Name: "Favorites", IsPublic: false},
		"c2": {ID: "c2", UserID: "stranger", Name: "Public picks", IsPublic: true},
		"c3": {ID: "c3", UserID: "stranger", Name: "Private stash", IsPublic: false},
	}
	return &fakeCollectionStore{byID: byID}
}

// withCollectionCtx attaches chi route params and (optionally) an authenticated
// user id to the request context.
func withCollectionCtx(req *http.Request, id, recipeID, userID string) *http.Request {
	rctx := chi.NewRouteContext()
	if id != "" {
		rctx.URLParams.Add("id", id)
	}
	if recipeID != "" {
		rctx.URLParams.Add("recipeID", recipeID)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	if userID != "" {
		ctx = context.WithValue(ctx, appmiddleware.UserIDKey, userID)
	}
	return req.WithContext(ctx)
}

func TestCollectionHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		userID     string
		wantStatus int
	}{
		{"valid", `{"name":"Weeknight dinners","is_public":true}`, testOwner, http.StatusCreated},
		{"missing name", `{"description":"no name"}`, testOwner, http.StatusBadRequest},
		{"invalid json", `not json`, testOwner, http.StatusBadRequest},
		{"unauthenticated", `{"name":"x"}`, "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewCollectionHandler(newTestCollectionStore())
			req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(tc.body))
			req = withCollectionCtx(req, "", "", tc.userID)
			w := httptest.NewRecorder()

			handler.HandleCreate(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if resp["user_id"] != testOwner {
					t.Errorf("user_id = %v, want %s", resp["user_id"], testOwner)
				}
				if resp["id"] != "col-new" {
					t.Errorf("id = %v, want col-new", resp["id"])
				}
			}
		})
	}
}

func TestCollectionHandler_List(t *testing.T) {
	handler := NewCollectionHandler(newTestCollectionStore())
	req := httptest.NewRequest(http.MethodGet, "/api/collections", nil)
	req = withCollectionCtx(req, "", "", testOwner)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("got %d collections, want 1 (only the owner's)", len(resp))
	}
	if resp[0]["id"] != "c1" {
		t.Errorf("id = %v, want c1", resp[0]["id"])
	}
}

func TestCollectionHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		userID     string
		wantStatus int
	}{
		{"own private", "c1", testOwner, http.StatusOK},
		{"foreign public", "c2", testOwner, http.StatusOK},
		{"foreign private hidden", "c3", testOwner, http.StatusNotFound},
		{"missing", "nope", testOwner, http.StatusNotFound},
		{"unauthenticated", "c1", "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewCollectionHandler(newTestCollectionStore())
			req := httptest.NewRequest(http.MethodGet, "/api/collections/"+tc.id, nil)
			req = withCollectionCtx(req, tc.id, "", tc.userID)
			w := httptest.NewRecorder()

			handler.HandleGet(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestCollectionHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		userID     string
		wantStatus int
	}{
		{"owner updates", "c1", `{"name":"Renamed"}`, testOwner, http.StatusOK},
		{"foreign forbidden as 404", "c3", `{"name":"hijack"}`, testOwner, http.StatusNotFound},
		{"invalid body", "c1", `nope`, testOwner, http.StatusBadRequest},
		{"empty name rejected", "c1", `{"name":""}`, testOwner, http.StatusBadRequest},
		{"unauthenticated", "c1", `{"name":"x"}`, "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewCollectionHandler(newTestCollectionStore())
			req := httptest.NewRequest(http.MethodPut, "/api/collections/"+tc.id, strings.NewReader(tc.body))
			req = withCollectionCtx(req, tc.id, "", tc.userID)
			w := httptest.NewRecorder()

			handler.HandleUpdate(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

func TestCollectionHandler_Delete(t *testing.T) {
	store := newTestCollectionStore()
	handler := NewCollectionHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/collections/c1", nil)
	req = withCollectionCtx(req, "c1", "", testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "c1" {
		t.Errorf("deleted = %v, want [c1]", store.deleted)
	}
}

func TestCollectionHandler_Delete_Foreign(t *testing.T) {
	store := newTestCollectionStore()
	handler := NewCollectionHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/collections/c3", nil)
	req = withCollectionCtx(req, "c3", "", testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if len(store.deleted) != 0 {
		t.Errorf("foreign collection must not be deleted, got %v", store.deleted)
	}
}

func TestCollectionHandler_AddRecipe(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		userID     string
		wantStatus int
	}{
		{"owner adds", "c1", `{"recipe_id":"r1"}`, testOwner, http.StatusNoContent},
		{"missing recipe_id", "c1", `{}`, testOwner, http.StatusBadRequest},
		{"invalid body", "c1", `nope`, testOwner, http.StatusBadRequest},
		{"foreign 404", "c3", `{"recipe_id":"r1"}`, testOwner, http.StatusNotFound},
		{"unauthenticated", "c1", `{"recipe_id":"r1"}`, "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestCollectionStore()
			handler := NewCollectionHandler(store)
			req := httptest.NewRequest(http.MethodPost, "/api/collections/"+tc.id+"/recipes", strings.NewReader(tc.body))
			req = withCollectionCtx(req, tc.id, "", tc.userID)
			w := httptest.NewRecorder()

			handler.HandleAddRecipe(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusNoContent {
				if len(store.added) != 1 || store.added[0] != [2]string{"c1", "r1"} {
					t.Errorf("added = %v, want [[c1 r1]]", store.added)
				}
			}
		})
	}
}

func TestCollectionHandler_RemoveRecipe(t *testing.T) {
	store := newTestCollectionStore()
	handler := NewCollectionHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/collections/c1/recipes/r9", nil)
	req = withCollectionCtx(req, "c1", "r9", testOwner)
	w := httptest.NewRecorder()

	handler.HandleRemoveRecipe(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if len(store.removed) != 1 || store.removed[0] != [2]string{"c1", "r9"} {
		t.Errorf("removed = %v, want [[c1 r9]]", store.removed)
	}
}

func TestCollectionHandler_RemoveRecipe_Foreign(t *testing.T) {
	store := newTestCollectionStore()
	handler := NewCollectionHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/collections/c3/recipes/r9", nil)
	req = withCollectionCtx(req, "c3", "r9", testOwner)
	w := httptest.NewRecorder()

	handler.HandleRemoveRecipe(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if len(store.removed) != 0 {
		t.Errorf("foreign removal must not happen, got %v", store.removed)
	}
}
