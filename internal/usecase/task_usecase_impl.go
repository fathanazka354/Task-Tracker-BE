package usecase

import (
	"task-tracker-backend/internal/domain"
	"task-tracker-backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

type taskUseCaseImpl struct {
	repo repository.TaskRepository
}

func NewTaskUseCase(repo repository.TaskRepository) TaskUseCase {
	return &taskUseCaseImpl{repo: repo}
}

func (u *taskUseCaseImpl) GetTasks(page, perPage int) (domain.PaginatedTasks, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return u.repo.GetAll(page, perPage)
}

func (u *taskUseCaseImpl) GetTaskByID(id string) (*domain.Task, error) {
	return u.repo.GetByID(id)
}

func (u *taskUseCaseImpl) CreateTask(req domain.CreateTaskRequest) (*domain.Task, error) {
	now := time.Now()
	task := &domain.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := u.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *taskUseCaseImpl) UpdateTaskStatus(id string, req domain.UpdateTaskStatusRequest) (*domain.Task, error) {
	return u.repo.UpdateStatus(id, req.Status)
}
