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
		appmiddleware.WriteJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "Método no permitido.")
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
		appmiddleware.WriteJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "Método no permitido.")
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
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No se pudieron listar las recetas en este momento.")
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
				logger.FromContext(r.Context()).Error("Failed to render recipe cards template", "error", err)
				appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No se pudieron listar las recetas en este momento.")
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
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request payload", "No pudimos procesar la receta. Revisa los datos e inténtalo de nuevo.")
		return
	}

	if err := recipe.Validate(); err != nil {
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "RECIPE_VALIDATION_FAILED", "Invalid recipe data", err.Error())
		return
	}

	// Owner comes from the authenticated context, never from the request body.
	if userID, ok := appmiddleware.GetUserID(r.Context()); ok {
		recipe.UserID = userID
	}

	if err := h.recipeRepo.CreateRecipe(&recipe); err != nil {
		log.Error("Failed to create recipe", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No se pudo crear la receta en este momento.")
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
		appmiddleware.WriteJSONError(w, http.StatusNotFound, "RECIPE_NOT_FOUND", "Resource not found", "No encontramos la receta solicitada.")
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
				logger.FromContext(r.Context()).Error("Failed to render recipe detail template", "error", err)
				appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No pudimos cargar la receta en este momento.")
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
		appmiddleware.WriteJSONError(w, http.StatusNotFound, "RECIPE_NOT_FOUND", "Resource not found", "No encontramos la receta solicitada.")
		return
	}

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "Debes iniciar sesión para editar esta receta.")
		return
	}
	if existing.UserID != userID {
		appmiddleware.WriteJSONError(w, http.StatusForbidden, "FORBIDDEN", "Access forbidden", "No tienes permisos para editar esta receta.")
		return
	}

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request payload", "No pudimos procesar la receta. Revisa los datos e inténtalo de nuevo.")
		return
	}

	if err := recipe.Validate(); err != nil {
		appmiddleware.WriteJSONError(w, http.StatusBadRequest, "RECIPE_VALIDATION_FAILED", "Invalid recipe data", err.Error())
		return
	}

	// Preserve identity and ownership; they are not client-controlled.
	recipe.ID = recipeID
	recipe.UserID = existing.UserID

	if err := h.recipeRepo.UpdateRecipe(&recipe); err != nil {
		log.Error("Failed to update recipe", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No se pudo actualizar la receta en este momento.")
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
		appmiddleware.WriteJSONError(w, http.StatusNotFound, "RECIPE_NOT_FOUND", "Resource not found", "No encontramos la receta solicitada.")
		return
	}

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		appmiddleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "Debes iniciar sesión para eliminar esta receta.")
		return
	}
	if existing.UserID != userID {
		appmiddleware.WriteJSONError(w, http.StatusForbidden, "FORBIDDEN", "Access forbidden", "No tienes permisos para eliminar esta receta.")
		return
	}

	if err := h.recipeRepo.DeleteRecipe(recipeID); err != nil {
		log.Error("Failed to delete recipe", "error", err)
		appmiddleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "No se pudo eliminar la receta en este momento.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
