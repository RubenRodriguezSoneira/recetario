package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// RecipeStore describes the recipe data-access methods the APIHandler depends on.
// Depending on an interface (rather than the concrete *repositories.RecipeRepository)
// keeps the handler thin and testable: production code injects the real repository,
// tests inject a fake. The concrete *repositories.RecipeRepository satisfies it.
type RecipeStore interface {
	GetRecipes(limit, offset int, search, difficulty string, maxCookTime int) ([]*models.Recipe, error)
	GetRecipe(id string) (*models.Recipe, error)
	CreateRecipe(recipe *models.Recipe) error
	UpdateRecipe(recipe *models.Recipe) error
	DeleteRecipe(id string) error
}

type APIHandler struct {
	templates  *template.Template
	recipeRepo RecipeStore
}

func NewAPIHandler(recipeRepo RecipeStore) *APIHandler {
	templates, err := template.ParseFiles("web/templates/recipe-cards.html", "web/templates/recipe-detail-content.html")
	if err != nil {
		// Templates not found, create empty template for tests
		templates = template.New("")
	}
	return &APIHandler{
		templates:  templates,
		recipeRepo: recipeRepo,
	}
}

func (h *APIHandler) HandleRecipes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRecipes(w, r)
	case http.MethodPost:
		h.createRecipe(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) HandleRecipe(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRecipe(w, r)
	case http.MethodPut:
		h.updateRecipe(w, r)
	case http.MethodDelete:
		h.deleteRecipe(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) HandleCreateRecipe(w http.ResponseWriter, r *http.Request) {
	h.createRecipe(w, r)
}

func (h *APIHandler) HandleUpdateRecipe(w http.ResponseWriter, r *http.Request) {
	h.updateRecipe(w, r)
}

func (h *APIHandler) HandleDeleteRecipe(w http.ResponseWriter, r *http.Request) {
	h.deleteRecipe(w, r)
}

func (h *APIHandler) getRecipes(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	difficulty := r.URL.Query().Get("difficulty")
	maxCookTimeStr := r.URL.Query().Get("cook_time")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var maxCookTime int
	if maxCookTimeStr != "" {
		maxCookTime, _ = strconv.Atoi(maxCookTimeStr)
	}

	limit := 20 // default limit
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	offset := 0 // default offset
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	recipes, err := h.recipeRepo.GetRecipes(limit, offset, search, difficulty, maxCookTime)
	if err != nil {
		logger.FromContext(r.Context()).Error("Failed to list recipes", "error", err)
		http.Error(w, "Failed to list recipes", http.StatusInternalServerError)
		return
	}

	// Convert recipes to map format for templates/JSON
	recipeMaps := make([]map[string]interface{}, len(recipes))
	for i, recipe := range recipes {
		recipeMaps[i] = map[string]interface{}{
			"id":          recipe.ID,
			"title":       recipe.Title,
			"description": recipe.Description,
			"cook_time":   recipe.CookTime,
			"difficulty":  recipe.Difficulty,
			"category":    recipe.Category,
			"cuisine":     recipe.Cuisine,
			"image_url":   recipe.ImageURL,
			"created_at":  recipe.CreatedAt,
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tmpl := h.templates.Lookup("recipe-cards.html")
		if tmpl != nil {
			data := map[string]interface{}{"recipes": recipeMaps}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Template execution error", http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipeMaps)
}

func (h *APIHandler) createRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Creating new recipe")

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := recipe.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Owner comes from the authenticated context, never from the request body.
	if userID, ok := appmiddleware.GetUserID(r.Context()); ok {
		recipe.UserID = userID
	}

	if err := h.recipeRepo.CreateRecipe(&recipe); err != nil {
		log.Error("Failed to create recipe", "error", err)
		http.Error(w, "Failed to create recipe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(recipe)
}

func (h *APIHandler) getRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	recipe, err := h.recipeRepo.GetRecipe(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	currentUserID, hasUser := appmiddleware.GetUserID(r.Context())
	isOwner := hasUser && currentUserID == recipe.UserID

	recipeMap := map[string]interface{}{
		"id":           recipe.ID,
		"user_id":      recipe.UserID,
		"is_owner":     isOwner,
		"title":        recipe.Title,
		"description":  recipe.Description,
		"prep_time":    recipe.PrepTime,
		"cook_time":    recipe.CookTime,
		"servings":     recipe.Servings,
		"difficulty":   recipe.Difficulty,
		"category":     recipe.Category,
		"cuisine":      recipe.Cuisine,
		"image_url":    recipe.ImageURL,
		"ingredients":  recipe.Ingredients,
		"instructions": recipe.Instructions,
		"tags":         recipe.Tags,
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tmpl := h.templates.Lookup("recipe-detail-content.html")
		if tmpl != nil {
			data := map[string]interface{}{"recipe": recipeMap}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Template execution error", http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipeMap)
}

func (h *APIHandler) updateRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	existing, err := h.recipeRepo.GetRecipe(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if existing.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := recipe.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Preserve identity and ownership; they are not client-controlled.
	recipe.ID = recipeID
	recipe.UserID = existing.UserID

	if err := h.recipeRepo.UpdateRecipe(&recipe); err != nil {
		log.Error("Failed to update recipe", "error", err)
		http.Error(w, "Failed to update recipe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipe)
}

func (h *APIHandler) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	existing, err := h.recipeRepo.GetRecipe(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if existing.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.recipeRepo.DeleteRecipe(recipeID); err != nil {
		log.Error("Failed to delete recipe", "error", err)
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
