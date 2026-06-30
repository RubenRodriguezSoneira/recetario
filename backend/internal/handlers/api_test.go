package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

// fakeRecipeStore is an in-memory RecipeStore used to exercise the API handlers
// without a database. Tests inject canned data and control error paths.
type fakeRecipeStore struct {
	recipes   []*models.Recipe
	byID      map[string]*models.Recipe
	createErr error
	updateErr error
	deleteErr error
	deleted   []string
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

func (f *fakeRecipeStore) UpdateRecipe(recipe *models.Recipe) error {
	return f.updateErr
}

func (f *fakeRecipeStore) DeleteRecipe(id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, id)
	return nil
}

const testOwner = "owner1"

func newTestStore() *fakeRecipeStore {
	recipes := []*models.Recipe{
		{
			ID:          "1",
			UserID:      testOwner,
			Title:       "Spaghetti Bolognese",
			Description: "Classic Italian pasta dish with rich meat sauce",
			CookTime:    30,
			Difficulty:  "medium",
			Ingredients: []models.Ingredient{{Amount: "2", Unit: "tazas", Name: "Harina"}},
			Instructions: []models.Instruction{
				{Text: "Mezclar ingredientes"},
			},
		},
		{ID: "2", UserID: testOwner, Title: "Chicken Curry", Description: "Spicy and aromatic Indian curry with tender chicken", CookTime: 45, Difficulty: "hard"},
		{ID: "3", UserID: testOwner, Title: "Caesar Salad", Description: "Fresh romaine lettuce with creamy Caesar dressing", CookTime: 15, Difficulty: "easy"},
		{ID: "4", UserID: testOwner, Title: "Beef Tacos", Description: "Mexican-style tacos with seasoned ground beef", CookTime: 25, Difficulty: "medium"},
		{ID: "5", UserID: testOwner, Title: "Chocolate Cake", Description: "Rich and moist chocolate cake with fudge frosting", CookTime: 60, Difficulty: "hard"},
		{ID: "6", UserID: testOwner, Title: "Greek Salad", Description: "Mediterranean salad with feta cheese and olives", CookTime: 10, Difficulty: "easy"},
	}
	byID := make(map[string]*models.Recipe, len(recipes))
	for _, r := range recipes {
		byID[r.ID] = r
	}
	return &fakeRecipeStore{recipes: recipes, byID: byID}
}

func mustLoadDetailTemplate(t *testing.T) *template.Template {
	t.Helper()

	templates, err := template.ParseFiles(
		filepath.Join("..", "..", "web", "templates", "recipe-cards.html"),
		filepath.Join("..", "..", "web", "templates", "recipe-detail-content.html"),
	)
	if err != nil {
		t.Fatalf("failed to parse templates for HTMX test: %v", err)
	}

	return templates
}

// withRecipeCtx attaches a chi route param ("id") and, when userID is non-empty,
// an authenticated user id to the request context.
func withRecipeCtx(req *http.Request, id, userID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	if userID != "" {
		ctx = context.WithValue(ctx, appmiddleware.UserIDKey, userID)
	}
	return req.WithContext(ctx)
}

func TestAPIHandler_GetRecipes(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name           string
		method         string
		userID         string
		expectedStatus int
		expectedBody   []map[string]interface{}
	}{
		{
			name:           "GET recipes returns 200",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{"id": "1", "title": "Spaghetti Bolognese", "is_owner": false},
				{"id": "2", "title": "Chicken Curry", "is_owner": false},
				{"id": "3", "title": "Caesar Salad", "is_owner": false},
				{"id": "4", "title": "Beef Tacos", "is_owner": false},
				{"id": "5", "title": "Chocolate Cake", "is_owner": false},
				{"id": "6", "title": "Greek Salad", "is_owner": false},
			},
		},
		{
			name:           "GET recipes marks owner when authenticated",
			method:         http.MethodGet,
			userID:         testOwner,
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{"id": "1", "title": "Spaghetti Bolognese", "is_owner": true},
				{"id": "2", "title": "Chicken Curry", "is_owner": true},
				{"id": "3", "title": "Caesar Salad", "is_owner": true},
				{"id": "4", "title": "Beef Tacos", "is_owner": true},
				{"id": "5", "title": "Chocolate Cake", "is_owner": true},
				{"id": "6", "title": "Greek Salad", "is_owner": true},
			},
		},
		{
			name:           "POST to recipes endpoint creates and returns 201",
			method:         http.MethodPost,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *strings.Reader
			if tt.method == http.MethodPost {
				body = strings.NewReader(`{"title":"New Recipe","description":"Fresh and quick","cook_time":10,"difficulty":"easy"}`)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, "/api/recipes", body)
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), appmiddleware.UserIDKey, tt.userID))
			}
			w := httptest.NewRecorder()

			handler.HandleRecipes(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == http.MethodGet {
				var response []map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
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
					if response[i]["is_owner"] != expected["is_owner"] {
						t.Errorf("Expected is_owner %v, got %v", expected["is_owner"], response[i]["is_owner"])
					}
				}
			}
		})
	}
}

