# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Running the application:**
```bash
# Using Docker Compose (includes PostgreSQL container)
docker-compose up

# For development with rebuilding:
docker-compose up --build

# Stop all containers:
docker-compose down

# Check logs (container name is "stampkeeper", not "golang"):
docker-compose logs stampkeeper
docker-compose logs stampkeeper-db
```

**Native Go commands (when not using Docker):**
```bash
# Run the application directly:
go run main.go

# Build the application:
go build -o stampkeeper main.go

# Run the built binary:
./stampkeeper

# Install dependencies:
go mod download

# Update dependencies:
go mod tidy
```

**Database operations:**
- PostgreSQL runs in separate container (`stampkeeper-db`)
- Database migrations run automatically on startup via `database.Migrate(db)` in main.go
- Sample data seeding runs automatically via `database.Seed(db)` in main.go
- Uses PostgreSQL database with connection string configuration
- Data persisted in `./postgres/` directory
- To reset database: stop containers, delete `./postgres/` directory, restart containers

**Environment variables:**
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Full PostgreSQL connection string (overrides individual DB vars)
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username (read from .env file)
- `DB_PASSWORD` - PostgreSQL password (read from .env file)
- `DB_NAME` - PostgreSQL database name (default: stampkeeper)
- `DB_SSLMODE` - PostgreSQL SSL mode (default: disable)
- `POSTGRES_USER` - Used by postgres container (set in .env)
- `POSTGRES_PASSWORD` - Used by postgres container (set in .env)
- `POSTGRES_DB` - Used by postgres container (set in .env)

## Architecture Overview

**Multi-layered Go web application:**
- `main.go` - Entry point: loads config, connects to database, runs migrations/seeding, sets up router, starts HTTP server
- `internal/config/` - Environment-based configuration management
- `internal/database/` - PostgreSQL connection, migrations (table creation), seeding (sample data), and query builder utilities
- `internal/models/` - Core domain models (Stamp, StampInstance, StorageBox, Tag, Stats, UserPreferences, view models)
- `internal/handlers/` - HTTP request handlers organized by domain:
  - `stamps.go` - JSON API for stamp CRUD operations
  - `instances.go` - JSON API for stamp instance CRUD
  - `boxes.go` - JSON API for storage box operations
  - `tags.go` - JSON API for tag operations
  - `stats.go` - JSON API for collection statistics
  - `views.go` - HTML fragment rendering for main views
  - `htmx.go` - HTMX-specific endpoints for dynamic UI updates
  - `preferences.go` - User preference management
- `internal/services/` - Business logic layer that handlers call:
  - `stamps.go` - Stamp querying, filtering, sorting, pagination logic
  - `instances.go` - Instance management logic
  - `boxes.go` - Box operations with stamp counts
  - `tags.go` - Tag operations with usage counts
  - `stats.go` - Statistics calculation
- `internal/router/` - Gorilla Mux routing setup with custom template functions
- `internal/middleware/` - Session middleware for cookie-based user preferences

**Request flow:**
1. HTTP request → Router (internal/router/router.go)
2. Middleware processes request (session handling)
3. Handler receives request (internal/handlers/)
4. Handler calls service layer for business logic (internal/services/)
5. Service queries database using sql.DB
6. Handler renders template or returns JSON
7. Response sent to client

**Frontend architecture (HTML over the wire):**
- Server-rendered HTML application using Go templates
- Main page served from `templates/index.html` with user preferences injection
- HTML templates in `templates/` for all UI components and fragments:
  - Main views: index.html, gallery-view.html, list-view.html, stamp-detail.html, settings.html
  - Partial templates: _gallery-page.html, _list-page.html, pagination.html
  - Component templates: box-list.html, boxes-table.html, new-stamp-form.html, new-instance-row.html
  - Stamp detail sections: stamp-details-section.html, stamp-image-section.html, stamp-notes-section.html, stamp-tags-section.html, your-copies-section.html
- HTMX for dynamic interactions and partial page updates
- Minimal vanilla JavaScript for essential UI behaviors only
- Custom CSS in `static/css/` for styling (custom.css, settings.css, stamp-detail.css)
- JavaScript files in `static/js/` for component behavior (alpine-components.js, new-stamp.js, stamp-instance.js)
- User preferences stored in URL-encoded cookies and applied server-side

**Key domain concepts:**
- **Stamp** - Abstract stamp design with metadata (Scott numbers, series, etc.)
- **StampInstance** - Physical copies grouped by condition/quantity/location
- **StorageBox** - Organizational containers for physical stamps
- **Tag** - Categorization system for stamps
- **Stats** - Collection statistics and summary data

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

**Database schema:**
- `stamps` - Abstract stamp designs with metadata (scott_number, series, issue_date, notes, image_url)
- `stamp_instances` - Physical copies with quantity, condition, box location (foreign keys to stamps and storage_boxes)
- `storage_boxes` - Organizational containers for stamps
- `tags` - Tag names for categorization
- `stamp_tags` - Many-to-many join table linking stamps to tags

**Database design patterns:**
- PostgreSQL with foreign key constraints (CASCADE on delete for stamps, SET NULL for boxes)
- Soft deletes using `date_deleted` timestamp fields (nullable)
- Calculated fields like `is_owned` boolean on stamps (true if any instances exist)
- UUID primary keys (VARCHAR(36))
- Uses numbered parameter placeholders ($1, $2, etc.) for SQL queries in Go
- Unique constraint on stamp_instances(stamp_id, condition, box_id) prevents duplicates
- Auto-migration on startup creates tables if they don't exist

**Key dependencies:**
- `github.com/gorilla/mux` - HTTP router
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- Go 1.24.3