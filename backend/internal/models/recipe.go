package models

import (
	"fmt"
	"time"
)

type Recipe struct {
	ID           string        `json:"id" db:"id"`
	UserID       string        `json:"user_id" db:"user_id"`
	Title        string        `json:"title" db:"title"`
	Description  string        `json:"description" db:"description"`
	PrepTime     int           `json:"prep_time" db:"prep_time"` // minutes
	CookTime     int           `json:"cook_time" db:"cook_time"` // minutes
	Servings     int           `json:"servings" db:"servings"`
	Difficulty   string        `json:"difficulty" db:"difficulty"` // easy, medium, hard
	Category     string        `json:"category" db:"category"`
	Cuisine      string        `json:"cuisine" db:"cuisine"`
	Ingredients  []Ingredient  `json:"ingredients"`
	Instructions []Instruction `json:"instructions"`
	Tags         []string      `json:"tags"`
	ImageURL     string        `json:"image_url" db:"image_url"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

type Ingredient struct {
	ID       string `json:"id" db:"id"`
	RecipeID string `json:"recipe_id" db:"recipe_id"`
	Name     string `json:"name" db:"name"`
	Amount   string `json:"amount" db:"amount"`
	Unit     string `json:"unit" db:"unit"`
	Notes    string `json:"notes" db:"notes"`
	Position int    `json:"position" db:"position"`
}

type Instruction struct {
	ID          string `json:"id" db:"id"`
	RecipeID    string `json:"recipe_id" db:"recipe_id"`
	Text        string `json:"text" db:"text"`
	Position    int    `json:"position" db:"position"`
	Duration    int    `json:"duration" db:"duration"`       // optional time in minutes
	Temperature int    `json:"temperature" db:"temperature"` // optional temperature in F/C
}

type RecipeFilter struct {
	Category    string   `json:"category"`
	Cuisine     string   `json:"cuisine"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	MaxPrepTime int      `json:"max_prep_time"`
	MaxCookTime int      `json:"max_cook_time"`
	MinServings int      `json:"min_servings"`
	MaxServings int      `json:"max_servings"`
}

type SearchResult struct {
	Recipes []Recipe `json:"recipes"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	PerPage int      `json:"per_page"`
}

func (r *Recipe) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.PrepTime < 0 {
		return fmt.Errorf("prep time cannot be negative")
	}
	if r.CookTime < 0 {
		return fmt.Errorf("cook time cannot be negative")
	}
	if r.Servings < 0 {
		return fmt.Errorf("servings cannot be negative")
	}
	if r.Difficulty != "" && r.Difficulty != "easy" && r.Difficulty != "medium" && r.Difficulty != "hard" {
		return fmt.Errorf("difficulty must be easy, medium, or hard")
	}
	return nil
}

func (rf *RecipeFilter) Validate() error {
	if rf.Difficulty != "" && rf.Difficulty != "easy" && rf.Difficulty != "medium" && rf.Difficulty != "hard" {
		return fmt.Errorf("difficulty must be easy, medium, or hard")
	}
	if rf.MaxPrepTime < 0 {
		return fmt.Errorf("max prep time cannot be negative")
	}
	if rf.MaxCookTime < 0 {
		return fmt.Errorf("max cook time cannot be negative")
	}
	if rf.MinServings < 0 {
		return fmt.Errorf("min servings cannot be negative")
	}
	if rf.MaxServings < 0 {
		return fmt.Errorf("max servings cannot be negative")
	}
	if rf.MinServings > 0 && rf.MaxServings > 0 && rf.MinServings > rf.MaxServings {
		return fmt.Errorf("min servings cannot be greater than max servings")
	}
	return nil
}

func (i *Ingredient) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("ingredient name is required")
	}
	if i.Position < 0 {
		return fmt.Errorf("ingredient position cannot be negative")
	}
	return nil
}

func (i *Instruction) Validate() error {
	if i.Text == "" {
		return fmt.Errorf("instruction text is required")
	}
	if i.Position < 0 {
		return fmt.Errorf("instruction position cannot be negative")
	}
	if i.Duration < 0 {
		return fmt.Errorf("instruction duration cannot be negative")
	}
	if i.Temperature < 0 {
		return fmt.Errorf("instruction temperature cannot be negative")
	}
	return nil
}

func (sr *SearchResult) TotalPages() int {
	if sr.PerPage <= 0 {
		return 0
	}
	return (sr.Total + sr.PerPage - 1) / sr.PerPage
}
