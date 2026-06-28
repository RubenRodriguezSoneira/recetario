package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-app/internal/models"
)

type fakeTagStore struct {
	recipes    map[string]*models.Recipe
	allTags    []string
	recipeTags map[string][]string
	added      []string
	deleted    []string
	addErr     error
	delErr     error
}

func newTagStore() *fakeTagStore {
	return &fakeTagStore{
		recipes: map[string]*models.Recipe{
			"1": {ID: "1", UserID: testOwner, Title: "Cake"},
			"2": {ID: "2", UserID: "other", Title: "Soup"},
		},
		allTags:    []string{"dessert", "italian", "quick"},
		recipeTags: map[string][]string{"1": {"dessert"}},
	}
}

func (f *fakeTagStore) GetRecipe(id string) (*models.Recipe, error) {
	if r, ok := f.recipes[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("recipe not found")
}

func (f *fakeTagStore) GetAllTags(search string) ([]string, error) {
	return f.allTags, nil
}

func (f *fakeTagStore) GetRecipeTags(recipeID string) ([]string, error) {
	return f.recipeTags[recipeID], nil
}

func (f *fakeTagStore) AddRecipeTag(recipeID, tag string) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.added = append(f.added, tag)
	f.recipeTags[recipeID] = append(f.recipeTags[recipeID], tag)
	return nil
}

func (f *fakeTagStore) DeleteRecipeTag(recipeID, tag string) error {
	if f.delErr != nil {
		return f.delErr
	}
	f.deleted = append(f.deleted, tag)
	return nil
}

func TestTagHandler_ListAll(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	w := httptest.NewRecorder()

	handler.HandleListAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var tags []string
	if err := json.Unmarshal(w.Body.Bytes(), &tags); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

func TestTagHandler_ListForRecipe(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1/tags", nil)
	req = withParams(req, map[string]string{"id": "1"}, "")
	w := httptest.NewRecorder()

	handler.HandleListForRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var tags []string
	if err := json.Unmarshal(w.Body.Bytes(), &tags); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(tags) != 1 || tags[0] != "dessert" {
		t.Errorf("Expected [dessert], got %v", tags)
	}
}

func TestTagHandler_ListForRecipe_NotFound(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/999/tags", nil)
	req = withParams(req, map[string]string{"id": "999"}, "")
	w := httptest.NewRecorder()

	handler.HandleListForRecipe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTagHandler_Add(t *testing.T) {
	store := newTagStore()
	handler := NewTagHandler(store)

	body := strings.NewReader(`{"tags":["italian","quick"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/tags", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleAdd(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}
	if len(store.added) != 2 {
		t.Errorf("Expected 2 tags added, got %v", store.added)
	}
}

func TestTagHandler_Add_SingleField(t *testing.T) {
	store := newTagStore()
	handler := NewTagHandler(store)

	body := strings.NewReader(`{"tag":"vegan"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/tags", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleAdd(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}
	if len(store.added) != 1 || store.added[0] != "vegan" {
		t.Errorf("Expected [vegan] added, got %v", store.added)
	}
}

func TestTagHandler_Add_Empty(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/tags", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleAdd(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty tags, got %d", w.Code)
	}
}

func TestTagHandler_Add_Unauthorized(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	body := strings.NewReader(`{"tag":"vegan"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/tags", body)
	req = withParams(req, map[string]string{"id": "1"}, "")
	w := httptest.NewRecorder()

	handler.HandleAdd(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestTagHandler_Add_NonOwner(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	body := strings.NewReader(`{"tag":"vegan"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/tags", body)
	req = withParams(req, map[string]string{"id": "1"}, "intruder")
	w := httptest.NewRecorder()

	handler.HandleAdd(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-owner, got %d", w.Code)
	}
}

func TestTagHandler_Delete(t *testing.T) {
	store := newTagStore()
	handler := NewTagHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1/tags/dessert", nil)
	req = withParams(req, map[string]string{"id": "1", "tag": "dessert"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", w.Code)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "dessert" {
		t.Errorf("Expected [dessert] deleted, got %v", store.deleted)
	}
}

func TestTagHandler_Delete_NonOwner(t *testing.T) {
	handler := NewTagHandler(newTagStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1/tags/dessert", nil)
	req = withParams(req, map[string]string{"id": "1", "tag": "dessert"}, "intruder")
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-owner, got %d", w.Code)
	}
}
