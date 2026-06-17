package main

import (
	"log"
	"os"
	"task-tracker-backend/internal/handler"
	"task-tracker-backend/internal/repository"
	"task-tracker-backend/internal/usecase"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/tasks.db"
	}

	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	repo, err := repository.NewSQLiteTaskRepository(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	taskUseCase := usecase.NewTaskUseCase(repo)
	taskHandler := handler.NewTaskHandler(taskUseCase)

	router := handler.SetupRouter(taskHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
