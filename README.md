# StampKeeper

A modern web application for managing and organizing stamp collections. Built with Go and PostgreSQL, featuring server-rendered HTML with HTMX for dynamic interactions.

## Features

- **Stamp Collection Management**: Track stamps with detailed metadata including Scott numbers, series, and descriptions
- **Physical Instance Tracking**: Manage multiple copies of stamps with condition, quantity, and storage location details
- **Storage Organization**: Organize stamps using customizable storage boxes
- **Tagging System**: Categorize stamps with flexible tags for easy searching and filtering
- **Multiple Views**: Switch between gallery and list views with user preferences
- **Collection Statistics**: View comprehensive stats about your collection
- **Search & Filter**: Find stamps by various criteria including tags, boxes, and ownership status
- **Responsive Design**: Works on desktop and mobile devices

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Running the Application

1. Clone the repository:
```bash
git clone https://github.com/jeepinbird/stampkeeper
cd stampkeeper
```

2. Start the application with Docker Compose:
```bash
docker-compose up
```

The application will be available at `http://localhost:8080`

### Using StampKeeper

Once the application is running, you can:

1. **Browse Your Collection**: The main page shows all stamps in your collection. Use the view toggle to switch between gallery (grid) and list views.

2. **Search and Filter**: 
   - Use the search bar to find stamps by name, description, or Scott number
   - Filter by tags using the tag buttons
   - Filter by storage box or ownership status
   - Use the "Show Only Owned" toggle to see only stamps you physically own

3. **View Stamp Details**: Click on any stamp to see detailed information including:
   - High-resolution images
   - Complete metadata (Scott numbers, series, year, etc.)
   - Your physical copies with condition and location
   - Notes and tags
   - Related stamps in the same series

4. **Manage Your Collection**:
   - Add new stamp instances by clicking "Add Copy" on stamp detail pages
   - Edit existing copies to update condition, quantity, or storage location
   - Add tags to categorize stamps
   - Make notes about individual stamps

5. **Organize Storage**: 
   - Create and manage storage boxes to organize your physical stamps
   - Assign stamps to specific boxes for easy location
   - View box contents and statistics such as total stamps and owned copies
   
6. **Customize Preferences**: Use the settings page to:
   - Set default view preferences (gallery vs list)
   - Configure sorting options
   - Adjust items per page
   - Manage display preferences

### Development Mode

For development with automatic rebuilding:
```bash
docker-compose up --build
```

Check application logs:
```bash
docker-compose logs stampkeeper
```

## Architecture

- **Backend**: Go web server using Gorilla Mux router
- **Database**: PostgreSQL with automatic migrations and sample data seeding
- **Frontend**: Server-rendered HTML templates with HTMX for dynamic updates
- **Styling**: Custom CSS with minimal JavaScript
- **Session Management**: Cookie-based user preferences

## Database

- PostgreSQL runs in a separate Docker container
- Database migrations run automatically on startup
- Sample data is seeded for immediate use
- Data is persisted in the `./postgres/` directory

## Configuration

Environment variables can be configured in `.env` file:

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Full PostgreSQL connection string
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - Database name (default: stampkeeper)
- `DB_SSLMODE` - SSL mode (default: disable)

## Project Structure

```
├── main.go                # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and utilities
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # Session and preference middleware
│   ├── models/            # Domain models
│   ├── router/            # Route definitions
│   └── services/          # Business logic
├── templates/             # HTML templates
├── static/                # CSS, JavaScript, and static assets
└── docker-compose.yml     # Docker configuration
```

## Contributing

1. Make changes to the codebase
2. Test locally using `docker-compose up --build`
3. Submit a pull request

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.