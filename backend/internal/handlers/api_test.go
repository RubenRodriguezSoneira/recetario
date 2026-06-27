package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAPIHandler_GetRecipes(t *testing.T) {
	// For testing, we'll pass nil repository to use mock-like behavior
	handler := NewAPIHandler(nil)

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
				{"id": "1", "title": "Spaghetti Bolognese", "description": "Classic Italian pasta dish with rich meat sauce", "cook_time": 30, "difficulty": "medium"},
				{"id": "2", "title": "Chicken Curry", "description": "Spicy and aromatic Indian curry with tender chicken", "cook_time": 45, "difficulty": "hard"},
				{"id": "3", "title": "Caesar Salad", "description": "Fresh romaine lettuce with creamy Caesar dressing", "cook_time": 15, "difficulty": "easy"},
				{"id": "4", "title": "Beef Tacos", "description": "Mexican-style tacos with seasoned ground beef", "cook_time": 25, "difficulty": "medium"},
				{"id": "5", "title": "Chocolate Cake", "description": "Rich and moist chocolate cake with fudge frosting", "cook_time": 60, "difficulty": "hard"},
				{"id": "6", "title": "Greek Salad", "description": "Mediterranean salad with feta cheese and olives", "cook_time": 10, "difficulty": "easy"},
			},
		},
		{
			name:           "POST to recipes endpoint handled by getRecipes method",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/recipes", nil)
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
	handler := NewAPIHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/recipes", nil)
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

func TestAPIHandler_GetRecipe(t *testing.T) {
	handler := NewAPIHandler(nil)

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

func TestAPIHandler_UpdateRecipe(t *testing.T) {
	handler := NewAPIHandler(nil)

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
	handler := NewAPIHandler(nil)

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
	handler := NewAPIHandler(nil)

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
