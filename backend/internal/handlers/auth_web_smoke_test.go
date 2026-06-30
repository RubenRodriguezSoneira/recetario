package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/models"
)

type smokeRecipeStore struct {
	byID    map[string]*models.Recipe
	counter int
}

func newSmokeRecipeStore() *smokeRecipeStore {
	return &smokeRecipeStore{
		byID: map[string]*models.Recipe{
			"victim-recipe": {
				ID:         "victim-recipe",
				UserID:     "victim-id",
				Title:      "Victim Recipe",
				Difficulty: "easy",
			},
		},
		counter: 1,
	}
}

func (s *smokeRecipeStore) GetRecipes(limit, offset int, search, difficulty string, maxCookTime int) ([]*models.Recipe, error) {
	recipes := make([]*models.Recipe, 0, len(s.byID))
	for _, r := range s.byID {
		recipes = append(recipes, r)
	}
	return recipes, nil
}

func (s *smokeRecipeStore) GetRecipe(id string) (*models.Recipe, error) {
	r, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("recipe not found")
	}
	return r, nil
}

func (s *smokeRecipeStore) CreateRecipe(recipe *models.Recipe) error {
	id := fmt.Sprintf("r%d", s.counter)
	s.counter++
	recipe.ID = id
	cloned := *recipe
	s.byID[id] = &cloned
	return nil
}

func (s *smokeRecipeStore) UpdateRecipe(recipe *models.Recipe) error {
	if _, ok := s.byID[recipe.ID]; !ok {
		return fmt.Errorf("recipe not found")
	}
	cloned := *recipe
	s.byID[recipe.ID] = &cloned
	return nil
}

func (s *smokeRecipeStore) DeleteRecipe(id string) error {
	delete(s.byID, id)
	return nil
}

func TestAuthWebCookieSmokeFlow(t *testing.T) {
	authService := appmiddleware.NewAuthService("smoke-secret")
	userStore := newFakeUserStore()
	recipeStore := newSmokeRecipeStore()

	authHandler := NewAuthHandler(authService, userStore)
	apiHandler := NewAPIHandler(recipeStore)

	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.HandleRegister)
	r.Post("/api/auth/logout", authHandler.HandleLogout)
	r.Route("/api/recipes", func(r chi.Router) {
		r.With(authService.AuthMiddleware).Post("/", apiHandler.HandleCreateRecipe)
		r.Route("/{id}", func(r chi.Router) {
			r.With(authService.AuthMiddleware).Put("/", apiHandler.HandleUpdateRecipe)
			r.With(authService.AuthMiddleware).Delete("/", apiHandler.HandleDeleteRecipe)
		})
	})

	registerBody := `{"email":"cook@example.com","username":"cook","password":"password123","first_name":"Cook"}`
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(registerBody))
	registerRes := httptest.NewRecorder()
	r.ServeHTTP(registerRes, registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerRes.Code, http.StatusCreated)
	}

	var authCookie *http.Cookie
	for _, c := range registerRes.Result().Cookies() {
		if c.Name == appmiddleware.AuthCookieName {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("expected auth cookie after register")
	}

	createBody := `{
		"title":"Smoke Recipe",
		"description":"Created in smoke flow",
		"prep_time":10,
		"cook_time":20,
		"servings":2,
		"difficulty":"easy",
		"ingredients":[{"name":"Rice","amount":"1","unit":"cup"}],
		"instructions":[{"text":"Cook rice","position":1}],
		"tags":["smoke"]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader(createBody))
	createReq.AddCookie(authCookie)
	createRes := httptest.NewRecorder()
	r.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d (body=%q)", createRes.Code, http.StatusCreated, createRes.Body.String())
	}

	var created models.Recipe
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created recipe id")
	}

	updateBody := `{
		"title":"Smoke Recipe Updated",
		"description":"Updated in smoke flow",
		"prep_time":12,
		"cook_time":22,
		"servings":3,
		"difficulty":"medium",
		"ingredients":[{"name":"Rice","amount":"2","unit":"cups"}],
		"instructions":[{"text":"Cook better","position":1}],
		"tags":["smoke","updated"]
	}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/recipes/"+created.ID, strings.NewReader(updateBody))
	updateReq.AddCookie(authCookie)
	updateRes := httptest.NewRecorder()
	r.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d (body=%q)", updateRes.Code, http.StatusOK, updateRes.Body.String())
	}

	foreignDeleteReq := httptest.NewRequest(http.MethodDelete, "/api/recipes/victim-recipe", nil)
	foreignDeleteReq.AddCookie(authCookie)
	foreignDeleteRes := httptest.NewRecorder()
	r.ServeHTTP(foreignDeleteRes, foreignDeleteReq)
	if foreignDeleteRes.Code != http.StatusForbidden {
		t.Fatalf("delete non-owner status = %d, want %d", foreignDeleteRes.Code, http.StatusForbidden)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/recipes/"+created.ID, nil)
	deleteReq.AddCookie(authCookie)
	deleteRes := httptest.NewRecorder()
	r.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("delete owner status = %d, want %d", deleteRes.Code, http.StatusNoContent)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(authCookie)
	logoutRes := httptest.NewRecorder()
	r.ServeHTTP(logoutRes, logoutReq)
	if logoutRes.Code != http.StatusSeeOther {
		t.Fatalf("logout status = %d, want %d", logoutRes.Code, http.StatusSeeOther)
	}

	clearedCookieFound := false
	for _, c := range logoutRes.Result().Cookies() {
		if c.Name == appmiddleware.AuthCookieName {
			clearedCookieFound = true
			if c.MaxAge >= 0 {
				t.Fatalf("expected cleared cookie MaxAge < 0, got %d", c.MaxAge)
			}
		}
	}
	if !clearedCookieFound {
		t.Fatal("expected cleared auth cookie on logout")
	}
}
