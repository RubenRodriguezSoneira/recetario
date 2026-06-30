package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"

	appmiddleware "recipe-app/internal/appmiddleware"
	"recipe-app/internal/database"
	"recipe-app/internal/handlers"
	"recipe-app/internal/logger"
	"recipe-app/internal/repositories"
)

func main() {
	log := logger.New()

	// Initialize SQLite database
	dbPath := "recipeapp.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Error("Failed to connect to SQLite database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("Connected to SQLite database", "path", dbPath)

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		log.Error("Failed to create tables", "error", err)
		os.Exit(1)
	}

	// Seed data
	if err := seedData(db); err != nil {
		log.Error("Failed to seed data", "error", err)
	} else {
		log.Info("Database seeded successfully")
	}

	// Initialize repositories
	recipeRepo := repositories.NewRecipeRepository(db)
	userRepo := repositories.NewUserRepository(db)
	collectionRepo := repositories.NewCollectionRepository(db)

	r := chi.NewRouter()

	// Initialize middleware
	rateLimiter := appmiddleware.NewRateLimiter(100, time.Minute)
	authService := appmiddleware.NewAuthService(getJWTSecret())

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(appmiddleware.CanonicalLoopbackHost)
	r.Use(appmiddleware.RequestLogger)
	r.Use(appmiddleware.ErrorHandler)
	r.Use(appmiddleware.CORS(appmiddleware.DefaultCORSConfig()))
	r.Use(appmiddleware.RateLimit(rateLimiter))
	r.Use(appmiddleware.SecurityHeaders)

	// Initialize handlers
	webHandler := handlers.NewWebHandler(userRepo)
	apiHandler := handlers.NewAPIHandler(recipeRepo)
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	collectionHandler := handlers.NewCollectionHandler(collectionRepo)
	fileHandler := handlers.NewFileHandler()
	searchHandler := handlers.NewSearchHandler(recipeRepo)
	ingredientHandler := handlers.NewIngredientHandler(recipeRepo)
	tagHandler := handlers.NewTagHandler(recipeRepo)

	// Routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.HandleRegister)
			r.Post("/login", authHandler.HandleLogin)
			r.Post("/logout", authHandler.HandleLogout)
			r.Post("/refresh", authHandler.HandleRefresh)
		})

		r.Route("/recipes", func(r chi.Router) {
			r.With(authService.OptionalAuthMiddleware).Get("/", apiHandler.HandleRecipes)
			r.With(authService.AuthMiddleware).Post("/", apiHandler.HandleCreateRecipe)
			r.Route("/{id}", func(r chi.Router) {
				r.With(authService.OptionalAuthMiddleware).Get("/", apiHandler.HandleRecipe)
				r.With(authService.AuthMiddleware).Put("/", apiHandler.HandleUpdateRecipe)
				r.With(authService.AuthMiddleware).Delete("/", apiHandler.HandleDeleteRecipe)

				// Recipe-scoped ingredients
				r.With(authService.OptionalAuthMiddleware).Get("/ingredients", ingredientHandler.HandleList)
				r.With(authService.AuthMiddleware).Post("/ingredients", ingredientHandler.HandleCreate)
				r.With(authService.AuthMiddleware).Put("/ingredients/order", ingredientHandler.HandleReorder)
				r.With(authService.AuthMiddleware).Put("/ingredients/{ingredientID}", ingredientHandler.HandleUpdate)
				r.With(authService.AuthMiddleware).Delete("/ingredients/{ingredientID}", ingredientHandler.HandleDelete)

				// Recipe-scoped tags
				r.With(authService.OptionalAuthMiddleware).Get("/tags", tagHandler.HandleListForRecipe)
				r.With(authService.AuthMiddleware).Post("/tags", tagHandler.HandleAdd)
				r.With(authService.AuthMiddleware).Delete("/tags/{tag}", tagHandler.HandleDelete)
			})
		})

		// Dedicated search and discovery
		r.With(authService.OptionalAuthMiddleware).Get("/search", searchHandler.HandleSearch)
		r.Get("/search/suggestions", searchHandler.HandleSuggestions)
		r.Get("/search/tags/popular", searchHandler.HandlePopularTags)

		// Tag catalog across all recipes
		r.Get("/tags", tagHandler.HandleListAll)

		// Image uploads
		r.With(authService.AuthMiddleware).Post("/upload", fileHandler.HandleUpload)
		r.With(authService.AuthMiddleware).Post("/upload/multiple", fileHandler.HandleMultiUpload)
		r.With(authService.AuthMiddleware).Delete("/upload/{filename}", fileHandler.HandleDelete)

		r.With(authService.AuthMiddleware).Get("/users/profile", userHandler.HandleProfile)
		r.With(authService.AuthMiddleware).Put("/users/profile", userHandler.HandleUpdateProfile)

		r.Route("/collections", func(r chi.Router) {
			r.Use(authService.AuthMiddleware)
			r.Get("/", collectionHandler.HandleList)
			r.Post("/", collectionHandler.HandleCreate)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", collectionHandler.HandleGet)
				r.Put("/", collectionHandler.HandleUpdate)
				r.Delete("/", collectionHandler.HandleDelete)
				r.Post("/recipes", collectionHandler.HandleAddRecipe)
				r.Delete("/recipes/{recipeID}", collectionHandler.HandleRemoveRecipe)
			})
		})
	})

	r.With(authService.OptionalAuthMiddleware).Get("/", webHandler.HandleIndex)
	r.Route("/recipes", func(r chi.Router) {
		r.With(authService.OptionalAuthMiddleware).Get("/", webHandler.HandleRecipes)
		r.With(authService.OptionalAuthMiddleware).Get("/new", webHandler.HandleNewRecipe)
		r.With(authService.OptionalAuthMiddleware).Get("/{id}", webHandler.HandleRecipeDetail)
		r.With(authService.OptionalAuthMiddleware).Get("/{id}/edit", webHandler.HandleEditRecipe)
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Serve uploaded images
	r.Get("/uploads/{filename}", fileHandler.ServeFile)

	// Serve static files
	fileServer := http.FileServer(http.Dir("web/static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	log.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return "dev-secret-change-in-production"
}

func createTables(db *sql.DB) error {
	if err := database.ApplySchema(db); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

func seedData(db *sql.DB) error {
	// Sample recipes data
	recipes := []map[string]interface{}{
		{
			"id":          "1",
			"title":       "Spaghetti Bolognese",
			"description": "Classic Italian pasta dish with rich meat sauce",
			"prep_time":   15,
			"cook_time":   45,
			"servings":    4,
			"difficulty":  "medium",
			"category":    "Pasta",
			"cuisine":     "Italian",
			"image_url":   "/images/spaghetti-bolognese.jpg",
		},
		{
			"id":          "2",
			"title":       "Chicken Curry",
			"description": "Spicy and aromatic Indian curry with tender chicken",
			"prep_time":   20,
			"cook_time":   35,
			"servings":    4,
			"difficulty":  "hard",
			"category":    "Curry",
			"cuisine":     "Indian",
			"image_url":   "/images/chicken-curry.jpg",
		},
		{
			"id":          "3",
			"title":       "Caesar Salad",
			"description": "Fresh romaine lettuce with creamy Caesar dressing",
			"prep_time":   10,
			"cook_time":   0,
			"servings":    2,
			"difficulty":  "easy",
			"category":    "Salad",
			"cuisine":     "American",
			"image_url":   "/images/caesar-salad.jpg",
		},
		{
			"id":          "4",
			"title":       "Beef Tacos",
			"description": "Mexican-style tacos with seasoned ground beef",
			"prep_time":   15,
			"cook_time":   20,
			"servings":    4,
			"difficulty":  "medium",
			"category":    "Mexican",
			"cuisine":     "Mexican",
			"image_url":   "/images/beef-tacos.jpg",
		},
		{
			"id":          "5",
			"title":       "Greek Salad",
			"description": "Mediterranean salad with feta cheese and olives",
			"prep_time":   15,
			"cook_time":   0,
			"servings":    3,
			"difficulty":  "easy",
			"category":    "Salad",
			"cuisine":     "Greek",
			"image_url":   "/images/greek-salad.jpg",
		},
		{
			"id":          "6",
			"title":       "Chocolate Cake",
			"description": "Rich and moist chocolate cake with fudge frosting",
			"prep_time":   25,
			"cook_time":   35,
			"servings":    8,
			"difficulty":  "hard",
			"category":    "Dessert",
			"cuisine":     "American",
			"image_url":   "/images/chocolate-cake.jpg",
		},
	}

	// Create the sample user first so seeded recipes can reference it.
	userQuery := `INSERT OR IGNORE INTO users (id, email, username, first_name, last_name, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	if _, err := db.Exec(userQuery,
		"user1",
		"demo@recipeapp.com",
		"demo",
		"Demo",
		"User",
		"$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // bcrypt hash of "password"
	); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Insert recipes owned by the sample user.
	for _, recipe := range recipes {
		query := `INSERT OR IGNORE INTO recipes (id, user_id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

		_, err := db.Exec(query,
			recipe["id"],
			"user1",
			recipe["title"],
			recipe["description"],
			recipe["prep_time"],
			recipe["cook_time"],
			recipe["servings"],
			recipe["difficulty"],
			recipe["category"],
			recipe["cuisine"],
			recipe["image_url"],
		)
		if err != nil {
			return fmt.Errorf("failed to insert recipe %s: %w", recipe["title"], err)
		}
	}

	return nil
}
