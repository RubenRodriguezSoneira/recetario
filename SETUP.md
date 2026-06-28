# 🍲 RecipeApp Development Setup

This document provides step-by-step instructions for setting up and running the RecipeApp locally.

## 🚀 Quick Start (Recommended)

### Prerequisites
- Go 1.25.6+
- A C compiler (e.g. `gcc`) with `CGO_ENABLED=1` — required to build the SQLite driver

### Run the server
```bash
cd backend
go build -o recipe-server ./cmd
./recipe-server
```

The server will be available at: http://localhost:8080

On first run it creates `recipeapp.db` in the `backend/` directory, applies the
schema (`internal/database/schema.sql`), and seeds sample data — no additional
setup is required.

## 📊 Database Setup Details

### SQLite Database (Local Development)
- **Database File**: `recipeapp.db` (created automatically in `backend/`)
- **Schema**: applied at startup from `internal/database/schema.sql`
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
# List available targets
make help

# Build the server binary
make build

# Build and run the server
make run

# Run tests
make test

# Lint code
make lint

# Delete the local SQLite database (recreated and seeded on next run)
make db-reset
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
│   ├── web/              # Web templates and assets
│   │   ├── static/         # CSS, JS, images
│   │   └── templates/      # HTML templates
│   ├── go.mod             # Go module definition
│   └── Makefile          # Build and development commands
├── DEVELPMENT_PLAN.md     # Detailed implementation plan
└── README.md              # This file
```

## 🔧 Configuration

### Environment Variables
- `JWT_SECRET`: Secret for JWT authentication (default: dev-secret)

### Database Configuration
- Uses **SQLite** (`recipeapp.db`, created automatically in `backend/`)
- Automatic schema creation at startup
- Seed data with sample recipes

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

#### Database Issues
- **Error**: `failed to open database` or schema errors
- **Solution**: Ensure the process can write to the `backend/` directory. Delete
  `recipeapp.db` (or run `make db-reset`) to recreate it from scratch.

#### Server Already Running
- **Error**: `bind: address already in use`
- **Solution**: Stop the existing process listening on `:8080`, then start again.

#### Build Issues
- **Error**: Module dependency issues
- **Solution**: Run `go mod tidy` to fix dependencies

#### CGO / SQLite Build Errors
- **Error**: `cgo: C compiler ... not found` or SQLite driver fails to build
- **Solution**: Install a C compiler (e.g. `gcc`) and build with `CGO_ENABLED=1`

## 📚 Development Workflow

1. **Setup**: Build the server with `make build` (the SQLite database is created automatically on first run)
2. **Development**: Use `make run` for live development
3. **Testing**: Run `make test` before committing changes
4. **Database Management**: Use `make db-reset` to recreate the local SQLite database

## 🎯 Next Steps for Production

1. Set up environment variables for production
2. Implement HTTPS and security headers
3. Add comprehensive error logging
4. Set up CI/CD pipeline
5. Deploy the compiled binary alongside its `recipeapp.db`

## 📞 Support

For issues or questions:
1. Check the troubleshooting section above
2. Review the development plan (`DEVELOPMENT_PLAN.md`)
3. Run tests to isolate issues
4. Check logs: `tail -f logs/application.log`

---

**🍲 RecipeApp is ready for development!** 

Choose your setup method above and start building amazing recipes!