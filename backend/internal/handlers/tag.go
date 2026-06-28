package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// TagStore describes the tag data-access methods the handler depends on.
// *repositories.RecipeRepository satisfies it; tests inject a fake.
type TagStore interface {
	GetRecipe(id string) (*models.Recipe, error)
	GetAllTags(search string) ([]string, error)
	GetRecipeTags(recipeID string) ([]string, error)
	AddRecipeTag(recipeID, tag string) error
	DeleteRecipeTag(recipeID, tag string) error
}

type TagHandler struct {
	store TagStore
}

func NewTagHandler(store TagStore) *TagHandler {
	return &TagHandler{store: store}
}

type addTagsRequest struct {
	Tag  string   `json:"tag"`
	Tags []string `json:"tags"`
}

// ownedRecipe loads a recipe and verifies the requester owns it. A missing or
// foreign recipe both yield 404.
func (h *TagHandler) ownedRecipe(w http.ResponseWriter, r *http.Request, recipeID string) (*models.Recipe, bool) {
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

// HandleListAll returns the distinct tags used across all recipes, optionally
// filtered by a ?search= substring.
func (h *TagHandler) HandleListAll(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	tags, err := h.store.GetAllTags(strings.TrimSpace(r.URL.Query().Get("search")))
	if err != nil {
		log.Error("Failed to list tags", "error", err)
		http.Error(w, "Failed to list tags", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []string{}
	}

	writeJSON(w, http.StatusOK, tags)
}

// HandleListForRecipe returns the tags of a single recipe. Reading is public;
// the recipe must exist.
func (h *TagHandler) HandleListForRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	if _, err := h.store.GetRecipe(recipeID); err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	tags, err := h.store.GetRecipeTags(recipeID)
	if err != nil {
		log.Error("Failed to list recipe tags", "error", err)
		http.Error(w, "Failed to list recipe tags", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []string{}
	}

	writeJSON(w, http.StatusOK, tags)
}

// HandleAdd attaches one or more tags to a recipe owned by the requester and
// returns the recipe's full tag list. Adding an existing tag is idempotent.
func (h *TagHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")

	recipe, ok := h.ownedRecipe(w, r, recipeID)
	if !ok {
		return
	}

	var req addTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tags := normalizeTags(append(req.Tags, req.Tag))
	if len(tags) == 0 {
		http.Error(w, "at least one tag is required", http.StatusBadRequest)
		return
	}

	for _, tag := range tags {
		if err := h.store.AddRecipeTag(recipe.ID, tag); err != nil {
			log.Error("Failed to add recipe tag", "error", err)
			http.Error(w, "Failed to add recipe tag", http.StatusInternalServerError)
			return
		}
	}

	updated, err := h.store.GetRecipeTags(recipe.ID)
	if err != nil {
		log.Error("Failed to load recipe tags", "error", err)
		http.Error(w, "Failed to add recipe tag", http.StatusInternalServerError)
		return
	}
	if updated == nil {
		updated = []string{}
	}

	writeJSON(w, http.StatusCreated, updated)
}

// HandleDelete removes a single tag from a recipe owned by the requester.
func (h *TagHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	recipeID := chi.URLParam(r, "id")
	tag := strings.TrimSpace(chi.URLParam(r, "tag"))

	recipe, ok := h.ownedRecipe(w, r, recipeID)
	if !ok {
		return
	}
	if tag == "" {
		http.Error(w, "tag is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteRecipeTag(recipe.ID, tag); err != nil {
		log.Error("Failed to delete recipe tag", "error", err)
		http.Error(w, "Failed to delete recipe tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// normalizeTags trims, drops empties and de-duplicates a set of tag values
// while preserving first-seen order.
func normalizeTags(values []string) []string {
	var tags []string
	seen := make(map[string]bool)
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		tags = append(tags, t)
	}
	return tags
}
