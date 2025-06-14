# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Running the application:**
```bash
# Using Docker Compose (includes PostgreSQL container)
docker-compose up

# For development with rebuilding:
docker-compose up --build

# Check logs (container name is "stampkeeper", not "golang"):
docker-compose logs stampkeeper
```

**Database operations:**
- PostgreSQL runs in separate container (`stampkeeper-db`)
- Database migrations run automatically on startup via `database.Migrate(db)`
- Sample data seeding runs automatically via `database.Seed(db)`
- Uses PostgreSQL database with connection string configuration
- Data persisted in `./postgres/` directory

**Environment variables:**
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Full PostgreSQL connection string (overrides individual DB vars)
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username (default: read .env file)
- `DB_PASSWORD` - PostgreSQL password (default: read .env file)
- `DB_NAME` - PostgreSQL database name (default: stampkeeper)
- `DB_SSLMODE` - PostgreSQL SSL mode (default: disable)

## Architecture Overview

**Multi-layered Go web application:**
- `main.go` - Entry point with config loading, database connection, and server startup
- `internal/config/` - Environment-based configuration management
- `internal/database/` - PostgreSQL connection, migrations, and seeding
- `internal/models/` - Core domain models (Stamp, StampInstance, StorageBox, Tag)
- `internal/handlers/` - HTTP request handlers organized by domain
- `internal/services/` - Business logic layer
- `internal/router/` - Gorilla Mux routing with custom template functions
- `internal/middleware/` - Session management and user preferences

**Frontend architecture (HTML over the wire):**
- Server-rendered HTML application using Go templates
- Main page served from `templates/index.html` with user preferences injection
- HTML templates in `templates/` for all UI components and fragments
- HTMX for dynamic interactions and partial page updates
- Minimal vanilla JavaScript for essential UI behaviors only
- Custom CSS in `static/css/` for styling
- User preferences stored in URL-encoded cookies and applied server-side

**Key domain concepts:**
- **Stamp** - Abstract stamp design with metadata (Scott numbers, series, etc.)
- **StampInstance** - Physical copies grouped by condition/quantity/location
- **StorageBox** - Organizational containers for physical stamps
- **Tag** - Categorization system for stamps

**API structure:**
- RESTful JSON API under `/api/` prefix for data operations and preferences
- View endpoints under `/views/` return server-rendered HTML fragments
- HTMX endpoints under `/htmx/` for interactive UI updates
- Static files served from `/static/` (CSS, JS, images)
- Template rendering uses custom functions: `substr`, `deref`, `json`, `eq`, `add`
- Main application route `/` serves dynamic template with user preferences

**User preferences system:**
- Cookie-based storage with URL encoding for JSON data
- Automatic injection into request context via session middleware
- Preferences include: default view (gallery/list), sort order, sort direction, items per page
- Server-side template rendering ensures UI reflects saved preferences
- Real-time updates via HTMX form submissions

**Database design:**
- PostgreSQL with foreign key constraints
- Soft deletes using `date_deleted` fields
- Calculated fields like `is_owned` for stamps based on instances
- Uses numbered parameter placeholders ($1, $2, etc.) for SQL queries