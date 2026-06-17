package repository

import (
	"database/sql"
	"fmt"
	"math"
	"task-tracker-backend/internal/domain"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(dbPath string) (*SQLiteTaskRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteTaskRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteTaskRepository) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS tasks (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL
    );`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteTaskRepository) GetAll(page, perPage int) (domain.PaginatedTasks, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		return domain.PaginatedTasks{}, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		"SELECT id, title, description, status, created_at, updated_at FROM tasks ORDER BY created_at DESC LIMIT ? OFFSET ?",
		perPage, offset,
	)
	if err != nil {
		return domain.PaginatedTasks{}, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		var createdAt, updatedAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &createdAt, &updatedAt); err != nil {
			return domain.PaginatedTasks{}, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []domain.Task{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return domain.PaginatedTasks{
		Data:       tasks,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *SQLiteTaskRepository) GetByID(id string) (*domain.Task, error) {
	var t domain.Task
	var createdAt, updatedAt string
	err := r.db.QueryRow(
		"SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = ?", id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &t, nil
}

func (r *SQLiteTaskRepository) Create(task *domain.Task) error {
	_, err := r.db.Exec(
		"INSERT INTO tasks (id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		task.ID, task.Title, task.Description, task.Status,
		task.CreatedAt.Format(time.RFC3339),
		task.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteTaskRepository) UpdateStatus(id string, status domain.TaskStatus) (*domain.Task, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.Exec(
		"UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?",
		status, now, id,
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("task not found")
	}
	return r.GetByID(id)
}

func (r *SQLiteTaskRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}
