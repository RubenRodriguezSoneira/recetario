package repositories

import (
	"database/sql"
	"fmt"

	"recipe-app/internal/models"
)

// GetIngredient retrieves a single ingredient by ID, translating a missing row
// into a domain "not found" error.
func (r *RecipeRepository) GetIngredient(id string) (*models.Ingredient, error) {
	ingredient := &models.Ingredient{}
	err := r.db.QueryRow(
		"SELECT id, recipe_id, name, amount, unit, notes, position FROM ingredients WHERE id = ?",
		id,
	).Scan(
		&ingredient.ID,
		&ingredient.RecipeID,
		&ingredient.Name,
		&ingredient.Amount,
		&ingredient.Unit,
		&ingredient.Notes,
		&ingredient.Position,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ingredient not found")
		}
		return nil, fmt.Errorf("failed to get ingredient: %w", err)
	}
	return ingredient, nil
}

// AddIngredient appends an ingredient to a recipe, assigning it the next
// position. The position lookup and insert run in one transaction so concurrent
// appends cannot collide on the same position.
func (r *RecipeRepository) AddIngredient(recipeID string, ingredient *models.Ingredient) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var nextPosition int
	if err := tx.QueryRow(
		"SELECT COALESCE(MAX(position), 0) + 1 FROM ingredients WHERE recipe_id = ?",
		recipeID,
	).Scan(&nextPosition); err != nil {
		return fmt.Errorf("failed to compute ingredient position: %w", err)
	}

	ingredient.ID = ""
	ingredient.RecipeID = recipeID
	ingredient.Position = nextPosition
	if err := r.insertIngredient(tx, ingredient); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// UpdateIngredient updates an ingredient's editable fields. Position is managed
// separately via ReorderIngredients.
func (r *RecipeRepository) UpdateIngredient(ingredient *models.Ingredient) error {
	result, err := r.db.Exec(
		"UPDATE ingredients SET name = ?, amount = ?, unit = ?, notes = ? WHERE id = ?",
		ingredient.Name,
		ingredient.Amount,
		ingredient.Unit,
		ingredient.Notes,
		ingredient.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ingredient: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("ingredient not found")
	}
	return nil
}

// DeleteIngredient removes a single ingredient by ID.
func (r *RecipeRepository) DeleteIngredient(id string) error {
	if _, err := r.db.Exec("DELETE FROM ingredients WHERE id = ?", id); err != nil {
		return fmt.Errorf("failed to delete ingredient: %w", err)
	}
	return nil
}

// ReorderIngredients assigns positions 1..N to the given ingredient IDs in the
// order provided. Every ID must belong to recipeID; otherwise the whole
// reorder is rolled back. Callers should pass the recipe's full ingredient set.
func (r *RecipeRepository) ReorderIngredients(recipeID string, orderedIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range orderedIDs {
		result, err := tx.Exec(
			"UPDATE ingredients SET position = ? WHERE id = ? AND recipe_id = ?",
			i+1, id, recipeID,
		)
		if err != nil {
			return fmt.Errorf("failed to reorder ingredient: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to read reorder result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("ingredient %s does not belong to recipe", id)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