func TestAPIHandler_GetRecipes_HTMX_OwnerShowsEditAndDeleteInCards(t *testing.T) {
	handler := NewAPIHandler(newTestStore())
	handler.templates = mustLoadDetailTemplate(t)

	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	req = req.WithContext(context.WithValue(req.Context(), appmiddleware.UserIDKey, testOwner))
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	handler.HandleRecipes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("expected text/html content type, got %q", got)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Ver receta") {
		t.Fatalf("expected cards body to include recipe link, got: %s", body)
	}
	if !strings.Contains(body, "/recipes/1/edit") || !strings.Contains(body, "deleteRecipe('1', this)") {
		t.Fatalf("expected owner edit/delete controls in cards, got: %s", body)
	}
}

func TestAPIHandler_GetRecipes_HTMX_AuthenticatedNonOwnerHidesEditAndDeleteInCards(t *testing.T) {
	handler := NewAPIHandler(newTestStore())
	handler.templates = mustLoadDetailTemplate(t)

	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	req = req.WithContext(context.WithValue(req.Context(), appmiddleware.UserIDKey, "intruder"))
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	handler.HandleRecipes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Ver receta") {
		t.Fatalf("expected cards body to include recipe link, got: %s", body)
	}
	if strings.Contains(body, "/edit") || strings.Contains(body, "deleteRecipe(") {
		t.Fatalf("expected owner edit/delete controls hidden in cards for non-owner, got: %s", body)
	}
}

func TestAPIHandler_CreateRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := strings.NewReader(`{"title":"New Recipe","description":"Fresh and quick","cook_time":10,"difficulty":"easy"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes", body)
	req = req.WithContext(context.WithValue(req.Context(), appmiddleware.UserIDKey, testOwner))
	w := httptest.NewRecorder()

	handler.HandleCreateRecipe(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "3" {
		t.Errorf("Expected id '3', got %v", response["id"])
	}
	if response["user_id"] != testOwner {
		t.Errorf("Expected user_id '%s', got %v", testOwner, response["user_id"])
	}
}

func TestAPIHandler_CreateRecipe_InvalidBody(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.HandleCreateRecipe(w, req)
	assertErrorContract(t, w, http.StatusBadRequest, "INVALID_JSON", "No pudimos procesar la receta. Revisa los datos e inténtalo de nuevo.")
}

func TestAPIHandler_CreateRecipe_InvalidRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name     string
		body     string
		wantDesc string
	}{
		{
			name:     "missing description",
			body:     `{"title":"New Recipe","cook_time":10,"difficulty":"easy"}`,
			wantDesc: "description is required",
		},
		{
			name:     "missing description and title returns description first",
			body:     `{"cook_time":10,"difficulty":"easy"}`,
			wantDesc: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.HandleCreateRecipe(w, req)
			assertErrorContract(t, w, http.StatusBadRequest, "RECIPE_VALIDATION_FAILED", tt.wantDesc)
		})
	}
}

func TestAPIHandler_GetRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", "")
	w := httptest.NewRecorder()

	handler.HandleRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "1" {
		t.Errorf("Expected recipe ID '1', got %v", response["id"])
	}
	if response["title"] != "Spaghetti Bolognese" {
		t.Errorf("Expected recipe title 'Spaghetti Bolognese', got %v", response["title"])
	}
}

func TestAPIHandler_GetRecipe_NotFound(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/999", nil)
	req = withRecipeCtx(req, "999", "")
	w := httptest.NewRecorder()

	handler.HandleRecipe(w, req)

	assertErrorContract(t, w, http.StatusNotFound, "RECIPE_NOT_FOUND", "No encontramos la receta solicitada.")
}

func TestAPIHandler_GetRecipe_HTMX_OwnerShowsEditAndDelete(t *testing.T) {
	handler := NewAPIHandler(newTestStore())
	handler.templates = mustLoadDetailTemplate(t)

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", testOwner)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	handler.HandleRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("expected text/html content type, got %q", got)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Editar") || !strings.Contains(body, "Eliminar receta") {
		t.Fatalf("expected owner actions in HTMX body, got: %s", body)
	}
	if strings.Contains(body, `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("unexpected error JSON leaked into HTMX body: %s", body)
	}
	if !strings.Contains(body, "Harina") || !strings.Contains(body, "Mezclar ingredientes") {
		t.Fatalf("expected ingredient and instruction content in HTMX body, got: %s", body)
	}
}

func TestAPIHandler_GetRecipe_HTMX_NonOwnerHidesEditAndDelete(t *testing.T) {
	handler := NewAPIHandler(newTestStore())
	handler.templates = mustLoadDetailTemplate(t)

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", "intruder")
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	handler.HandleRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "Editar") || strings.Contains(body, "Eliminar receta") {
		t.Fatalf("expected owner actions hidden for non-owner, got: %s", body)
	}
	if strings.Contains(body, `"code":"INTERNAL_ERROR"`) || strings.Contains(body, `"error":`) {
		t.Fatalf("unexpected JSON error leaked into HTMX body: %s", body)
	}
}

