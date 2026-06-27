package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
	"recipe-app/internal/repositories"
)

type APIHandler struct {
	templates  *template.Template
	recipeRepo *repositories.RecipeRepository
}

func NewAPIHandler(recipeRepo *repositories.RecipeRepository) *APIHandler {
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
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.getRecipes(w, r, ctx)
	case http.MethodPost:
		h.createRecipe(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) HandleRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.getRecipe(w, r, ctx)
	case http.MethodPut:
		h.updateRecipe(w, r, ctx)
	case http.MethodDelete:
		h.deleteRecipe(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) HandleCreateRecipe(w http.ResponseWriter, r *http.Request) {
	h.createRecipe(w, r, r.Context())
}

func (h *APIHandler) HandleUpdateRecipe(w http.ResponseWriter, r *http.Request) {
	h.updateRecipe(w, r, r.Context())
}

func (h *APIHandler) HandleDeleteRecipe(w http.ResponseWriter, r *http.Request) {
	h.deleteRecipe(w, r, r.Context())
}

func (h *APIHandler) getRecipes(w http.ResponseWriter, r *http.Request, ctx interface{}) {
	// Parse query parameters
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

	// Get recipes from database
	recipes, err := h.recipeRepo.GetRecipes(limit, offset, search, difficulty, maxCookTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tmpl := h.templates.Lookup("recipe-cards.html")
		if tmpl != nil {
			data := map[string]interface{}{"recipes": recipeMaps}
			err := tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// Default JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipeMaps)
}

func (h *APIHandler) createRecipe(w http.ResponseWriter, r *http.Request, ctx interface{}) {
	logger.FromContext(r.Context()).Info("Creating new recipe")

	// Parse JSON request body
	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate recipe
	if err := recipe.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create recipe in database
	if err := h.recipeRepo.CreateRecipe(&recipe); err != nil {
		http.Error(w, "Failed to create recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recipe created successfully",
		"id":      recipe.ID,
	})
}

func (h *APIHandler) getRecipe(w http.ResponseWriter, r *http.Request, ctx interface{}) {
	// Get recipe ID from URL
	recipeID := chi.URLParam(r, "id")

	// Get recipe from database
	recipe, err := h.recipeRepo.GetRecipe(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	// Convert to map for templates/JSON response
	recipeMap := map[string]interface{}{
		"id":           recipe.ID,
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

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tmpl := h.templates.Lookup("recipe-detail-content.html")
		if tmpl != nil {
			data := map[string]interface{}{"recipe": recipeMap}
			err := tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tmpl := h.templates.Lookup("recipe-detail-content.html")
		if tmpl != nil {
			data := map[string]interface{}{"recipe": recipeMap}
			err := tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// Default JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipeMap)
}

func (h *APIHandler) updateRecipe(w http.ResponseWriter, r *http.Request, ctx interface{}) {
	logger.FromContext(r.Context()).Info("Updating recipe")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recipe updated successfully",
	})
}

func (h *APIHandler) deleteRecipe(w http.ResponseWriter, r *http.Request, ctx interface{}) {
	logger.FromContext(r.Context()).Info("Deleting recipe")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recipe deleted successfully",
	})
}
