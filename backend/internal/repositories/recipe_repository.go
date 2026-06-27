package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"recipe-app/internal/models"
)

// sqlExecutor is satisfied by both *sql.DB and *sql.Tx, letting the insert
// helpers run either standalone or inside a transaction.
type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type RecipeRepository struct {
	db *sql.DB
}

func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// CreateRecipe creates a recipe and its child rows atomically.
func (r *RecipeRepository) CreateRecipe(recipe *models.Recipe) error {
	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
	}
	now := time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO recipes (id, user_id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(
		query,
		recipe.ID,
		recipe.UserID,
		recipe.Title,
		recipe.Description,
		recipe.PrepTime,
		recipe.CookTime,
		recipe.Servings,
		recipe.Difficulty,
		recipe.Category,
		recipe.Cuisine,
		recipe.ImageURL,
		now,
		now,
	); err != nil {
		return fmt.Errorf("failed to create recipe: %w", err)
	}

	for i := range recipe.Ingredients {
		recipe.Ingredients[i].RecipeID = recipe.ID
		recipe.Ingredients[i].Position = i + 1
		if err := r.insertIngredient(tx, &recipe.Ingredients[i]); err != nil {
			return err
		}
	}

	for i := range recipe.Instructions {
		recipe.Instructions[i].RecipeID = recipe.ID
		recipe.Instructions[i].Position = i + 1
		if err := r.insertInstruction(tx, &recipe.Instructions[i]); err != nil {
			return err
		}
	}

	for _, tag := range recipe.Tags {
		if err := r.insertRecipeTag(tx, recipe.ID, tag); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	recipe.CreatedAt = now
	recipe.UpdatedAt = now
	return nil
}

// GetRecipe retrieves a recipe by ID with ingredients, instructions, and tags.
func (r *RecipeRepository) GetRecipe(id string) (*models.Recipe, error) {
	query := `
		SELECT id, user_id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at
		FROM recipes
		WHERE id = ?`

	recipe := &models.Recipe{}
	err := r.db.QueryRow(query, id).Scan(
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	ingredients, err := r.GetRecipeIngredients(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe ingredients: %w", err)
	}
	recipe.Ingredients = ingredients

	instructions, err := r.GetRecipeInstructions(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe instructions: %w", err)
	}
	recipe.Instructions = instructions

	tags, err := r.GetRecipeTags(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe tags: %w", err)
	}
	recipe.Tags = tags

	return recipe, nil
}

// GetRecipes retrieves recipes with optional search and filters.
func (r *RecipeRepository) GetRecipes(limit, offset int, search string, difficulty string, maxCookTime int) ([]*models.Recipe, error) {
	query := `
		SELECT id, user_id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at
		FROM recipes
		WHERE 1=1`

	args := []interface{}{}

	if search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	if difficulty != "" {
		query += " AND difficulty = ?"
		args = append(args, difficulty)
	}

	if maxCookTime > 0 {
		query += " AND cook_time <= ?"
		args = append(args, maxCookTime)
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	} else if offset > 0 {
		// SQLite requires a LIMIT before OFFSET; -1 means "no limit".
		query += " LIMIT -1"
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}
	defer rows.Close()

	var recipes []*models.Recipe
	for rows.Next() {
		recipe := &models.Recipe{}
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recipes: %w", err)
	}

	return recipes, nil
}

// UpdateRecipe updates a recipe and rewrites its child rows atomically.
func (r *RecipeRepository) UpdateRecipe(recipe *models.Recipe) error {
	now := time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE recipes
		SET title = ?, description = ?, prep_time = ?, cook_time = ?, servings = ?,
		    difficulty = ?, category = ?, cuisine = ?, image_url = ?, updated_at = ?
		WHERE id = ?`
	if _, err := tx.Exec(
		query,
		recipe.Title,
		recipe.Description,
		recipe.PrepTime,
		recipe.CookTime,
		recipe.Servings,
		recipe.Difficulty,
		recipe.Category,
		recipe.Cuisine,
		recipe.ImageURL,
		now,
		recipe.ID,
	); err != nil {
		return fmt.Errorf("failed to update recipe: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM ingredients WHERE recipe_id = ?", recipe.ID); err != nil {
		return fmt.Errorf("failed to delete existing ingredients: %w", err)
	}
	for i := range recipe.Ingredients {
		recipe.Ingredients[i].RecipeID = recipe.ID
		recipe.Ingredients[i].Position = i + 1
		if err := r.insertIngredient(tx, &recipe.Ingredients[i]); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM instructions WHERE recipe_id = ?", recipe.ID); err != nil {
		return fmt.Errorf("failed to delete existing instructions: %w", err)
	}
	for i := range recipe.Instructions {
		recipe.Instructions[i].RecipeID = recipe.ID
		recipe.Instructions[i].Position = i + 1
		if err := r.insertInstruction(tx, &recipe.Instructions[i]); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", recipe.ID); err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}
	for _, tag := range recipe.Tags {
		if err := r.insertRecipeTag(tx, recipe.ID, tag); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	recipe.UpdatedAt = now
	return nil
}

// DeleteRecipe removes a recipe and its child rows atomically.
func (r *RecipeRepository) DeleteRecipe(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	statements := []string{
		"DELETE FROM ingredients WHERE recipe_id = ?",
		"DELETE FROM instructions WHERE recipe_id = ?",
		"DELETE FROM recipe_tags WHERE recipe_id = ?",
		"DELETE FROM recipes WHERE id = ?",
	}
	for _, stmt := range statements {
		if _, err := tx.Exec(stmt, id); err != nil {
			return fmt.Errorf("failed to delete recipe: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateIngredient inserts a single ingredient.
func (r *RecipeRepository) CreateIngredient(ingredient *models.Ingredient) error {
	return r.insertIngredient(r.db, ingredient)
}

func (r *RecipeRepository) insertIngredient(ex sqlExecutor, ingredient *models.Ingredient) error {
	if ingredient.ID == "" {
		ingredient.ID = uuid.New().String()
	}
	query := `
		INSERT INTO ingredients (id, recipe_id, name, amount, unit, notes, position)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := ex.Exec(
		query,
		ingredient.ID,
		ingredient.RecipeID,
		ingredient.Name,
		ingredient.Amount,
		ingredient.Unit,
		ingredient.Notes,
		ingredient.Position,
	); err != nil {
		return fmt.Errorf("failed to create ingredient: %w", err)
	}
	return nil
}

// GetRecipeIngredients retrieves all ingredients for a recipe.
func (r *RecipeRepository) GetRecipeIngredients(recipeID string) ([]models.Ingredient, error) {
	query := `
		SELECT id, recipe_id, name, amount, unit, notes, position
		FROM ingredients
		WHERE recipe_id = ?
		ORDER BY position`

	rows, err := r.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ingredient models.Ingredient
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.RecipeID,
			&ingredient.Name,
			&ingredient.Amount,
			&ingredient.Unit,
			&ingredient.Notes,
			&ingredient.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate ingredients: %w", err)
	}

	return ingredients, nil
}

// DeleteRecipeIngredients deletes all ingredients for a recipe.
func (r *RecipeRepository) DeleteRecipeIngredients(recipeID string) error {
	_, err := r.db.Exec("DELETE FROM ingredients WHERE recipe_id = ?", recipeID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe ingredients: %w", err)
	}
	return nil
}

// CreateInstruction inserts a single instruction.
func (r *RecipeRepository) CreateInstruction(instruction *models.Instruction) error {
	return r.insertInstruction(r.db, instruction)
}

func (r *RecipeRepository) insertInstruction(ex sqlExecutor, instruction *models.Instruction) error {
	if instruction.ID == "" {
		instruction.ID = uuid.New().String()
	}
	query := `
		INSERT INTO instructions (id, recipe_id, text, position, duration, temperature)
		VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := ex.Exec(
		query,
		instruction.ID,
		instruction.RecipeID,
		instruction.Text,
		instruction.Position,
		instruction.Duration,
		instruction.Temperature,
	); err != nil {
		return fmt.Errorf("failed to create instruction: %w", err)
	}
	return nil
}

// GetRecipeInstructions retrieves all instructions for a recipe.
func (r *RecipeRepository) GetRecipeInstructions(recipeID string) ([]models.Instruction, error) {
	query := `
		SELECT id, recipe_id, text, position, duration, temperature
		FROM instructions
		WHERE recipe_id = ?
		ORDER BY position`

	rows, err := r.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe instructions: %w", err)
	}
	defer rows.Close()

	var instructions []models.Instruction
	for rows.Next() {
		var instruction models.Instruction
		err := rows.Scan(
			&instruction.ID,
			&instruction.RecipeID,
			&instruction.Text,
			&instruction.Position,
			&instruction.Duration,
			&instruction.Temperature,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instruction: %w", err)
		}
		instructions = append(instructions, instruction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate instructions: %w", err)
	}

	return instructions, nil
}

// DeleteRecipeInstructions deletes all instructions for a recipe.
func (r *RecipeRepository) DeleteRecipeInstructions(recipeID string) error {
	_, err := r.db.Exec("DELETE FROM instructions WHERE recipe_id = ?", recipeID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe instructions: %w", err)
	}
	return nil
}

// CreateRecipeTag inserts a single recipe tag.
func (r *RecipeRepository) CreateRecipeTag(recipeID, tag string) error {
	return r.insertRecipeTag(r.db, recipeID, tag)
}

func (r *RecipeRepository) insertRecipeTag(ex sqlExecutor, recipeID, tag string) error {
	if _, err := ex.Exec("INSERT INTO recipe_tags (recipe_id, tag) VALUES (?, ?)", recipeID, tag); err != nil {
		return fmt.Errorf("failed to create recipe tag: %w", err)
	}
	return nil
}

// GetRecipeTags retrieves all tags for a recipe.
func (r *RecipeRepository) GetRecipeTags(recipeID string) ([]string, error) {
	query := `SELECT tag FROM recipe_tags WHERE recipe_id = ? ORDER BY tag`

	rows, err := r.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe tags: %w", err)
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

// DeleteRecipeTags deletes all tags for a recipe.
func (r *RecipeRepository) DeleteRecipeTags(recipeID string) error {
	_, err := r.db.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", recipeID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe tags: %w", err)
	}
	return nil
}
