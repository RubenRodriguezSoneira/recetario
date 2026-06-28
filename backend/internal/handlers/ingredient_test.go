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

type fakeIngredientStore struct {
	recipes     map[string]*models.Recipe
	ingredients map[string][]models.Ingredient
	byID        map[string]*models.Ingredient
	addErr      error
	updateErr   error
	deleteErr   error
	reorderErr  error
	reordered   []string
}

func newIngredientStore() *fakeIngredientStore {
	ing1 := &models.Ingredient{ID: "ing1", RecipeID: "1", Name: "Flour", Amount: "200", Unit: "g", Position: 1}
	ing2 := &models.Ingredient{ID: "ing2", RecipeID: "1", Name: "Sugar", Amount: "100", Unit: "g", Position: 2}
	foreign := &models.Ingredient{ID: "ing9", RecipeID: "2", Name: "Salt", Position: 1}
	return &fakeIngredientStore{
		recipes: map[string]*models.Recipe{
			"1": {ID: "1", UserID: testOwner, Title: "Cake"},
			"2": {ID: "2", UserID: "other", Title: "Soup"},
		},
		ingredients: map[string][]models.Ingredient{
			"1": {*ing1, *ing2},
			"2": {*foreign},
		},
		byID: map[string]*models.Ingredient{
			"ing1": ing1, "ing2": ing2, "ing9": foreign,
		},
	}
}

func (f *fakeIngredientStore) GetRecipe(id string) (*models.Recipe, error) {
	if r, ok := f.recipes[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("recipe not found")
}

func (f *fakeIngredientStore) GetRecipeIngredients(recipeID string) ([]models.Ingredient, error) {
	return f.ingredients[recipeID], nil
}

func (f *fakeIngredientStore) GetIngredient(id string) (*models.Ingredient, error) {
	if ing, ok := f.byID[id]; ok {
		return ing, nil
	}
	return nil, fmt.Errorf("ingredient not found")
}

func (f *fakeIngredientStore) AddIngredient(recipeID string, ingredient *models.Ingredient) error {
	if f.addErr != nil {
		return f.addErr
	}
	ingredient.ID = "new"
	ingredient.RecipeID = recipeID
	ingredient.Position = len(f.ingredients[recipeID]) + 1
	return nil
}

func (f *fakeIngredientStore) UpdateIngredient(ingredient *models.Ingredient) error {
	return f.updateErr
}

func (f *fakeIngredientStore) DeleteIngredient(id string) error {
	return f.deleteErr
}

func (f *fakeIngredientStore) ReorderIngredients(recipeID string, orderedIDs []string) error {
	if f.reorderErr != nil {
		return f.reorderErr
	}
	f.reordered = orderedIDs
	return nil
}

func TestIngredientHandler_List(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/1/ingredients", nil)
	req = withParams(req, map[string]string{"id": "1"}, "")
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var ingredients []models.Ingredient
	if err := json.Unmarshal(w.Body.Bytes(), &ingredients); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(ingredients) != 2 {
		t.Errorf("Expected 2 ingredients, got %d", len(ingredients))
	}
}

func TestIngredientHandler_List_RecipeNotFound(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/999/ingredients", nil)
	req = withParams(req, map[string]string{"id": "999"}, "")
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestIngredientHandler_Create(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	body := strings.NewReader(`{"name":"Butter","amount":"50","unit":"g"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/ingredients", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}
	var ing models.Ingredient
	if err := json.Unmarshal(w.Body.Bytes(), &ing); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if ing.Name != "Butter" || ing.RecipeID != "1" {
		t.Errorf("Unexpected created ingredient: %+v", ing)
	}
}

func TestIngredientHandler_Create_Unauthorized(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	body := strings.NewReader(`{"name":"Butter"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/ingredients", body)
	req = withParams(req, map[string]string{"id": "1"}, "")
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestIngredientHandler_Create_NonOwner(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	body := strings.NewReader(`{"name":"Butter"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/ingredients", body)
	req = withParams(req, map[string]string{"id": "1"}, "intruder")
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-owner, got %d", w.Code)
	}
}

func TestIngredientHandler_Create_Invalid(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	body := strings.NewReader(`{"amount":"50"}`) // missing name
	req := httptest.NewRequest(http.MethodPost, "/api/recipes/1/ingredients", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestIngredientHandler_Update(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	body := strings.NewReader(`{"name":"Brown Sugar","amount":"120","unit":"g"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1/ingredients/ing2", body)
	req = withParams(req, map[string]string{"id": "1", "ingredientID": "ing2"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var ing models.Ingredient
	if err := json.Unmarshal(w.Body.Bytes(), &ing); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if ing.Name != "Brown Sugar" {
		t.Errorf("Expected updated name, got %q", ing.Name)
	}
}

func TestIngredientHandler_Update_WrongRecipe(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	// ing9 belongs to recipe "2"; updating it via recipe "1" must 404.
	body := strings.NewReader(`{"name":"Hack"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1/ingredients/ing9", body)
	req = withParams(req, map[string]string{"id": "1", "ingredientID": "ing9"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for foreign ingredient, got %d", w.Code)
	}
}

func TestIngredientHandler_Delete(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/1/ingredients/ing1", nil)
	req = withParams(req, map[string]string{"id": "1", "ingredientID": "ing1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestIngredientHandler_Reorder(t *testing.T) {
	store := newIngredientStore()
	handler := NewIngredientHandler(store)

	body := strings.NewReader(`{"ingredient_ids":["ing2","ing1"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1/ingredients/order", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleReorder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	if len(store.reordered) != 2 || store.reordered[0] != "ing2" {
		t.Errorf("Expected reordered [ing2 ing1], got %v", store.reordered)
	}
}

func TestIngredientHandler_Reorder_Mismatch(t *testing.T) {
	handler := NewIngredientHandler(newIngredientStore())

	// Missing ing2 / unknown id -> does not match the recipe's set.
	body := strings.NewReader(`{"ingredient_ids":["ing1","ing999"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/recipes/1/ingredients/order", body)
	req = withParams(req, map[string]string{"id": "1"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleReorder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for mismatched order, got %d", w.Code)
	}
}
