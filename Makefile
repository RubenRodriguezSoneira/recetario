.PHONY: help build run test lint db-reset

# Default target
help:
	@echo "🍲 RecipeApp Development Commands"
	@echo "=============================="
	@echo ""
	@echo "📱 Development:"
	@echo "  make run          - Build and run the server (creates the SQLite db on first run)"
	@echo "  make build        - Build the server binary"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run Go linter"
	@echo ""
	@echo "📊 Database (SQLite):"
	@echo "  make db-reset     - Delete the local SQLite database (recreated on next run)"
	@echo ""
	@echo "Example usage:"
	@echo "  make run"

build:
	@echo "🔨 Building RecipeApp..."
	cd backend && go build -o recipe-server ./cmd

run: build
	@echo "🚀 Starting RecipeApp server..."
	cd backend && ./recipe-server

test:
	@echo "🧪 Running tests..."
	cd backend && go test ./...

lint:
	@echo "🔍 Running linter..."
	cd backend && golangci-lint run

# Database commands (SQLite)
# The server creates and seeds backend/recipeapp.db automatically on startup via
# database.ApplySchema, so there is no separate migrate/seed step.
db-reset:
	@echo "⚠️ Resetting the local SQLite database (dangerous)..."
	@echo "This deletes backend/recipeapp.db; it is recreated and seeded on the next run."
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]
	rm -f backend/recipeapp.db
	@echo "Database removed. Run 'make run' to recreate and seed it."
