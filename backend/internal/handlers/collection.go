package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// CollectionStore describes the collection data-access methods the handler
// depends on. The concrete *repositories.CollectionRepository satisfies it;
// tests inject a fake.
type CollectionStore interface {
	CreateCollection(c *models.RecipeCollection) error
	GetCollection(id string) (*models.RecipeCollection, error)
	GetCollectionsByUser(userID string) ([]*models.RecipeCollection, error)
	UpdateCollection(c *models.RecipeCollection) error
	DeleteCollection(id string) error
	AddRecipe(collectionID, recipeID string) error
	RemoveRecipe(collectionID, recipeID string) error
}

type CollectionHandler struct {
	collections CollectionStore
}

func NewCollectionHandler(collections CollectionStore) *CollectionHandler {
	return &CollectionHandler{collections: collections}
}

type collectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type addRecipeRequest struct {
	RecipeID string `json:"recipe_id"`
}

// loadOwned fetches a collection and verifies the requester owns it. It writes
// the appropriate error response and returns ok=false when the caller should
// stop. A missing collection and a foreign collection both yield 404 so callers
// cannot probe for the existence of other users' collections.
func (h *CollectionHandler) loadOwned(w http.ResponseWriter, r *http.Request, id string) (*models.RecipeCollection, string, bool) {
	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, "", false
	}

	collection, err := h.collections.GetCollection(id)
	if err != nil {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return nil, "", false
	}
	if collection.UserID != userID {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return nil, "", false
	}
	return collection, userID, true
}

func (h *CollectionHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req collectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	collection := &models.RecipeCollection{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}
	if err := collection.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.collections.CreateCollection(collection); err != nil {
		log.Error("Failed to create collection", "error", err)
		http.Error(w, "Failed to create collection", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, collection)
}

func (h *CollectionHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collections, err := h.collections.GetCollectionsByUser(userID)
	if err != nil {
		log.Error("Failed to list collections", "error", err)
		http.Error(w, "Failed to list collections", http.StatusInternalServerError)
		return
	}
	if collections == nil {
		collections = []*models.RecipeCollection{}
	}

	writeJSON(w, http.StatusOK, collections)
}

func (h *CollectionHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collection, err := h.collections.GetCollection(id)
	if err != nil {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	// Owners always see their collections; others only public ones.
	if collection.UserID != userID && !collection.IsPublic {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, collection)
}

func (h *CollectionHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	id := chi.URLParam(r, "id")

	collection, _, ok := h.loadOwned(w, r, id)
	if !ok {
		return
	}

	var req collectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	collection.Name = req.Name
	collection.Description = req.Description
	collection.IsPublic = req.IsPublic
	if err := collection.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.collections.UpdateCollection(collection); err != nil {
		log.Error("Failed to update collection", "error", err)
		http.Error(w, "Failed to update collection", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, collection)
}

func (h *CollectionHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	id := chi.URLParam(r, "id")

	if _, _, ok := h.loadOwned(w, r, id); !ok {
		return
	}

	if err := h.collections.DeleteCollection(id); err != nil {
		log.Error("Failed to delete collection", "error", err)
		http.Error(w, "Failed to delete collection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CollectionHandler) HandleAddRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	id := chi.URLParam(r, "id")

	if _, _, ok := h.loadOwned(w, r, id); !ok {
		return
	}

	var req addRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.RecipeID == "" {
		http.Error(w, "recipe_id is required", http.StatusBadRequest)
		return
	}

	if err := h.collections.AddRecipe(id, req.RecipeID); err != nil {
		log.Error("Failed to add recipe to collection", "error", err)
		http.Error(w, "Failed to add recipe to collection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CollectionHandler) HandleRemoveRecipe(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	id := chi.URLParam(r, "id")
	recipeID := chi.URLParam(r, "recipeID")

	if _, _, ok := h.loadOwned(w, r, id); !ok {
		return
	}

	if err := h.collections.RemoveRecipe(id, recipeID); err != nil {
		log.Error("Failed to remove recipe from collection", "error", err)
		http.Error(w, "Failed to remove recipe from collection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
