package repositories

import (
	"fmt"
	"strings"

	"recipe-app/internal/models"
)

// SearchRecipes returns the recipes matching the filter together with the total
// number of matches (ignoring limit/offset) for pagination. Results are list
// rows only (no ingredients/instructions/tags) to keep the query cheap; callers
// fetch full detail via GetRecipe when needed.
//
// All user-supplied values are bound through placeholders. The only identifiers
// interpolated into SQL are the sort column and direction, which are validated
// against a whitelist first.
func (r *RecipeRepository) SearchRecipes(filter *models.RecipeFilter, limit, offset int) ([]*models.Recipe, int, error) {
	where, args := buildRecipeSearchWhere(filter)

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM recipes WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count recipes: %w", err)
	}

	sortBy := "created_at"
	if filter.SortBy != "" && models.AllowedRecipeSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, user_id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at
		FROM recipes
		WHERE ` + where + `
		ORDER BY ` + sortBy + ` ` + sortOrder + `
		LIMIT ?`
	selectArgs := append(append([]interface{}{}, args...), limit)
	if offset > 0 {
		query += " OFFSET ?"
		selectArgs = append(selectArgs, offset)
	}

	rows, err := r.db.Query(query, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search recipes: %w", err)
	}
	defer rows.Close()

	var recipes []*models.Recipe
	for rows.Next() {
		recipe := &models.Recipe{}
		if err := rows.Scan(
			&recipe.ID,
			&recipe.UserID,
			&recipe.Title,
			&recipe.Description,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.Servings,
			&recipe.Difficulty,
			&recipe.Category,
			&recipe.Cuisine,
			&recipe.ImageURL,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan recipe: %w", err)
		}
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate recipes: %w", err)
	}

	return recipes, total, nil
}

// SuggestTitles returns distinct recipe titles whose prefix matches the query,
// for search autocompletion.
func (r *RecipeRepository) SuggestTitles(query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(
		"SELECT DISTINCT title FROM recipes WHERE title LIKE ? ORDER BY title LIMIT ?",
		query+"%", limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get title suggestions: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, fmt.Errorf("failed to scan title: %w", err)
		}
		titles = append(titles, title)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate titles: %w", err)
	}

	return titles, nil
}

// buildRecipeSearchWhere builds a parameterized WHERE body (without the leading
// "WHERE") and the matching argument slice from the filter. Every dynamic value
// is bound with a "?" placeholder; nothing user-supplied is interpolated.
func buildRecipeSearchWhere(f *models.RecipeFilter) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{}

	if f.Query != "" {
		clauses = append(clauses,
			"(title LIKE ? OR description LIKE ? OR EXISTS (SELECT 1 FROM ingredients i WHERE i.recipe_id = recipes.id AND i.name LIKE ?))")
		like := "%" + f.Query + "%"
		args = append(args, like, like, like)
	}
	if f.Difficulty != "" {
		clauses = append(clauses, "difficulty = ?")
		args = append(args, f.Difficulty)
	}
	if f.Category != "" {
		clauses = append(clauses, "category = ?")
		args = append(args, f.Category)
	}
	if f.Cuisine != "" {
		clauses = append(clauses, "cuisine = ?")
		args = append(args, f.Cuisine)
	}
	if f.MinPrepTime > 0 {
		clauses = append(clauses, "prep_time >= ?")
		args = append(args, f.MinPrepTime)
	}
	if f.MaxPrepTime > 0 {
		clauses = append(clauses, "prep_time <= ?")
		args = append(args, f.MaxPrepTime)
	}
	if f.MinCookTime > 0 {
		clauses = append(clauses, "cook_time >= ?")
		args = append(args, f.MinCookTime)
	}
	if f.MaxCookTime > 0 {
		clauses = append(clauses, "cook_time <= ?")
		args = append(args, f.MaxCookTime)
	}
	if f.MinServings > 0 {
		clauses = append(clauses, "servings >= ?")
		args = append(args, f.MinServings)
	}
	if f.MaxServings > 0 {
		clauses = append(clauses, "servings <= ?")
		args = append(args, f.MaxServings)
	}
	for _, tag := range f.Tags {
		if tag == "" {
			continue
		}
		clauses = append(clauses, "EXISTS (SELECT 1 FROM recipe_tags rt WHERE rt.recipe_id = recipes.id AND rt.tag = ?)")
		args = append(args, tag)
	}

	return strings.Join(clauses, " AND "), args
}