func TestAPIHandler_HandleRecipe_NonOwnerMutationForbidden(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name       string
		method     string
		body       string
		wantDetail string
	}{
		{
			name:       "PUT returns 403 for non-owner",
			method:     http.MethodPut,
			body:       `{"title":"Hijack","cook_time":20,"difficulty":"medium"}`,
			wantDetail: "No tienes permisos para editar esta receta.",
		},
		{
			name:       "DELETE returns 403 for non-owner",
			method:     http.MethodDelete,
			wantDetail: "No tienes permisos para eliminar esta receta.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/recipes/1", strings.NewReader(tt.body))
			req = withRecipeCtx(req, "1", "intruder")
			w := httptest.NewRecorder()

			handler.HandleRecipe(w, req)

			assertErrorContract(t, w, http.StatusForbidden, "FORBIDDEN", tt.wantDetail)
		})
	}
}

func TestAPIHandler_UpdateRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := `{"title":"Updated Title","description":"Updated description","cook_time":20,"difficulty":"medium"}`
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader(body))
	req = withRecipeCtx(req, "1", testOwner)
	w := httptest.NewRecorder()

	handler.HandleUpdateRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if response["title"] != "Updated Title" {
		t.Errorf("Expected updated title, got %v", response["title"])
	}
	if response["id"] != "1" {
		t.Errorf("Expected id preserved as '1', got %v", response["id"])
	}
}

func TestAPIHandler_UpdateRecipe_Forbidden(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := `{"title":"Hijack","cook_time":20,"difficulty":"medium"}`
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader(body))
	req = withRecipeCtx(req, "1", "intruder")
	w := httptest.NewRecorder()

	handler.HandleUpdateRecipe(w, req)
	assertErrorContract(t, w, http.StatusForbidden, "FORBIDDEN", "No tienes permisos para editar esta receta.")
}

func TestAPIHandler_UpdateRecipe_Unauthorized(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := `{"title":"Anon","cook_time":20,"difficulty":"medium"}`
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader(body))
	req = withRecipeCtx(req, "1", "") // no authenticated user
	w := httptest.NewRecorder()

	handler.HandleUpdateRecipe(w, req)
	assertErrorContract(t, w, http.StatusUnauthorized, "UNAUTHORIZED", "Debes iniciar sesión para editar esta receta.")
}

func TestAPIHandler_UpdateRecipe_InvalidBody(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader("not json"))
	req = withRecipeCtx(req, "1", testOwner)
	w := httptest.NewRecorder()

	handler.HandleUpdateRecipe(w, req)

	assertErrorContract(t, w, http.StatusBadRequest, "INVALID_JSON", "No pudimos procesar la receta. Revisa los datos e inténtalo de nuevo.")
}

func TestAPIHandler_UpdateRecipe_InvalidRecipe(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	tests := []struct {
		name     string
		body     string
		wantDesc string
	}{
		{
			name:     "missing description",
			body:     `{"title":"Updated Title","cook_time":20,"difficulty":"medium"}`,
			wantDesc: "description is required",
		},
		{
			name:     "missing description and title returns description first",
			body:     `{"cook_time":20,"difficulty":"medium"}`,
			wantDesc: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader(tt.body))
			req = withRecipeCtx(req, "1", testOwner)
			w := httptest.NewRecorder()

			handler.HandleUpdateRecipe(w, req)

			assertErrorContract(t, w, http.StatusBadRequest, "RECIPE_VALIDATION_FAILED", tt.wantDesc)
		})
	}
}

func TestAPIHandler_UpdateRecipe_NotFound(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	body := `{"title":"Ghost","cook_time":20,"difficulty":"medium"}`
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/999", strings.NewReader(body))
	req = withRecipeCtx(req, "999", testOwner)
	w := httptest.NewRecorder()

	handler.HandleUpdateRecipe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for missing recipe, got %d", w.Code)
	}
}

func TestAPIHandler_DeleteRecipe(t *testing.T) {
	store := newTestStore()
	handler := NewAPIHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", testOwner)
	w := httptest.NewRecorder()

	handler.HandleDeleteRecipe(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "1" {
		t.Errorf("Expected recipe '1' to be deleted, got %v", store.deleted)
	}
}

func TestAPIHandler_DeleteRecipe_Forbidden(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", "intruder")
	w := httptest.NewRecorder()

	handler.HandleDeleteRecipe(w, req)
	assertErrorContract(t, w, http.StatusForbidden, "FORBIDDEN", "No tienes permisos para eliminar esta receta.")
}

func TestAPIHandler_DeleteRecipe_Unauthorized(t *testing.T) {
	handler := NewAPIHandler(newTestStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1", nil)
	req = withRecipeCtx(req, "1", "")
	w := httptest.NewRecorder()

	handler.HandleDeleteRecipe(w, req)

	assertErrorContract(t, w, http.StatusUnauthorized, "UNAUTHORIZED", "Debes iniciar sesión para eliminar esta receta.")
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
				req = withRecipeCtx(req, "1", "")
				handler.HandleRecipe(w, req)
			}

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405, got %d", w.Code)
			}
		})
	}
}
