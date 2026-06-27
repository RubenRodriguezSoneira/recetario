package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	appmiddleware "recipe-app/internal/appmiddleware"
	"recipe-app/internal/handlers"
	"recipe-app/internal/logger"
	"recipe-app/internal/repositories"
)

func main() {
	log := logger.New()

	// Initialize SQLite database
	dbPath := "recipeapp.db"
	db, err := sql.Open("sqlite3", dbPath)
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

	r := chi.NewRouter()

	// Initialize middleware
	rateLimiter := appmiddleware.NewRateLimiter(100, time.Minute)
	authService := appmiddleware.NewAuthService(getJWTSecret())

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(appmiddleware.RequestLogger)
	r.Use(appmiddleware.ErrorHandler)
	r.Use(appmiddleware.CORS(appmiddleware.DefaultCORSConfig()))
	r.Use(appmiddleware.RateLimit(rateLimiter))
	r.Use(appmiddleware.SecurityHeaders)

	// Initialize handlers
	webHandler := handlers.NewWebHandler()
	apiHandler := handlers.NewAPIHandler(recipeRepo)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler()

	// Routes
	r.Get("/", webHandler.HandleIndex)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.HandleRegister)
			r.Post("/login", authHandler.HandleLogin)
			r.Post("/refresh", authHandler.HandleRefresh)
		})

		r.Route("/recipes", func(r chi.Router) {
			r.With(authService.OptionalAuthMiddleware).Get("/", apiHandler.HandleRecipes)
			r.With(authService.AuthMiddleware).Post("/", apiHandler.HandleCreateRecipe)
			r.Route("/{id}", func(r chi.Router) {
				r.With(authService.OptionalAuthMiddleware).Get("/", apiHandler.HandleRecipe)
				r.With(authService.AuthMiddleware).Put("/", apiHandler.HandleUpdateRecipe)
				r.With(authService.AuthMiddleware).Delete("/", apiHandler.HandleDeleteRecipe)
			})
		})

		r.With(authService.AuthMiddleware).Get("/users/profile", userHandler.HandleProfile)
		r.With(authService.AuthMiddleware).Put("/users/profile", userHandler.HandleUpdateProfile)
	})

	r.Route("/recipes", func(r chi.Router) {
		r.Get("/", webHandler.HandleRecipes)
		r.With(authService.AuthMiddleware).Get("/new", webHandler.HandleNewRecipe)
		r.Get("/{id}", webHandler.HandleRecipeDetail)
	})

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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			first_name TEXT,
			last_name TEXT,
			password TEXT NOT NULL,
			avatar_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS recipes (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			prep_time INTEGER,
			cook_time INTEGER,
			servings INTEGER,
			difficulty TEXT CHECK (difficulty IN ('easy', 'medium', 'hard')),
			category TEXT,
			cuisine TEXT,
			image_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ingredients (
			id TEXT PRIMARY KEY,
			recipe_id TEXT NOT NULL,
			name TEXT NOT NULL,
			amount TEXT,
			unit TEXT,
			notes TEXT,
			position INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS instructions (
			id TEXT PRIMARY KEY,
			recipe_id TEXT NOT NULL,
			text TEXT NOT NULL,
			position INTEGER NOT NULL,
			duration INTEGER,
			temperature INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS recipe_tags (
			recipe_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			PRIMARY KEY (recipe_id, tag)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
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

	// Insert recipes
	for _, recipe := range recipes {
		query := `INSERT INTO recipes (id, title, description, prep_time, cook_time, servings, difficulty, category, cuisine, image_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

		_, err := db.Exec(query,
			recipe["id"],
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

	// Create sample user
	userQuery := `INSERT INTO users (id, email, username, first_name, last_name, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := db.Exec(userQuery,
		"user1",
		"demo@recipeapp.com",
		"demo",
		"Demo",
		"User",
		"$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
