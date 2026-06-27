#!/bin/bash

echo "🍲 RecipeApp PostgreSQL Setup Script"
echo "================================="

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed."
    echo ""
    echo "Choose one of these installation methods:"
    echo ""
    echo "🍺 Option 1: Install with Homebrew (macOS)"
    echo "   brew install postgresql"
    echo "   brew services start postgresql"
    echo ""
    echo "🐧 Option 2: Install with apt (Ubuntu/Debian)"
    echo "   sudo apt update"
    echo "   sudo apt install postgresql postgresql-contrib"
    echo "   sudo systemctl start postgresql"
    echo "   sudo systemctl enable postgresql"
    echo ""
    echo "🐳 Option 3: Use Podman (Recommended for development)"
    echo "   podman run --name recipe-postgres \\"
    echo "     -e POSTGRES_PASSWORD=password \\"
    echo "     -e POSTGRES_USER=recipeapp \\"
    echo "     -e POSTGRES_DB=recipeapp \\"
    echo "     -p 5432:5432 \\"
    echo "     -v recipe-postgres-data:/var/lib/postgresql/data \\"
    echo "     postgres:13"
    echo ""
    exit 1
fi

echo "✅ PostgreSQL is installed!"

# Check if PostgreSQL is running
if ! pg_isready -q; then
    echo "🔄 Starting PostgreSQL..."
    
    # Try different start methods based on system
    if command -v brew &> /dev/null; then
        echo "🍺 Starting PostgreSQL with Homebrew..."
        brew services start postgresql
        sleep 3
    elif command -v systemctl &> /dev/null; then
        echo "🐧 Starting PostgreSQL with systemctl..."
        sudo systemctl start postgresql
        sleep 3
    elif command -v pg_ctl &> /dev/null; then
        echo "🗄️ Starting PostgreSQL with pg_ctl..."
        pg_ctl -D /usr/local/var/postgresql start -l /usr/local/var/postgresql/logfile
        sleep 3
    else
        echo "⚠️ Could not determine how to start PostgreSQL."
        echo "Please start PostgreSQL manually and re-run this script."
        exit 1
    fi
fi

# Check again if PostgreSQL is ready
if ! pg_isready -q; then
    echo "❌ Failed to start PostgreSQL. Please check the logs above."
    exit 1
fi

echo "✅ PostgreSQL is running!"

# Database setup
DB_NAME="recipeapp"
DB_USER="recipeapp"
DB_PASSWORD="password"
DB_HOST="localhost"
DB_PORT="5432"

echo ""
echo "📊 Setting up RecipeApp database..."
echo "================================="

# Check if database exists
if psql -h $DB_HOST -p $DB_PORT -U postgres -t -c "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" 2>/dev/null | grep -q 1; then
    echo "✅ Database '$DB_NAME' already exists"
else
    echo "🆕 Creating database '$DB_NAME'..."
    createdb -h $DB_HOST -p $DB_PORT -U postgres $DB_NAME
    echo "✅ Database created"
fi

# Check if user exists
if psql -h $DB_HOST -p $DB_PORT -U postgres -t -c "SELECT 1 FROM pg_roles WHERE rolname = '$DB_USER'" 2>/dev/null | grep -q 1; then
    echo "✅ User '$DB_USER' already exists"
else
    echo "👤 Creating user '$DB_USER'..."
    createuser -h $DB_HOST -p $DB_PORT -U postgres -s $DB_USER
    echo "✅ User created"
fi

# Grant permissions
echo "🔐 Setting permissions..."
psql -h $DB_HOST -p $DB_PORT -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null
psql -h $DB_HOST -p $DB_PORT -U postgres -c "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null

# Run database migrations
echo "📋 Running database migrations..."
if [ -f "migrations/001_initial_schema.sql" ]; then
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql
    echo "✅ Migrations completed"
else
    echo "❌ Migration file not found at migrations/001_initial_schema.sql"
fi

# Run seed data
echo "🌱 Running seed data..."
if [ -f "seed_data.go" ]; then
    export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
    go run seed_data.go
    echo "✅ Seed data completed"
else
    echo "❌ Seed data file not found"
fi

echo ""
echo "🎉 PostgreSQL setup completed!"
echo "================================="
echo "Database connection string: postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
echo ""
echo "🚀 You can now run the RecipeApp server:"
echo "   cd backend"
echo "   export DATABASE_URL=\"postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable\""
echo "   go run cmd/main.go"
echo ""
echo "🌐 Server will be available at: http://localhost:8080"
echo ""
echo "💡 Quick commands:"
echo "   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME  # Connect to database"
echo "   brew services list | grep postgresql     # Check service status (macOS)"