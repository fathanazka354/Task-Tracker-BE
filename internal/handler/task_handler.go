package handler

import (
	"strconv"
	"task-tracker-backend/internal/domain"
	"task-tracker-backend/internal/usecase"
	"task-tracker-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	useCase usecase.TaskUseCase
}

func NewTaskHandler(useCase usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{useCase: useCase}
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	result, err := h.useCase.GetTasks(page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch tasks")
		return
	}
	response.OK(c, "Tasks retrieved successfully", result)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	task, err := h.useCase.GetTaskByID(id)
	if err != nil {
		response.InternalError(c, "Failed to fetch task")
		return
	}
	if task == nil {
		response.NotFound(c, "Task not found")
		return
	}
	response.OK(c, "Task retrieved successfully", task)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req domain.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	task, err := h.useCase.CreateTask(req)
	if err != nil {
		response.InternalError(c, "Failed to create task")
		return
	}
	response.Created(c, "Task created successfully", task)
}

func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	task, err := h.useCase.UpdateTaskStatus(id, req)
	if err != nil {
		if err.Error() == "task not found" {
			response.NotFound(c, "Task not found")
			return
		}
		response.InternalError(c, "Failed to update task")
		return
	}
	response.OK(c, "Task status updated successfully", task)
}
