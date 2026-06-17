package repository

import (
	"task-tracker-backend/internal/domain"
)

type TaskRepository interface {
	GetAll(page, perPage int) (domain.PaginatedTasks, error)
	GetByID(id string) (*domain.Task, error)
	Create(task *domain.Task) error
	UpdateStatus(id string, status domain.TaskStatus) (*domain.Task, error)
	Delete(id string) error
}
