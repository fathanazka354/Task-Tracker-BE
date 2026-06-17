# Task Tracker — Golang Backend

REST API untuk Task Tracker App, dibangun dengan **Golang**, **Gin**, **SQLite**, dan **Docker**.

---

## Daftar Isi

- [Cara Menjalankan Project](#cara-menjalankan-project)
- [Architecture Explanation](#architecture-explanation)
- [Alasan Memilih Approach Ini](#alasan-memilih-approach-ini)
- [API Endpoints](#api-endpoints)

---

## Cara Menjalankan Project

### Prasyarat

| Tool | Keterangan |
|------|-----------|
| Docker + Docker Compose | Untuk menjalankan via container (direkomendasikan) |
| Go 1.21+ | Untuk menjalankan secara manual |
| gcc / build-essential | Dibutuhkan karena SQLite driver menggunakan CGO |

### Opsi 1: Docker Compose (Direkomendasikan)

Tidak perlu install Go atau gcc. Cukup Docker.

```bash
# Masuk ke folder backend
cd task-tracker-app/backend

# Jalankan di background
docker-compose up -d

# Lihat status container
docker-compose ps

# Lihat log real-time
docker-compose logs -f

# Hentikan container
docker-compose down

# Hentikan dan hapus volume (data hilang)
docker-compose down -v
```

Server berjalan di `http://localhost:8080`.

Data SQLite tersimpan di Docker volume bernama `task_data` — data tidak hilang saat container di-restart.

#### Verifikasi Backend Berjalan

```bash
curl http://localhost:8080/health
# Response: {"status":"ok","service":"task-tracker-backend"}
```

### Opsi 2: Manual (Butuh Go 1.21+ dan gcc)

```bash
cd task-tracker-app/backend

# Salin konfigurasi environment
cp .env.example .env

# Download dependencies
go mod tidy

# Jalankan server
go run ./cmd/main.go
```

### Environment Variables

| Variable | Default | Keterangan |
|----------|---------|------------|
| `PORT` | `8080` | Port HTTP server |
| `DB_PATH` | `./data/tasks.db` | Path file database SQLite |

File database dibuat otomatis saat server pertama kali dijalankan. Direktori `./data/` juga dibuat otomatis.

---

## Architecture Explanation

Backend mengikuti **Clean Architecture** dengan dependency rule yang sama dengan frontend: layer luar boleh bergantung ke layer dalam, tidak sebaliknya.

```
┌──────────────────────────────────────┐
│             HANDLER                  │
│  Parse request │ Validasi input      │
│  Format response │ Route HTTP        │
├──────────────────────────────────────┤
│             USECASE                  │
│  Business rules │ UUID generation    │
│  Bounds check │ Orchestrasi         │
├──────────────────────────────────────┤
│           REPOSITORY                 │
│  Interface (domain) │ SQLite impl    │
│  SQL queries │ Pagination            │
├──────────────────────────────────────┤
│             DOMAIN                   │
│  Task struct │ TaskStatus constants  │
│  Request/Response types              │
└──────────────────────────────────────┘
```

Dependency flow: `main.go` → `handler` → `usecase` → `repository` → SQLite

### Domain (`internal/domain/task.go`)

Definisi core — tidak bergantung ke package lain:

```go
type Task struct {
    ID          string     `json:"id"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Status      TaskStatus `json:"status"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskStatus string
const (
    StatusPending TaskStatus = "pending"
    StatusDone    TaskStatus = "done"
)
```

### Repository (`internal/repository/`)

Dua file:
- `task_repository.go` — interface (kontrak)
- `sqlite_task_repository.go` — implementasi SQLite

Interface memastikan use case tidak tahu implementasi storage yang dipakai:

```go
type TaskRepository interface {
    GetAll(page, perPage int) (domain.PaginatedTasks, error)
    GetByID(id string) (*domain.Task, error)
    Create(task *domain.Task) error
    UpdateStatus(id string, status domain.TaskStatus) (*domain.Task, error)
    Delete(id string) error
}
```

SQLite migration dijalankan otomatis saat repository diinisialisasi — tidak perlu setup database manual.

### Use Case (`internal/usecase/`)

Business logic yang tidak terkait HTTP maupun storage:
- Validasi batas `page` dan `perPage`
- Generate UUID untuk task baru
- Set timestamp `created_at` dan `updated_at`

### Handler (`internal/handler/`)

- `task_handler.go` — parse request, panggil use case, format response
- `router.go` — setup route Gin + CORS middleware

### Response Helper (`pkg/response/`)

Format JSON yang seragam di seluruh API:

```go
// Sukses
response.OK(c, "Tasks retrieved successfully", data)
response.Created(c, "Task created successfully", task)

// Error
response.BadRequest(c, "Validation failed", err.Error())
response.NotFound(c, "Task not found")
response.InternalError(c, "Failed to fetch tasks")
```

Setiap response memiliki struktur:
```json
{
  "success": true | false,
  "message": "...",
  "data": { ... },       // ada jika sukses
  "error": "..."         // ada jika gagal
}
```

### Wiring di `cmd/main.go`

Dependency injection dilakukan secara manual — tidak menggunakan framework DI:

```go
repo, _ := repository.NewSQLiteTaskRepository(dbPath)
taskUseCase := usecase.NewTaskUseCase(repo)
taskHandler := handler.NewTaskHandler(taskUseCase)
router := handler.SetupRouter(taskHandler)
router.Run(":" + port)
```

Ini deliberate choice — transparansi lebih penting daripada otomatisasi DI untuk project skala ini.

### Docker Implementation

**Multi-stage Dockerfile:**

```
Stage 1 (builder): golang:alpine + gcc + semua dependencies
  → kompilasi binary dengan CGO_ENABLED=1

Stage 2 (runtime): alpine + sqlite-libs
  → hanya berisi binary final (~20MB total image)
```

**docker-compose.yml:**
- Volume persistent `task_data` untuk database SQLite
- Healthcheck setiap 30 detik ke `/health`
- Restart policy `unless-stopped`

---

## Alasan Memilih Approach Ini

### Mengapa Golang?

- **Performa**: Kompilasi native, overhead sangat rendah untuk REST API sederhana
- **Single binary**: Hasil kompilasi adalah satu file eksekutabel — mudah di-Dockerize
- **Strongly typed**: Lebih mudah di-maintain dibandingkan bahasa dinamis
- **Ekosistem**: Gin mature dan terdokumentasi dengan baik

### Mengapa SQLite, Bukan PostgreSQL?

Scope backend adalah "sederhana, fokus pada kualitas Flutter". SQLite:
- Zero infrastructure — tidak perlu service database terpisah
- Data dalam satu file, mudah di-backup
- Docker volume membuatnya persistent tanpa konfigurasi tambahan

Jika app berkembang ke production multi-user, migrasi ke PostgreSQL hanya perlu menulis ulang `sqlite_task_repository.go` — use case dan handler tidak perlu diubah karena repository interface tetap sama.

### Mengapa Manual DI di `main.go`?

Framework DI seperti `wire` atau `dig` menambah kompleksitas dan indirection yang tidak diperlukan di project ini. Manual wiring di `main.go` lebih transparan — mudah dibaca dan dipahami tanpa harus memahami cara kerja framework DI.

### Mengapa Gin?

Gin adalah salah satu HTTP framework Go paling mature dengan:
- Performa tinggi (router berbasis radix tree)
- Binding + validasi request yang built-in
- Middleware ecosystem yang lengkap (CORS, logger, recovery)
- Dokumentasi dan komunitas besar

---

## API Endpoints

### Health Check

```
GET /health
```
Response:
```json
{"status": "ok", "service": "task-tracker-backend"}
```

### Task Endpoints

| Method | Path | Deskripsi |
|--------|------|-----------|
| `GET` | `/api/v1/tasks` | Ambil daftar task (paginated) |
| `GET` | `/api/v1/tasks/:id` | Ambil task berdasarkan ID |
| `POST` | `/api/v1/tasks` | Buat task baru |
| `PATCH` | `/api/v1/tasks/:id/status` | Update status task |

### GET /api/v1/tasks

Query parameters:

| Parameter | Default | Keterangan |
|-----------|---------|-----------|
| `page` | `1` | Nomor halaman |
| `per_page` | `10` | Item per halaman (max 50) |

Contoh: `GET /api/v1/tasks?page=2&per_page=5`

Response:
```json
{
  "success": true,
  "message": "Tasks retrieved successfully",
  "data": {
    "data": [
      {
        "id": "uuid-xxx",
        "title": "Setup project",
        "description": "Initialize the repository",
        "status": "pending",
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 25,
    "page": 2,
    "per_page": 5,
    "total_pages": 5
  }
}
```

### POST /api/v1/tasks

Request body:
```json
{
  "title": "Judul task (min 3, max 100 karakter)",
  "description": "Deskripsi task (min 5 karakter)"
}
```

Response `201 Created`:
```json
{
  "success": true,
  "message": "Task created successfully",
  "data": { ... task object ... }
}
```

Validasi error `400 Bad Request`:
```json
{
  "success": false,
  "message": "Validation failed",
  "error": "Key: 'CreateTaskRequest.Title' Error:Field validation for 'Title' failed on the 'min' tag"
}
```

### PATCH /api/v1/tasks/:id/status

Request body:
```json
{
  "status": "done"
}
```

Nilai `status` yang valid: `"pending"` atau `"done"`

Response `200 OK`:
```json
{
  "success": true,
  "message": "Task status updated successfully",
  "data": { ... task object dengan status baru ... }
}
```

### GET /api/v1/tasks/:id

Response `200 OK`:
```json
{
  "success": true,
  "message": "Task retrieved successfully",
  "data": { ... task object ... }
}
```

Response `404 Not Found`:
```json
{
  "success": false,
  "message": "Task not found"
}
```
