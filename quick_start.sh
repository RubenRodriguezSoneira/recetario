#!/bin/bash

echo "🍲 RecipeApp Quick Start (SQLite Version)"
echo "===================================="

# Create SQLite database if it doesn't exist
if [ ! -f "recipeapp.db" ]; then
    echo "📊 Creating SQLite database..."
    touch recipeapp.db
fi

echo "🚀 Starting RecipeApp server..."

# Set environment variables for SQLite
export DATABASE_URL="recipeapp.db"

# Build and run the server
cd backend
go build -o recipe-server ./cmd
./recipe-server