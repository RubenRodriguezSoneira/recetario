package repositories

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"recipe-app/internal/database"
	"recipe-app/internal/models"
)

// newTestDB provisions a throwaway SQLite database on disk (a temp file is used
// instead of :memory: so the connection pool sees a single shared database) and
// applies the production schema.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.ApplySchema(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

func sampleRecipe() *models.Recipe {
	return &models.Recipe{
		UserID:      "user1",
		Title:       "Tortilla de patatas",
		Description: "Spanish omelette",
		PrepTime:    15,
		CookTime:    25,
		Servings:    4,
		Difficulty:  "easy",
		Category:    "Main",
		Cuisine:     "Spanish",
		ImageURL:    "/img/tortilla.jpg",
		Ingredients: []models.Ingredient{
			{Name: "Potatoes", Amount: "500", Unit: "g"},
			{Name: "Eggs", Amount: "6", Unit: "unit"},
		},
		Instructions: []models.Instruction{
			{Text: "Fry the potatoes"},
			{Text: "Beat the eggs and combine"},
		},
		Tags: []string{"vegetarian", "classic"},
	}
}

func TestRecipeRepository_CreateAndGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	in := sampleRecipe()
	if err := repo.CreateRecipe(in); err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}
	if in.ID == "" {
		t.Fatal("expected generated recipe ID")
	}

	got, err := repo.GetRecipe(in.ID)
	if err != nil {
		t.Fatalf("GetRecipe: %v", err)
	}

	if got.Title != in.Title {
		t.Errorf("title = %q, want %q", got.Title, in.Title)
	}
	if got.UserID != "user1" {
		t.Errorf("user_id = %q, want user1", got.UserID)
	}
	if len(got.Ingredients) != 2 {
		t.Fatalf("ingredients = %d, want 2", len(got.Ingredients))
	}
	// Children must come back ordered by position in insertion order.
	if got.Ingredients[0].Name != "Potatoes" || got.Ingredients[0].Position != 1 {
		t.Errorf("first ingredient = %+v, want Potatoes at position 1", got.Ingredients[0])
	}
	if len(got.Instructions) != 2 || got.Instructions[0].Position != 1 {
		t.Errorf("instructions = %+v, want 2 ordered", got.Instructions)
	}
	if len(got.Tags) != 2 {
		t.Errorf("tags = %v, want 2", got.Tags)
	}
}

func TestRecipeRepository_GetRecipe_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	if _, err := repo.GetRecipe("does-not-exist"); err == nil {
		t.Fatal("expected error for missing recipe")
	}
}

func TestRecipeRepository_GetRecipes_SearchAndFilter(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	seed := []*models.Recipe{
		{UserID: "u", Title: "Chicken Curry", Description: "spicy", Difficulty: "hard", CookTime: 35},
		{UserID: "u", Title: "Caesar Salad", Description: "fresh greens", Difficulty: "easy", CookTime: 0},
		{UserID: "u", Title: "Beef Tacos", Description: "mexican chicken-free", Difficulty: "medium", CookTime: 20},
	}
	for _, r := range seed {
		if err := repo.CreateRecipe(r); err != nil {
			t.Fatalf("seed CreateRecipe %q: %v", r.Title, err)
		}
	}

	tests := []struct {
		name       string
		search     string
		difficulty string
		maxCook    int
		wantTitles []string
	}{
		{name: "no filters returns all", wantTitles: []string{"Chicken Curry", "Caesar Salad", "Beef Tacos"}},
		{name: "search matches title", search: "Caesar", wantTitles: []string{"Caesar Salad"}},
		{name: "search matches description", search: "chicken", wantTitles: []string{"Chicken Curry", "Beef Tacos"}},
		{name: "filter by difficulty", difficulty: "easy", wantTitles: []string{"Caesar Salad"}},
		{name: "filter by max cook time", maxCook: 20, wantTitles: []string{"Caesar Salad", "Beef Tacos"}},
		{name: "combined search and difficulty", search: "chicken", difficulty: "hard", wantTitles: []string{"Chicken Curry"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetRecipes(0, 0, tc.search, tc.difficulty, tc.maxCook)
			if err != nil {
				t.Fatalf("GetRecipes: %v", err)
			}
			gotTitles := map[string]bool{}
			for _, r := range got {
				gotTitles[r.Title] = true
			}
			if len(got) != len(tc.wantTitles) {
				t.Fatalf("got %d recipes %v, want %d %v", len(got), titlesOf(got), len(tc.wantTitles), tc.wantTitles)
			}
			for _, want := range tc.wantTitles {
				if !gotTitles[want] {
					t.Errorf("missing %q in results %v", want, titlesOf(got))
				}
			}
		})
	}
}

