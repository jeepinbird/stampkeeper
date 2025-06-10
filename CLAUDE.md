# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Running the application:**
```bash
# Using Docker Compose (PostgreSQL required)
docker-compose up

# Note: Local development requires PostgreSQL connection
# Set DB_HOST=localhost to run locally with external PostgreSQL
```

**Database operations:**
- Database migrations run automatically on startup via `database.Migrate(db)`
- Sample data seeding runs automatically via `database.Seed(db)`
- Uses PostgreSQL database with connection string configuration

**Environment variables:**
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Full PostgreSQL connection string (overrides individual DB vars)
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username (default: bird)
- `DB_PASSWORD` - PostgreSQL password (default: birder13)
- `DB_NAME` - PostgreSQL database name (default: stampkeeper)
- `DB_SSLMODE` - PostgreSQL SSL mode (default: disable)

## Architecture Overview

**Multi-layered Go web application:**
- `main.go` - Entry point with config loading, database connection, and server startup
- `internal/config/` - Environment-based configuration management
- `internal/database/` - SQLite connection, migrations, and seeding
- `internal/models/` - Core domain models (Stamp, StampInstance, StorageBox, Tag)
- `internal/handlers/` - HTTP request handlers organized by domain
- `internal/services/` - Business logic layer
- `internal/router/` - Gorilla Mux routing with custom template functions

**Frontend architecture:**
- Single-page application served from `static/index.html`
- HTML templates in `templates/` for server-rendered fragments
- Custom CSS in `static/css/`
- Vanilla JavaScript in `static/js/`
- HTMX-style view endpoints return HTML fragments

**Key domain concepts:**
- **Stamp** - Abstract stamp design with metadata (Scott numbers, series, etc.)
- **StampInstance** - Physical copies grouped by condition/quantity/location
- **StorageBox** - Organizational containers for physical stamps
- **Tag** - Categorization system for stamps

**API structure:**
- RESTful JSON API under `/api/` prefix
- View endpoints under `/views/` return HTML fragments
- Static files served from `/static/`
- Template rendering uses custom functions: `substr`, `deref`, `json`, `eq`, `add`

**Database design:**
- PostgreSQL with foreign key constraints
- Soft deletes using `date_deleted` fields
- Calculated fields like `is_owned` for stamps based on instances
- Uses numbered parameter placeholders ($1, $2, etc.) for SQL queries