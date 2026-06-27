package main

import (
	"log"

	"recipe-app/internal/models"
	"recipe-app/internal/repositories"
	"recipe-app/internal/storage"
)

func main() {
	// Connect to database
	dbURL := "postgres://recipeapp:password@localhost/recipeapp?sslmode=disable"
	db, err := storage.NewDB(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	recipeRepo := repositories.NewRecipeRepository(db.DB)
	userRepo := repositories.NewUserRepository(db.DB)

	// Create sample user
	user := &models.User{
		Email:     "demo@recipeapp.com",
		Username:  "demo",
		FirstName: "Demo",
		LastName:  "User",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
	}

	if err := userRepo.CreateUser(user); err != nil {
		log.Printf("Failed to create user (may already exist): %v", err)
	} else {
		log.Println("Created sample user")
	}

	// Create sample recipes
	recipes := []*models.Recipe{
		{
			Title:       "Spaghetti Bolognese",
			Description: "Classic Italian pasta dish with rich meat sauce",
			PrepTime:    15,
			CookTime:    45,
			Servings:    4,
			Difficulty:  "medium",
			Category:    "Pasta",
			Cuisine:     "Italian",
			ImageURL:    "/images/spaghetti-bolognese.jpg",
			Ingredients: []models.Ingredient{
				{Name: "Spaghetti", Amount: "400g", Unit: "grams", Notes: ""},
				{Name: "Ground beef", Amount: "500g", Unit: "grams", Notes: "80/20"},
				{Name: "Tomato sauce", Amount: "800ml", Unit: "milliliters", Notes: ""},
				{Name: "Onion", Amount: "1", Unit: "large", Notes: "diced"},
				{Name: "Garlic", Amount: "3", Unit: "cloves", Notes: "minced"},
				{Name: "Olive oil", Amount: "2", Unit: "tablespoons", Notes: ""},
			},
			Instructions: []models.Instruction{
				{Text: "Bring a large pot of salted water to boil and cook spaghetti according to package directions.", Duration: 15},
				{Text: "Heat olive oil in a large pan over medium heat. Add chopped onion and cook until translucent.", Duration: 5},
				{Text: "Add minced garlic and cook for another minute until fragrant.", Duration: 1},
				{Text: "Add ground beef and cook until browned, breaking it up with a wooden spoon.", Duration: 10},
				{Text: "Pour in tomato sauce and simmer for 15-20 minutes, stirring occasionally.", Duration: 20},
				{Text: "Season with salt, pepper, and Italian herbs to taste.", Duration: 1},
				{Text: "Drain pasta and toss with the bolognese sauce. Serve hot with grated Parmesan cheese.", Duration: 2},
			},
			Tags: []string{"italian", "pasta", "dinner", "meat"},
		},
		{
			Title:       "Caesar Salad",
			Description: "Fresh romaine lettuce with creamy Caesar dressing",
			PrepTime:    10,
			CookTime:    0,
			Servings:    2,
			Difficulty:  "easy",
			Category:    "Salad",
			Cuisine:     "American",
			ImageURL:    "/images/caesar-salad.jpg",
			Ingredients: []models.Ingredient{
				{Name: "Romaine lettuce", Amount: "2", Unit: "heads", Notes: "chopped"},
				{Name: "Parmesan cheese", Amount: "50g", Unit: "grams", Notes: "shaved"},
				{Name: "Croutons", Amount: "1", Unit: "cup", Notes: ""},
				{Name: "Caesar dressing", Amount: "100ml", Unit: "milliliters", Notes: ""},
				{Name: "Lemon", Amount: "1", Unit: "whole", Notes: "for juice"},
				{Name: "Black pepper", Amount: "1/4", Unit: "teaspoon", Notes: "freshly ground"},
			},
			Instructions: []models.Instruction{
				{Text: "Wash and chop romaine lettuce into bite-sized pieces.", Duration: 3},
				{Text: "Make Caesar dressing with anchovies, garlic, lemon juice, and egg yolk.", Duration: 5},
				{Text: "Toss lettuce with Caesar dressing until well coated.", Duration: 1},
				{Text: "Add shaved Parmesan cheese and croutons.", Duration: 1},
				{Text: "Season with fresh black pepper and serve immediately.", Duration: 1},
			},
			Tags: []string{"salad", "healthy", "quick", "vegetarian"},
		},
		{
			Title:       "Chicken Curry",
			Description: "Spicy and aromatic Indian curry with tender chicken",
			PrepTime:    20,
			CookTime:    35,
			Servings:    4,
			Difficulty:  "hard",
			Category:    "Curry",
			Cuisine:     "Indian",
			ImageURL:    "/images/chicken-curry.jpg",
			Ingredients: []models.Ingredient{
				{Name: "Chicken breast", Amount: "600g", Unit: "grams", Notes: "cut into cubes"},
				{Name: "Coconut milk", Amount: "400ml", Unit: "milliliters", Notes: "full fat"},
				{Name: "Onion", Amount: "2", Unit: "medium", Notes: "sliced"},
				{Name: "Garlic", Amount: "4", Unit: "cloves", Notes: "minced"},
				{Name: "Ginger", Amount: "2", Unit: "tablespoons", Notes: "grated"},
				{Name: "Curry powder", Amount: "2", Unit: "tablespoons", Notes: ""},
				{Name: "Turmeric", Amount: "1", Unit: "teaspoon", Notes: ""},
				{Name: "Chili powder", Amount: "1", Unit: "teaspoon", Notes: "adjust to taste"},
			},
			Instructions: []models.Instruction{
				{Text: "Marinate chicken with curry powder, turmeric, and salt for 15 minutes.", Duration: 15},
				{Text: "Heat oil in a large pan and sauté onions until golden brown.", Duration: 8},
				{Text: "Add ginger and garlic paste, cook for 2 minutes until fragrant.", Duration: 2},
				{Text: "Add marinated chicken and cook until sealed on all sides.", Duration: 5},
				{Text: "Add coconut milk and bring to a simmer. Cover and cook for 15 minutes.", Duration: 15},
				{Text: "Season with salt and adjust spices. Garnish with fresh cilantro.", Duration: 2},
			},
			Tags: []string{"indian", "curry", "spicy", "chicken"},
		},
		{
			Title:       "Greek Salad",
			Description: "Mediterranean salad with feta cheese and olives",
			PrepTime:    15,
			CookTime:    0,
			Servings:    3,
			Difficulty:  "easy",
			Category:    "Salad",
			Cuisine:     "Greek",
			ImageURL:    "/images/greek-salad.jpg",
			Ingredients: []models.Ingredient{
				{Name: "Cucumber", Amount: "2", Unit: "medium", Notes: "diced"},
				{Name: "Tomatoes", Amount: "3", Unit: "large", Notes: "cut into wedges"},
				{Name: "Red onion", Amount: "1", Unit: "medium", Notes: "thinly sliced"},
				{Name: "Feta cheese", Amount: "200g", Unit: "grams", Notes: "crumbled"},
				{Name: "Kalamata olives", Amount: "100g", Unit: "grams", Notes: "pitted"},
				{Name: "Olive oil", Amount: "3", Unit: "tablespoons", Notes: "extra virgin"},
				{Name: "Lemon juice", Amount: "2", Unit: "tablespoons", Notes: "fresh"},
				{Name: "Oregano", Amount: "1", Unit: "teaspoon", Notes: "dried"},
			},
			Instructions: []models.Instruction{
				{Text: "Dice cucumber and cut tomatoes into wedges.", Duration: 5},
				{Text: "Thinly slice red onion and soak in cold water for 5 minutes to reduce sharpness.", Duration: 5},
				{Text: "Combine vegetables in a large bowl.", Duration: 2},
				{Text: "Add crumbled feta cheese and olives.", Duration: 2},
				{Text: "Drizzle with olive oil and lemon juice, season with oregano, salt, and pepper.", Duration: 1},
				{Text: "Toss gently and let marinate for 10 minutes before serving.", Duration: 10},
			},
			Tags: []string{"greek", "mediterranean", "salad", "vegetarian"},
		},
		{
			Title:       "Chocolate Cake",
			Description: "Rich and moist chocolate cake with fudge frosting",
			PrepTime:    25,
			CookTime:    35,
			Servings:    8,
			Difficulty:  "hard",
			Category:    "Dessert",
			Cuisine:     "American",
			ImageURL:    "/images/chocolate-cake.jpg",
			Ingredients: []models.Ingredient{
				{Name: "All-purpose flour", Amount: "2", Unit: "cups", Notes: ""},
				{Name: "Cocoa powder", Amount: "3/4", Unit: "cup", Notes: "unsweetened"},
				{Name: "Sugar", Amount: "2", Unit: "cups", Notes: "granulated"},
				{Name: "Eggs", Amount: "2", Unit: "large", Notes: "room temperature"},
				{Name: "Butter", Amount: "1", Unit: "cup", Notes: "softened"},
				{Name: "Milk", Amount: "1", Unit: "cup", Notes: "whole"},
				{Name: "Vanilla extract", Amount: "2", Unit: "teaspoons", Notes: ""},
				{Name: "Baking powder", Amount: "1", Unit: "teaspoon", Notes: ""},
				{Name: "Salt", Amount: "1/2", Unit: "teaspoon", Notes: ""},
			},
			Instructions: []models.Instruction{
				{Text: "Preheat oven to 350°F (175°C) and grease two 9-inch round pans.", Duration: 5},
				{Text: "Sift together flour, cocoa powder, baking powder, and salt.", Duration: 3},
				{Text: "In a separate bowl, cream butter and sugar until light and fluffy.", Duration: 5},
				{Text: "Beat in eggs one at a time, then add vanilla extract.", Duration: 3},
				{Text: "Gradually add dry ingredients to wet mixture, alternating with milk.", Duration: 5},
				{Text: "Pour batter into prepared pans and bake for 30-35 minutes.", Duration: 35},
				{Text: "Cool cakes in pans for 10 minutes, then turn out onto wire rack.", Duration: 10},
			},
			Tags: []string{"dessert", "chocolate", "cake", "baking"},
		},
		{
			Title:       "Beef Tacos",
			Description: "Mexican-style tacos with seasoned ground beef",
			PrepTime:    15,
			CookTime:    20,
			Servings:    4,
			Difficulty:  "medium",
			Category:    "Mexican",
			Cuisine:     "Mexican",
			ImageURL:    "/images/beef-tacos.jpg",
			Ingredients: []models.Ingredient{
				{Name: "Ground beef", Amount: "500g", Unit: "grams", Notes: "80/20"},
				{Name: "Taco shells", Amount: "8", Unit: "pieces", Notes: "hard shells"},
				{Name: "Lettuce", Amount: "1", Unit: "head", Notes: "shredded"},
				{Name: "Tomatoes", Amount: "2", Unit: "medium", Notes: "diced"},
				{Name: "Cheddar cheese", Amount: "200g", Unit: "grams", Notes: "shredded"},
				{Name: "Sour cream", Amount: "200ml", Unit: "milliliters", Notes: ""},
				{Name: "Taco seasoning", Amount: "2", Unit: "tablespoons", Notes: ""},
				{Name: "Onion", Amount: "1", Unit: "medium", Notes: "diced"},
			},
			Instructions: []models.Instruction{
				{Text: "Brown ground beef in a large skillet over medium-high heat.", Duration: 8},
				{Text: "Add diced onion and cook until softened, about 3 minutes.", Duration: 3},
				{Text: "Stir in taco seasoning and 1/4 cup water, simmer until thickened.", Duration: 5},
				{Text: "Warm taco shells according to package directions.", Duration: 2},
				{Text: "Fill shells with seasoned beef and desired toppings.", Duration: 2},
			},
			Tags: []string{"mexican", "tacos", "beef", "quick"},
		},
	}

	// Create recipes
	for _, recipe := range recipes {
		if err := recipeRepo.CreateRecipe(recipe); err != nil {
			log.Printf("Failed to create recipe '%s': %v", recipe.Title, err)
		} else {
			log.Printf("Created recipe: %s", recipe.Title)
		}
	}

	log.Println("Seed data creation completed!")
}
