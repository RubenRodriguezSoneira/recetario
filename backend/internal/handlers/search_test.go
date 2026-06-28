package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-app/internal/models"
)

type fakeSearchStore struct {
	recipes    []*models.Recipe
	total      int
	titles     []string
	popular    []models.TagCount
	searchErr  error
	lastFilter *models.RecipeFilter
	lastLimit  int
	lastOffset int
}

func (f *fakeSearchStore) SearchRecipes(filter *models.RecipeFilter, limit, offset int) ([]*models.Recipe, int, error) {
	f.lastFilter = filter
	f.lastLimit = limit
	f.lastOffset = offset
	if f.searchErr != nil {
		return nil, 0, f.searchErr
	}
	return f.recipes, f.total, nil
}

func (f *fakeSearchStore) SuggestTitles(query string, limit int) ([]string, error) {
	return f.titles, nil
}

func (f *fakeSearchStore) GetPopularTags(limit int) ([]models.TagCount, error) {
	return f.popular, nil
}

func newSearchStore() *fakeSearchStore {
	return &fakeSearchStore{
		recipes: []*models.Recipe{
			{ID: "1", UserID: testOwner, Title: "Spaghetti Bolognese"},
			{ID: "2", UserID: testOwner, Title: "Spinach Pie"},
		},
		total:   2,
		titles:  []string{"Spaghetti Bolognese", "Spinach Pie"},
		popular: []models.TagCount{{Tag: "italian", Count: 5}, {Tag: "quick", Count: 3}},
	}
}

func TestSearchHandler_Search(t *testing.T) {
	store := newSearchStore()
	handler := NewSearchHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=spa&tags=italian,quick&limit=10&offset=10", nil)
	w := httptest.NewRecorder()

	handler.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result models.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
	}
	if len(result.Recipes) != 2 {
		t.Errorf("Expected 2 recipes, got %d", len(result.Recipes))
	}
	if result.PerPage != 10 {
		t.Errorf("Expected per_page 10, got %d", result.PerPage)
	}
	if result.Page != 2 {
		t.Errorf("Expected page 2 (offset 10/limit 10 + 1), got %d", result.Page)
	}

	if store.lastFilter.Query != "spa" {
		t.Errorf("Expected query 'spa' forwarded, got %q", store.lastFilter.Query)
	}
	if len(store.lastFilter.Tags) != 2 {
		t.Errorf("Expected 2 tags forwarded, got %v", store.lastFilter.Tags)
	}
	if store.lastOffset != 10 {
		t.Errorf("Expected offset 10 forwarded, got %d", store.lastOffset)
	}
}

func TestSearchHandler_Search_LimitCapped(t *testing.T) {
	store := newSearchStore()
	handler := NewSearchHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/search?limit=5000", nil)
	w := httptest.NewRecorder()

	handler.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	if store.lastLimit != maxSearchLimit {
		t.Errorf("Expected limit capped at %d, got %d", maxSearchLimit, store.lastLimit)
	}
}

func TestSearchHandler_Search_InvalidFilter(t *testing.T) {
	handler := NewSearchHandler(newSearchStore())

	tests := []struct {
		name string
		url  string
	}{
		{"invalid difficulty", "/api/search?difficulty=spicy"},
		{"invalid sort order", "/api/search?sort_order=sideways"},
		{"invalid sort field", "/api/search?sort_by=secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			handler.HandleSearch(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestSearchHandler_Search_StoreError(t *testing.T) {
	store := newSearchStore()
	store.searchErr = fmt.Errorf("boom")
	handler := NewSearchHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	w := httptest.NewRecorder()

	handler.HandleSearch(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestSearchHandler_Suggestions(t *testing.T) {
	handler := NewSearchHandler(newSearchStore())

	req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions?q=sp", nil)
	w := httptest.NewRecorder()

	handler.HandleSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp map[string][]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(resp["suggestions"]) != 2 {
		t.Errorf("Expected 2 suggestions, got %v", resp["suggestions"])
	}
}

func TestSearchHandler_Suggestions_EmptyQuery(t *testing.T) {
	handler := NewSearchHandler(newSearchStore())

	req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions", nil)
	w := httptest.NewRecorder()

	handler.HandleSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp map[string][]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(resp["suggestions"]) != 0 {
		t.Errorf("Expected no suggestions for empty query, got %v", resp["suggestions"])
	}
}

func TestSearchHandler_PopularTags(t *testing.T) {
	handler := NewSearchHandler(newSearchStore())

	req := httptest.NewRequest(http.MethodGet, "/api/search/tags/popular?limit=5", nil)
	w := httptest.NewRecorder()

	handler.HandlePopularTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var tags []models.TagCount
	if err := json.Unmarshal(w.Body.Bytes(), &tags); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(tags) != 2 || tags[0].Tag != "italian" {
		t.Errorf("Expected popular tags, got %v", tags)
	}
}
