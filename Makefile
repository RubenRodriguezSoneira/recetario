.PHONY: help build run test docker-build docker-up docker-down docker-clean setup-db lint

# Default target
help:
	@echo "🍲 RecipeApp Development Commands"
	@echo "=============================="
	@echo ""
	@echo "📱 Development:"
	@echo "  make setup-db     - Setup PostgreSQL database"
	@echo "  make run          - Build and run the server"
	@echo "  make build        - Build the server binary"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run Go linter"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start services with Docker Compose"
	@echo "  make docker-down   - Stop services"
	@echo "  make docker-clean  - Clean Docker resources"
	@echo ""
	@echo "📊 Database:"
	@echo "  make db-connect   - Connect to PostgreSQL database"
	@echo "  make db-migrate   - Run database migrations"
	@echo "  make db-seed      - Run seed data"
	@echo "  make db-reset     - Reset database (dangerous)"
	@echo ""
	@echo "Example usage:"
	@echo "  make setup-db && make run"

# Development commands
setup-db:
	@echo "🔄 Setting up PostgreSQL..."
	@./setup_postgres.sh

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

# Docker commands
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t recipeapp .

docker-up:
	@echo "🐳 Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "🐳 Stopping Docker services..."
	docker-compose down

docker-clean:
	@echo "🧹 Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# Database commands
db-connect:
	@echo "📊 Connecting to RecipeApp database..."
	psql -h localhost -p 5432 -U recipeapp -d recipeapp

db-migrate:
	@echo "📋 Running database migrations..."
	PGPASSWORD=password psql -h localhost -p 5432 -U recipeapp -d recipeapp -f migrations/001_initial_schema.sql

db-seed:
	@echo "🌱 Running seed data..."
	cd backend && DATABASE_URL="postgres://recipeapp:password@localhost:5432/recipeapp?sslmode=disable" go run seed_data.go

db-reset:
	@echo "⚠️ Resetting RecipeApp database (dangerous)..."
	@echo "This will delete all data and recreate the database."
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]
	PGPASSWORD=password dropdb -h localhost -p 5432 -U postgres --if-exists recipeapp
	PGPASSWORD=password createdb -h localhost -p 5432 -U postgres recipeapp
	@echo "Database reset. Run 'make setup-db' to recreate."
	@echo "Then run 'make db-seed' to populate with sample data."

# Development workflow
dev-setup: setup-db run
	@echo "🎯 Complete development setup"