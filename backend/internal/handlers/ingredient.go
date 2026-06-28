package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// IngredientStore describes the recipe/ingredient data-access methods the
// handler depends on. *repositories.RecipeRepository satisfies it; tests inject
// a fake.
type IngredientStore interface {
	GetRecipe(id string) (*models.Recipe, error)
	GetRecipeIngredients(recipeID string) ([]models.Ingredient, error)
	GetIngredient(id string) (*models.Ingredient, error)
	AddIngredient(recipeID string, ingredient *models.Ingredient) error
	UpdateIngredient(ingredient *models.Ingredient) error
	DeleteIngredient(id string) error
	ReorderIngredients(recipeID string, orderedIDs []string) error
}

type IngredientHandler struct {
	store IngredientStore
}

func NewIngredientHandler(store IngredientStore) *IngredientHandler {
	return &IngredientHandler{store: store}
}

type ingredientRequest struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
	Notes  string `json:"notes"`
}

type reorderRequest struct {
	IngredientIDs []string `json:"ingredient_ids"`
}

// ownedRecipe loads a recipe and verifies the requester owns it. A missing or
// foreign recipe both yield 404 so callers cannot probe other users' recipes.
func (h *IngredientHandler) ownedRecipe(w http.ResponseWriter, r *http.Request, recipeID string) (*models.Recipe, bool) {
	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	recipe, err := h.store.GetRecipe(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return nil, false
	}
	if recipe.UserID != userID {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return nil, false
	}
	return recipe, true
}

// HandleList returns a recipe's ingredients ordered by position. Reading is
// public; the recipe must exist.
func (h *IngredientHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	if _, err := h.store.GetRecipe(recipeID); err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	ingredients, err := h.store.GetRecipeIngredients(recipeID)
	if err != nil {
		log.Error("Failed to list ingredients", "error", err)
		http.Error(w, "Failed to list ingredients", http.StatusInternalServerError)
		return
	}
	if ingredients == nil {
		ingredients = []models.Ingredient{}
	}

	writeJSON(w, http.StatusOK, ingredients)
}

// HandleCreate appends an ingredient to a recipe owned by the requester.
func (h *IngredientHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	recipe, ok := h.ownedRecipe(w, r, recipeID)
	if !ok {
		return
	}

	var req ingredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient := &models.Ingredient{
		Name:   req.Name,
		Amount: req.Amount,
		Unit:   req.Unit,
		Notes:  req.Notes,
	}
	if err := ingredient.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.AddIngredient(recipe.ID, ingredient); err != nil {
		log.Error("Failed to add ingredient", "error", err)
		http.Error(w, "Failed to add ingredient", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, ingredient)
}

// HandleUpdate edits an ingredient belonging to a recipe owned by the requester.
func (h *IngredientHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")
	ingredientID := chi.URLParam(r, "ingredientID")

	if _, ok := h.ownedRecipe(w, r, recipeID); !ok {
		return
	}

	ingredient, err := h.store.GetIngredient(ingredientID)
	if err != nil || ingredient.RecipeID != recipeID {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	var req ingredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient.Name = req.Name
	ingredient.Amount = req.Amount
	ingredient.Unit = req.Unit
	ingredient.Notes = req.Notes
	if err := ingredient.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateIngredient(ingredient); err != nil {
		log.Error("Failed to update ingredient", "error", err)
		http.Error(w, "Failed to update ingredient", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, ingredient)
}

// HandleDelete removes an ingredient belonging to a recipe owned by the requester.
func (h *IngredientHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")
	ingredientID := chi.URLParam(r, "ingredientID")

	if _, ok := h.ownedRecipe(w, r, recipeID); !ok {
		return
	}

	ingredient, err := h.store.GetIngredient(ingredientID)
	if err != nil || ingredient.RecipeID != recipeID {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	if err := h.store.DeleteIngredient(ingredientID); err != nil {
		log.Error("Failed to delete ingredient", "error", err)
		http.Error(w, "Failed to delete ingredient", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleReorder sets the position of every ingredient in a recipe from the
// provided ordered list of IDs. The list must contain exactly the recipe's
// current ingredient IDs.
func (h *IngredientHandler) HandleReorder(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	recipe, ok := h.ownedRecipe(w, r, recipeID)
	if !ok {
		return
	}

	var req reorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(req.IngredientIDs) == 0 {
		http.Error(w, "ingredient_ids is required", http.StatusBadRequest)
		return
	}

	current, err := h.store.GetRecipeIngredients(recipe.ID)
	if err != nil {
		log.Error("Failed to load ingredients for reorder", "error", err)
		http.Error(w, "Failed to reorder ingredients", http.StatusInternalServerError)
		return
	}
	if !sameIDSet(req.IngredientIDs, current) {
		http.Error(w, "ingredient_ids must list exactly the recipe's ingredients", http.StatusBadRequest)
		return
	}

	if err := h.store.ReorderIngredients(recipe.ID, req.IngredientIDs); err != nil {
		log.Error("Failed to reorder ingredients", "error", err)
		http.Error(w, "Failed to reorder ingredients", http.StatusInternalServerError)
		return
	}

	ingredients, err := h.store.GetRecipeIngredients(recipe.ID)
	if err != nil {
		log.Error("Failed to load reordered ingredients", "error", err)
		http.Error(w, "Failed to reorder ingredients", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ingredients)
}

// sameIDSet reports whether ids contains exactly the IDs of current, with no
// duplicates or extras.
func sameIDSet(ids []string, current []models.Ingredient) bool {
	if len(ids) != len(current) {
		return false
	}
	want := make(map[string]bool, len(current))
	for _, ing := range current {
		want[ing.ID] = true
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !want[id] || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
