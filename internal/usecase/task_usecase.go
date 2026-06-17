package usecase

import "task-tracker-backend/internal/domain"

type TaskUseCase interface {
	GetTasks(page, perPage int) (domain.PaginatedTasks, error)
	GetTaskByID(id string) (*domain.Task, error)
	CreateTask(req domain.CreateTaskRequest) (*domain.Task, error)
	UpdateTaskStatus(id string, req domain.UpdateTaskStatusRequest) (*domain.Task, error)
}
