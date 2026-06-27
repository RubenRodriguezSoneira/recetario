package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"recipe-app/internal/models"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// CreateCollection inserts a new collection, generating its ID when absent.
func (r *CollectionRepository) CreateCollection(c *models.RecipeCollection) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()

	query := `
		INSERT INTO collections (id, user_id, name, description, is_public, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := r.db.Exec(query, c.ID, c.UserID, c.Name, c.Description, c.IsPublic, now, now); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	c.CreatedAt = now
	c.UpdatedAt = now
	return nil
}

// GetCollection retrieves a collection by ID, including its recipe IDs.
func (r *CollectionRepository) GetCollection(id string) (*models.RecipeCollection, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM collections
		WHERE id = ?`

	c := &models.RecipeCollection{}
	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.Description,
		&c.IsPublic,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	recipeIDs, err := r.GetCollectionRecipeIDs(id)
	if err != nil {
		return nil, err
	}
	c.RecipeIDs = recipeIDs

	return c, nil
}

// GetCollectionsByUser lists all collections owned by a user. Recipe IDs are not
// populated here to keep the list query lean; use GetCollection for details.
func (r *CollectionRepository) GetCollectionsByUser(userID string) ([]*models.RecipeCollection, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM collections
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}
	defer rows.Close()

	var collections []*models.RecipeCollection
	for rows.Next() {
		c := &models.RecipeCollection{}
		if err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.Description,
			&c.IsPublic,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan collection: %w", err)
		}
		collections = append(collections, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate collections: %w", err)
	}

	return collections, nil
}

// UpdateCollection updates a collection's mutable fields.
func (r *CollectionRepository) UpdateCollection(c *models.RecipeCollection) error {
	now := time.Now()
	query := `
		UPDATE collections
		SET name = ?, description = ?, is_public = ?, updated_at = ?
		WHERE id = ?`
	if _, err := r.db.Exec(query, c.Name, c.Description, c.IsPublic, now, c.ID); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}
	c.UpdatedAt = now
	return nil
}

// DeleteCollection removes a collection and its recipe associations atomically.
func (r *CollectionRepository) DeleteCollection(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM collection_recipes WHERE collection_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete collection recipes: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM collections WHERE id = ?", id); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// AddRecipe associates a recipe with a collection (idempotent).
func (r *CollectionRepository) AddRecipe(collectionID, recipeID string) error {
	query := `INSERT OR IGNORE INTO collection_recipes (collection_id, recipe_id) VALUES (?, ?)`
	if _, err := r.db.Exec(query, collectionID, recipeID); err != nil {
		return fmt.Errorf("failed to add recipe to collection: %w", err)
	}
	return nil
}

// RemoveRecipe removes a recipe association from a collection.
func (r *CollectionRepository) RemoveRecipe(collectionID, recipeID string) error {
	query := `DELETE FROM collection_recipes WHERE collection_id = ? AND recipe_id = ?`
	if _, err := r.db.Exec(query, collectionID, recipeID); err != nil {
		return fmt.Errorf("failed to remove recipe from collection: %w", err)
	}
	return nil
}

// GetCollectionRecipeIDs returns the recipe IDs associated with a collection.
func (r *CollectionRepository) GetCollectionRecipeIDs(collectionID string) ([]string, error) {
	query := `SELECT recipe_id FROM collection_recipes WHERE collection_id = ? ORDER BY recipe_id`

	rows, err := r.db.Query(query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection recipe ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan recipe id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recipe ids: %w", err)
	}

	return ids, nil
}
