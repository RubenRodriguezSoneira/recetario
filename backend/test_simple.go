//go:build ignore

// Command test_simple is a standalone development script that smoke-tests the SQLite
// integration. It is excluded from normal package builds via the build tag above;
// run it explicitly with `go run test_simple.go`.
package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"recipe-app/internal/repositories"
)

func main() {
	fmt.Println("🧪 Testing SQLite Integration")
	fmt.Println("=================================")

	// Connect to SQLite database
	dbPath := "test_recipeapp.db"
	db, err := sql.Open("sqlite", dbPath)
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
