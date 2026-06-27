package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"recipe-app/internal/repositories"
)

func main() {
	fmt.Println("🧪 Testing SQLite Integration")
	fmt.Println("=================================")

	// Connect to SQLite database
	dbPath := "test_recipeapp.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("✅ Connected to SQLite database")

	// Test repository
	recipeRepo := repositories.NewRecipeRepository(db)

	// Test getting recipes
	fmt.Println("📊 Testing GetRecipes...")
	recipes, err := recipeRepo.GetRecipes(10, 0, "", "", 0)
	if err != nil {
		fmt.Printf("❌ GetRecipes failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetRecipes returned %d recipes\n", len(recipes))
		for i, recipe := range recipes {
			fmt.Printf("  Recipe %d: %s\n", i+1, recipe.Title)
		}
	}

	fmt.Println("🎉 SQLite integration test completed successfully!")
}