func TestRecipeRepository_GetRecipes_Pagination(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	for _, title := range []string{"A", "B", "C"} {
		if err := repo.CreateRecipe(&models.Recipe{UserID: "u", Title: title, Difficulty: "easy"}); err != nil {
			t.Fatalf("seed %q: %v", title, err)
		}
	}

	limited, err := repo.GetRecipes(2, 0, "", "", 0)
	if err != nil {
		t.Fatalf("GetRecipes limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("limit=2 returned %d", len(limited))
	}

	// offset>0 with no limit must still work (SQLite needs LIMIT before OFFSET).
	offset, err := repo.GetRecipes(0, 1, "", "", 0)
	if err != nil {
		t.Fatalf("GetRecipes offset: %v", err)
	}
	if len(offset) != 2 {
		t.Errorf("offset=1 returned %d, want 2", len(offset))
	}
}

func TestRecipeRepository_AllowsEmptyDifficulty(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	r := &models.Recipe{UserID: "u", Title: "No difficulty"}
	if err := repo.CreateRecipe(r); err != nil {
		t.Fatalf("CreateRecipe with empty difficulty: %v", err)
	}
}

func TestRecipeRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	in := sampleRecipe()
	if err := repo.CreateRecipe(in); err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}

	in.Title = "Updated title"
	in.Difficulty = "medium"
	in.Ingredients = []models.Ingredient{{Name: "Olive oil", Amount: "2", Unit: "tbsp"}}
	in.Instructions = []models.Instruction{{Text: "Only one step now"}}
	in.Tags = []string{"updated"}

	if err := repo.UpdateRecipe(in); err != nil {
		t.Fatalf("UpdateRecipe: %v", err)
	}

	got, err := repo.GetRecipe(in.ID)
	if err != nil {
		t.Fatalf("GetRecipe: %v", err)
	}
	if got.Title != "Updated title" || got.Difficulty != "medium" {
		t.Errorf("update not persisted: %+v", got)
	}
	if len(got.Ingredients) != 1 || got.Ingredients[0].Name != "Olive oil" {
		t.Errorf("ingredients not rewritten: %+v", got.Ingredients)
	}
	if len(got.Instructions) != 1 {
		t.Errorf("instructions = %d, want 1", len(got.Instructions))
	}
	if len(got.Tags) != 1 || got.Tags[0] != "updated" {
		t.Errorf("tags not rewritten: %v", got.Tags)
	}
}

func TestRecipeRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := NewRecipeRepository(db)

	in := sampleRecipe()
	if err := repo.CreateRecipe(in); err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}

	if err := repo.DeleteRecipe(in.ID); err != nil {
		t.Fatalf("DeleteRecipe: %v", err)
	}

	if _, err := repo.GetRecipe(in.ID); err == nil {
		t.Fatal("expected recipe to be gone")
	}

	// Child rows must be gone too.
	ingredients, err := repo.GetRecipeIngredients(in.ID)
	if err != nil {
		t.Fatalf("GetRecipeIngredients: %v", err)
	}
	if len(ingredients) != 0 {
		t.Errorf("orphan ingredients remain: %d", len(ingredients))
	}
}

func titlesOf(recipes []*models.Recipe) []string {
	out := make([]string, len(recipes))
	for i, r := range recipes {
		out[i] = r.Title
	}
	return out
}
