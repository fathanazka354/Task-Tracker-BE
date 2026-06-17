# Task Tracker API (Backend)

A RESTful API built with **Golang** for the Task Tracker application, demonstrating clean architecture, RESTful API design, and Docker containerization.

## Architecture

The backend follows **Clean Architecture** with clear separation of concerns:

| Layer | Responsibility |
|-------|---------------|
| `domain` | Core entities (`Task`), request/response types, status constants |
| `repository` | Data access interface + SQLite implementation |
| `usecase` | Business rules, orchestration, UUID generation |
| `handler` | HTTP request parsing, validation, routing |
| `pkg/response` | Uniform JSON response format |

Dependency flow: `handler` → `usecase` → `repository` → SQLite

## Tech Stack

- **Language**: Go 1.21
- **Framework**: Gin (HTTP router)
- **Database**: SQLite (via `go-sqlite3`)
- **UUID**: `google/uuid`
- **CORS**: `gin-contrib/cors`

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/v1/tasks` | List tasks (paginated) |
| `GET` | `/api/v1/tasks/:id` | Get task by ID |
| `POST` | `/api/v1/tasks` | Create new task |
| `PATCH` | `/api/v1/tasks/:id/status` | Update task status |

### Query Parameters (GET /api/v1/tasks)
- `page` (default: 1)
- `per_page` (default: 10, max: 50)

### Request Body Examples

**POST /api/v1/tasks**
```json
{
  "title": "Buy groceries",
  "description": "Milk, eggs, and bread from the store"
}
```

**PATCH /api/v1/tasks/:id/status**
```json
{
  "status": "done"
}
```

### Response Format
All responses follow a unified envelope:
```json
{
  "success": true,
  "message": "Tasks retrieved successfully",
  "data": { ... }
}
```

## Running the Backend

### Option 1: Docker Compose (Recommended)

```bash
# Start the server
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

The API will be available at `http://localhost:8080`.

### Option 2: Manual (requires Go 1.21+ and gcc for CGO)

```bash
# Copy environment file
cp .env.example .env

# Download dependencies
go mod tidy

# Run
go run ./cmd/main.go
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./data/tasks.db` | SQLite database file path |
