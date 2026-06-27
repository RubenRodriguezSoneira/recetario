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
	"recipe-app/internal/models"
)

// fakeRecipeStore is an in-memory RecipeStore used to exercise the API handlers
// without a database (the sqlite driver requires cgo and the queries use the
// PostgreSQL dialect, so unit tests inject canned data instead).
type fakeRecipeStore struct {
	recipes   []*models.Recipe
	byID      map[string]*models.Recipe
	createErr error
}

func (f *fakeRecipeStore) GetRecipes(limit, offset int, search, difficulty string, maxCookTime int) ([]*models.Recipe, error) {
	return f.recipes, nil
}

func (f *fakeRecipeStore) GetRecipe(id string) (*models.Recipe, error) {
	if r, ok := f.byID[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("recipe not found")
}

func (f *fakeRecipeStore) CreateRecipe(recipe *models.Recipe) error {
	if f.createErr != nil {
		return f.createErr
	}
	recipe.ID = "3"
	return nil
}

func newTestStore() *fakeRecipeStore {
	recipes := []*models.Recipe{
		{ID: "1", Title: "Spaghetti Bolognese", Description: "Classic Italian pasta dish with rich meat sauce", CookTime: 30, Difficulty: "medium"},
		{ID: "2", Title: "Chicken Curry", Description: "Spicy and aromatic Indian curry with tender chicken", CookTime: 45, Difficulty: "hard"},
		{ID: "3", Title: "Caesar Salad", Description: "Fresh romaine lettuce with creamy Caesar dressing", CookTime: 15, Difficulty: "easy"},
		{ID: "4", Title: "Beef Tacos", Description: "Mexican-style tacos with seasoned ground beef", CookTime: 25, Difficulty: "medium"},
		{ID: "5", Title: "Chocolate Cake", Description: "Rich and moist chocolate cake with fudge frosting", CookTime: 60, Difficulty: "hard"},
		{ID: "6", Title: "Greek Salad", Description: "Mediterranean salad with feta cheese and olives", CookTime: 10, Difficulty: "easy"},
	}
	byID := make(map[string]*models.Recipe, len(recipes))
	for _, r := range recipes {
		byID[r.ID] = r
	}
	return &fakeRecipeStore{recipes: recipes, byID: byID}
}

func TestAPIHandler_GetRecipes(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   []map[string]interface{}
	}{
		{
			name:           "GET recipes returns 200",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{"id": "1", "title": "Spaghetti Bolognese"},
				{"id": "2", "title": "Chicken Curry"},
				{"id": "3", "title": "Caesar Salad"},
				{"id": "4", "title": "Beef Tacos"},
				{"id": "5", "title": "Chocolate Cake"},
				{"id": "6", "title": "Greek Salad"},
			},
		},
		{
			name:           "POST to recipes endpoint handled by createRecipe method",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *strings.Reader
			if tt.method == http.MethodPost {
				body = strings.NewReader(`{"title":"New Recipe","cook_time":10,"difficulty":"easy"}`)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, "/api/recipes", body)
			w := httptest.NewRecorder()

			handler.HandleRecipes(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == http.MethodGet {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if len(response) != len(tt.expectedBody) {
					t.Errorf("Expected %d recipes, got %d", len(tt.expectedBody), len(response))
				}

				for i, expected := range tt.expectedBody {
					if response[i]["id"] != expected["id"] {
						t.Errorf("Expected recipe ID %s, got %s", expected["id"], response[i]["id"])
					}
					if response[i]["title"] != expected["title"] {
						t.Errorf("Expected recipe title %s, got %s", expected["title"], response[i]["title"])
					}
				}
			}
		})
	}
}

func TestAPIHandler_CreateRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := strings.NewReader(`{"title":"New Recipe","cook_time":10,"difficulty":"easy"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes", body)
	w := httptest.NewRecorder()

	handler.HandleCreateRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Recipe created successfully" {
		t.Errorf("Expected message 'Recipe created successfully', got %s", response["message"])
	}

	if response["id"] != "3" {
		t.Errorf("Expected id '3', got %s", response["id"])
	}
}

func TestAPIHandler_CreateRecipe_InvalidBody(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.HandleCreateRecipe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid body, got %d", w.Code)
	}
}

func TestAPIHandler_GetRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.HandleRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "1" {
		t.Errorf("Expected recipe ID '1', got %s", response["id"])
	}
	if response["title"] != "Spaghetti Bolognese" {
		t.Errorf("Expected recipe title 'Spaghetti Bolognese', got %s", response["title"])
	}
}

func TestAPIHandler_GetRecipe_NotFound(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.HandleRecipe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for missing recipe, got %d", w.Code)
	}
}

func TestAPIHandler_UpdateRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.HandleUpdateRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Recipe updated successfully" {
		t.Errorf("Expected message 'Recipe updated successfully', got %s", response["message"])
	}
}

func TestAPIHandler_DeleteRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.HandleDeleteRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Recipe deleted successfully" {
		t.Errorf("Expected message 'Recipe deleted successfully', got %s", response["message"])
	}
}

func TestAPIHandler_InvalidMethod(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name     string
		method   string
		endpoint string
	}{
		{"PATCH recipes", http.MethodPatch, "/api/recipes"},
		{"HEAD recipes", http.MethodHead, "/api/recipes"},
		{"OPTIONS recipe", http.MethodOptions, "/api/recipes/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			if tt.endpoint == "/api/recipes" {
				handler.HandleRecipes(w, req)
			} else {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", "1")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				handler.HandleRecipe(w, req)
			}

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405, got %d", w.Code)
			}
		})
	}
}
