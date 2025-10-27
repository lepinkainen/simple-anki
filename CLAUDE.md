# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Simple Anki is a lightweight web-based flashcard application for language learning with spaced repetition. It's built with:
- **Backend**: Go with embedded SQLite database
- **Frontend**: Vanilla HTML/CSS/JavaScript (embedded in the binary, no build step)
- **Architecture**: Single-file binary with all static assets embedded via `//go:embed`

## Common Commands

### Building and Running

```bash
# Install dependencies
go get github.com/mattn/go-sqlite3

# Build the application
go build -o simple-anki

# Run the application (default: port 8080, database flashcards.db)
./simple-anki

# Run with custom settings
./simple-anki -port 8080 -db flashcards.db
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestFunctionName ./...
```

### Development

```bash
# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Check for Go modules issues
go mod tidy
go mod verify
```

## Architecture

### File Structure

- **main.go** (main.go:1): Entry point. Sets up HTTP server, embeds static files, and initializes routing
- **database.go** (database.go:1): All database operations and spaced repetition (SM-2) algorithm implementation
- **handlers.go** (handlers.go:1): HTTP handlers for REST API endpoints
- **static/index.html**: Complete web UI (embedded in binary via go:embed)

### Key Components

**Database Layer (database.go)**
- Uses SQLite with a single `cards` table
- Core models: `Card` struct with SRS fields (ease, interval, next_review)
- Database functions follow pattern: `GetCard()`, `GetAllCards()`, `CreateCard()`, `UpdateCard()`, `DeleteCard()`
- Spaced repetition: `CalculateNextReview()` implements simplified SM-2 algorithm (database.go:201)

**HTTP Layer (handlers.go)**
- REST API with JSON responses
- Handler pattern: `CardsHandler` (collection), `CardHandler` (individual resource)
- Helper functions: `respondJSON()`, `respondError()` for consistent responses

**Static Assets**
- Embedded using `//go:embed static/*` directive (main.go:10)
- No build step required - vanilla JS frontend loads directly from embedded filesystem

### REST API Endpoints

All endpoints are defined in main.go:24-31:
- `GET/POST /api/cards` - List/create cards, optional `?deck=name` filter
- `GET/PUT/DELETE /api/cards/{id}` - Single card operations
- `GET /api/decks` - List all unique deck names
- `GET/POST /api/review` - Get due cards and submit review scores

### Spaced Repetition Logic

The SM-2 algorithm implementation (database.go:201-232):
- New cards start with ease=2.5, interval=0
- Score < 3 (Again/Hard): Reset to interval=0, decrease ease by 0.2
- Score >= 3 (Good/Easy): Progress intervals (1 day → 6 days → interval * ease)
- Ease adjustments: Hard (-0.15), Good (0), Easy (+0.15), clamped to [1.3, 2.5]
- Failed reviews schedule next review in 1 minute for immediate retry

## Development Guidelines

### Adding New Features

1. **Database changes**: Update schema in `InitDB()` (database.go:29), add migration if needed
2. **API changes**: Add handler in handlers.go, register route in main.go
3. **Frontend changes**: Modify static/index.html (remember it's embedded, requires rebuild)

### Code Organization

- Database functions are in database.go, prefixed by operation (Get/Create/Update/Delete)
- HTTP handlers are in handlers.go, named by route (e.g., `CardsHandler`, `ReviewHandler`)
- Keep main.go minimal - only setup and routing

### Testing Considerations

- CGO is required for go-sqlite3, so tests need CGO_ENABLED=1
- Database tests should use in-memory SQLite (`:memory:`) or temp files
- Consider testing SRS algorithm separately from database operations
