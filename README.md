# 🍲 Recipe App

A cross-platform recipe management application with Go backend, HTMX web frontend, and Android mobile client.

## Project Structure

```
recipe-app/
├── backend/          # Go API server
│   ├── cmd/         # Application entry points
│   ├── internal/    # Private application code
│   ├── pkg/         # Public library code
│   └── configs/     # Configuration files
├── web/             # HTMX frontend
└── mobile/          # Android application
```

## Quick Start

### Backend

```bash
cd backend
go run cmd/main.go
```

The server will start on `:8080` with:

- API endpoints at `/api/*`
- Web interface at `/`

The SQLite driver is pure Go (`modernc.org/sqlite`), so no CGO or C compiler is
required — the server builds and runs with `CGO_ENABLED=0`.

### Database Setup

The backend uses **SQLite** and provisions itself automatically: on first run it
creates `recipeapp.db` in the `backend/` directory, applies the schema
(`internal/database/schema.sql`), and seeds sample data. No manual database setup
is required.

## API Endpoints

### Auth

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Log in and receive a JWT
- `POST /api/auth/refresh` - Refresh a JWT

### Recipes

- `GET /api/recipes` - List recipes (supports `search`, `difficulty`, `cook_time`, `limit`, `offset`)
- `POST /api/recipes` - Create a new recipe (auth required)
- `GET /api/recipes/{id}` - Get a specific recipe
- `PUT /api/recipes/{id}` - Update a recipe (owner only)
- `DELETE /api/recipes/{id}` - Delete a recipe (owner only)

### Collections (auth required)

- `GET /api/collections` - List your collections
- `POST /api/collections` - Create a collection
- `GET /api/collections/{id}` - Get a collection (owner or public)
- `PUT /api/collections/{id}` - Update a collection (owner only)
- `DELETE /api/collections/{id}` - Delete a collection (owner only)
- `POST /api/collections/{id}/recipes` - Add a recipe to a collection
- `DELETE /api/collections/{id}/recipes/{recipeID}` - Remove a recipe from a collection

### Profile (auth required)

- `GET /api/users/profile` - Get the current user's profile
- `PUT /api/users/profile` - Update the current user's profile

## Testing

```bash
cd backend
go test ./...
```

## Tech Stack

- **Backend**: Go, chi router, `database/sql` (no ORM)
- **Database**: SQLite (`modernc.org/sqlite`, pure Go — no CGO)
- **Auth**: JWT (`golang-jwt`) with bcrypt password hashing
- **Web**: HTMX, Tailwind CSS
- **Mobile**: Android (Kotlin)
