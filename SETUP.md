# 🍲 RecipeApp Development Setup

This document provides step-by-step instructions for setting up and running the RecipeApp locally.

## 🚀 Quick Start (Recommended)

### Prerequisites
- Go 1.25.6+
- PostgreSQL (recommended) or SQLite (fallback)

### Option 1: Using Docker with PostgreSQL (Easiest)
```bash
# Clone and navigate to the project
cd recipe-app

# Start PostgreSQL and RecipeApp
docker-compose up -d

# Wait for services to be ready
sleep 10

# Check that everything is running
docker-compose ps
```

The server will be available at: http://localhost:8080

### Option 2: Local PostgreSQL Setup
```bash
# Run the setup script
./backend/setup_postgres.sh

# After setup completes:
cd backend
export DATABASE_URL="postgres://recipeapp:password@localhost:5432/recipeapp?sslmode=disable"
go build -o recipe-server ./cmd
./recipe-server
```

### Option 3: Quick SQLite Setup (No additional setup needed)
```bash
# Just run the server (it will use mock data by default)
cd backend
go build -o recipe-server ./cmd
./recipe-server
```

## 📊 Database Setup Details

### PostgreSQL Database
- **Database Name**: recipeapp
- **User**: recipeapp  
- **Password**: password
- **Host**: localhost
- **Port**: 5432
- **Connection String**: `postgres://recipeapp:password@localhost:5432/recipeapp?sslmode=disable`

### SQLite Database (Local Development)
- **Database File**: recipeapp.db (created automatically)
- **No additional setup required**

## 🌐 Running the Server

The RecipeApp server provides:

### API Endpoints
- `GET /api/recipes` - List all recipes with filtering
- `POST /api/recipes` - Create new recipe
- `GET /api/recipes/{id}` - Get single recipe
- `PUT /api/recipes/{id}` - Update recipe
- `DELETE /api/recipes/{id}` - Delete recipe
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration

### Web Pages
- `GET /` - Homepage with recent recipes
- `GET /recipes` - Recipe list with advanced filtering
- `GET /recipes/new` - Create new recipe form
- `GET /recipes/{id}` - Recipe detail view
- `GET /static/*` - Static assets (CSS, JS, images)

### Features Implemented ✅

#### Core Features
- ✅ Recipe CRUD operations (Create, Read, Update, Delete)
- ✅ Ingredient and instruction management
- ✅ Advanced search and filtering (category, cuisine, difficulty, time)
- ✅ Server-side validation and error handling
- ✅ HTMX-powered dynamic UI (no page reloads)
- ✅ Responsive design with Tailwind CSS
- ✅ Multi-step recipe creation form
- ✅ Recipe detail view with ingredients and instructions
- ✅ Mock data for immediate testing

#### Advanced Features
- ✅ Full-text search support
- ✅ Real-time filtering with multiple criteria
- ✅ Recipe categories and cuisines
- ✅ Authentication flow (ready for user implementation)
- ✅ RESTful API with proper HTTP methods
- ✅ Template-based rendering system
- ✅ Static file serving
- ✅ CORS, rate limiting, and security middleware

#### Web UI Features
- ✅ Modern responsive design
- ✅ Interactive forms with dynamic field management
- ✅ Real-time search results
- ✅ Recipe cards with preview information
- ✅ Modal-based authentication
- ✅ Progress indicators and loading states
- ✅ Keyboard shortcuts and accessibility features

## 🛠 Development Commands

### Using Makefile
```bash
# Setup development environment
make setup-db

# Build and run server
make run

# Run tests
make test

# Lint code
make lint
```

### Docker Commands
```bash
# Build and start with Docker
make docker-build
make docker-up
make docker-down
make docker-clean

# Access database directly
make db-connect

# Run database migrations
make db-migrate

# Seed sample data
make db-seed
```

## 📁 Project Structure

```
recipe-app/
├── backend/
│   ├── cmd/                 # Application entry points
│   ├── internal/            # Private application code
│   │   ├── appmiddleware/  # HTTP middleware
│   │   ├── handlers/      # HTTP handlers
│   │   ├── logger/        # Logging utilities
│   │   ├── models/         # Data models
│   │   ├── repositories/   # Database access layer
│   │   └── storage/        # Database connections
│   ├── migrations/         # Database migrations
│   ├── web/              # Web templates and assets
│   │   ├── static/         # CSS, JS, images
│   │   └── templates/      # HTML templates
│   ├── go.mod             # Go module definition
│   └── Makefile          # Build and development commands
├── docker-compose.yml        # Docker service definitions
├── DEVELPMENT_PLAN.md     # Detailed implementation plan
└── README.md              # This file
```

## 🔧 Configuration

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string (optional, defaults to SQLite)
- `JWT_SECRET`: Secret for JWT authentication (default: dev-secret)

### Database Configuration
- Supports both PostgreSQL and SQLite
- Automatic database schema creation
- Seed data with sample recipes
- Connection pooling and health checks

## 🧪 Testing

### Unit Tests
```bash
cd backend
go test ./...
```

### Manual Testing
```bash
# Test API endpoints
curl -s http://localhost:8080/api/recipes

# Test with filters
curl -s "http://localhost:8080/api/recipes?difficulty=easy&cook_time=30"

# Create recipe
curl -X POST -H "Content-Type: application/json" \
  -d '{"title":"Test Recipe","description":"Test description","cook_time":30,"difficulty":"easy"}' \
  http://localhost:8080/api/recipes
```

## 🐛 Troubleshooting

### Common Issues and Solutions

#### Database Connection Issues
- **Error**: `failed to connect to database`
- **Solution**: Check PostgreSQL is running and connection string is correct
- **Fallback**: Use SQLite (automatic with no setup required)

#### Server Already Running
- **Error**: `bind: address already in use`
- **Solution**: Kill existing process:
  ```bash
  pkill -f recipe-server
  ```

#### Build Issues
- **Error**: Module dependency issues
- **Solution**: Run `go mod tidy` to fix dependencies

#### Permission Issues
- **Error**: Permission denied on database connection
- **Solution**: Check database user permissions and PostgreSQL pg_hba.conf

## 📚 Development Workflow

1. **Setup**: Run `make setup-db` to initialize database
2. **Development**: Use `make run` for live development
3. **Testing**: Run `make test` before committing changes
4. **Database Management**: Use provided database commands for migrations and seeding

## 🎯 Next Steps for Production

1. Set up environment variables for production
2. Configure database connection pooling
3. Implement HTTPS and security headers
4. Add comprehensive error logging
5. Set up CI/CD pipeline
6. Deploy with Docker Compose

## 📞 Support

For issues or questions:
1. Check the troubleshooting section above
2. Review the development plan (`DEVELOPMENT_PLAN.md`)
3. Run tests to isolate issues
4. Check logs: `tail -f logs/application.log`

---

**🍲 RecipeApp is ready for development!** 

Choose your setup method above and start building amazing recipes!