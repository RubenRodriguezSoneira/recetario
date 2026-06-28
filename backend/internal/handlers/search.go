package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"recipe-app/internal/logger"
	"recipe-app/internal/models"
)

// SearchStore describes the discovery data-access methods the search handler
// depends on. *repositories.RecipeRepository satisfies it; tests inject a fake.
type SearchStore interface {
	SearchRecipes(filter *models.RecipeFilter, limit, offset int) ([]*models.Recipe, int, error)
	SuggestTitles(query string, limit int) ([]string, error)
	GetPopularTags(limit int) ([]models.TagCount, error)
}

type SearchHandler struct {
	store SearchStore
}

func NewSearchHandler(store SearchStore) *SearchHandler {
	return &SearchHandler{store: store}
}

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 100
)

// HandleSearch runs a filtered recipe search and returns a paginated result.
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	q := r.URL.Query()

	filter := &models.RecipeFilter{
		Query:       q.Get("q"),
		Category:    q.Get("category"),
		Cuisine:     q.Get("cuisine"),
		Difficulty:  q.Get("difficulty"),
		Tags:        parseTags(q["tags"]),
		MinPrepTime: parseIntDefault(q.Get("min_prep_time"), 0),
		MaxPrepTime: parseIntDefault(q.Get("max_prep_time"), 0),
		MinCookTime: parseIntDefault(q.Get("min_cook_time"), 0),
		MaxCookTime: parseIntDefault(q.Get("max_cook_time"), 0),
		MinServings: parseIntDefault(q.Get("min_servings"), 0),
		MaxServings: parseIntDefault(q.Get("max_servings"), 0),
		SortBy:      q.Get("sort_by"),
		SortOrder:   q.Get("sort_order"),
	}
	if err := filter.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	limit := parseIntDefault(q.Get("limit"), defaultSearchLimit)
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}
	offset := parseIntDefault(q.Get("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	recipes, total, err := h.store.SearchRecipes(filter, limit, offset)
	if err != nil {
		log.Error("Failed to search recipes", "error", err)
		http.Error(w, "Failed to search recipes", http.StatusInternalServerError)
		return
	}

	result := models.SearchResult{
		Recipes: derefRecipes(recipes),
		Total:   total,
		Page:    offset/limit + 1,
		PerPage: limit,
	}
	writeJSON(w, http.StatusOK, result)
}

// HandleSuggestions returns title autocompletions for a prefix query.
func (h *SearchHandler) HandleSuggestions(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"suggestions": []string{}})
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 10)

	titles, err := h.store.SuggestTitles(query, limit)
	if err != nil {
		log.Error("Failed to get suggestions", "error", err)
		http.Error(w, "Failed to get suggestions", http.StatusInternalServerError)
		return
	}
	if titles == nil {
		titles = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"suggestions": titles})
}

// HandlePopularTags returns the most-used tags across all recipes.
func (h *SearchHandler) HandlePopularTags(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	limit := parseIntDefault(r.URL.Query().Get("limit"), 10)

	tags, err := h.store.GetPopularTags(limit)
	if err != nil {
		log.Error("Failed to get popular tags", "error", err)
		http.Error(w, "Failed to get popular tags", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []models.TagCount{}
	}

	writeJSON(w, http.StatusOK, tags)
}

// parseTags flattens repeated and comma-separated tag query values into a
// trimmed, non-empty slice.
func parseTags(values []string) []string {
	var tags []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if t := strings.TrimSpace(part); t != "" {
				tags = append(tags, t)
			}
		}
	}
	return tags
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func derefRecipes(recipes []*models.Recipe) []models.Recipe {
	out := make([]models.Recipe, 0, len(recipes))
	for _, recipe := range recipes {
		if recipe != nil {
			out = append(out, *recipe)
		}
	}
	return out
}
