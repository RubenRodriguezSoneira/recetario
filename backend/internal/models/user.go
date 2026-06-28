package models

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Username  string    `json:"username" db:"username"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Password  string    `json:"-" db:"password_hash"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RecipeCollection struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	RecipeIDs   []string  `json:"recipe_ids"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks the required fields of a collection.
func (c *RecipeCollection) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("collection name is required")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("collection name must be 100 characters or less")
	}
	return nil
}

type Rating struct {
	ID        string    `json:"id" db:"id"`
	RecipeID  string    `json:"recipe_id" db:"recipe_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Score     int       `json:"score" db:"score"` // 1-5
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type NutritionInfo struct {
	ID          string  `json:"id" db:"id"`
	RecipeID    string  `json:"recipe_id" db:"recipe_id"`
	Calories    float64 `json:"calories" db:"calories"`
	Protein     float64 `json:"protein" db:"protein"`
	Carbs       float64 `json:"carbs" db:"carbs"`
	Fat         float64 `json:"fat" db:"fat"`
	Fiber       float64 `json:"fiber" db:"fiber"`
	Sugar       float64 `json:"sugar" db:"sugar"`
	Sodium      float64 `json:"sodium" db:"sodium"`
	ServingSize string  `json:"serving_size" db:"serving_size"`
}
