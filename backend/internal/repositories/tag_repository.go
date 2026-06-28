package repositories

import (
	"fmt"

	"recipe-app/internal/models"
)

// GetAllTags returns the distinct tags used across all recipes, optionally
// filtered by a case-insensitive substring, ordered alphabetically.
func (r *RecipeRepository) GetAllTags(search string) ([]string, error) {
	query := "SELECT DISTINCT tag FROM recipe_tags"
	args := []interface{}{}
	if search != "" {
		query += " WHERE tag LIKE ?"
		args = append(args, "%"+search+"%")
	}
	query += " ORDER BY tag"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// GetPopularTags returns the most-used tags ordered by descending usage count.
func (r *RecipeRepository) GetPopularTags(limit int) ([]models.TagCount, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(
		`SELECT tag, COUNT(*) AS count
		 FROM recipe_tags
		 GROUP BY tag
		 ORDER BY count DESC, tag ASC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular tags: %w", err)
	}
	defer rows.Close()

	var tags []models.TagCount
	for rows.Next() {
		var tc models.TagCount
		if err := rows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan popular tag: %w", err)
		}
		tags = append(tags, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate popular tags: %w", err)
	}

	return tags, nil
}

// DeleteRecipeTag removes a single tag from a recipe. It is a no-op if the tag
// is not present.
func (r *RecipeRepository) DeleteRecipeTag(recipeID, tag string) error {
	if _, err := r.db.Exec("DELETE FROM recipe_tags WHERE recipe_id = ? AND tag = ?", recipeID, tag); err != nil {
		return fmt.Errorf("failed to delete recipe tag: %w", err)
	}
	return nil
}

// AddRecipeTag attaches a tag to a recipe idempotently: re-adding an existing
// tag is a no-op rather than a primary-key violation.
func (r *RecipeRepository) AddRecipeTag(recipeID, tag string) error {
	if _, err := r.db.Exec("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)", recipeID, tag); err != nil {
		return fmt.Errorf("failed to add recipe tag: %w", err)
	}
	return nil
}
