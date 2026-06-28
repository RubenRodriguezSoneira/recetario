#!/bin/bash

echo "🍲 RecipeApp Quick Start (SQLite Version)"
echo "===================================="

echo "🚀 Starting RecipeApp server..."

# Build and run the server.
# On first run it creates backend/recipeapp.db, applies the schema, and seeds data.
cd backend
go build -o recipe-server ./cmd
./recipe-server