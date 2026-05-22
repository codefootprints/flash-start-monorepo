package main

import (
	"log"
	"start-app/database"
	"start-app/handlers"
	"start-app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// Inisialisasi Database dari package database
	db := database.InitDB()

	// Menjalankan AutoMigrate untuk membuat tabel otomatis
	err := db.AutoMigrate(&models.User{}, &models.Resource{}, &models.Task{})
	if err != nil {
		log.Fatal("Gagal melakukan migrasi database:", err)
	}

	// Inisialisasi Fiber
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // URL Vite
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	api := app.Group("/api")

	// Inisialisasi Handler
	resourceHandler := &handlers.ResourceHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}
	taskHandler := &handlers.TaskHandler{DB: db}

	// Resource Routes
	api.Get("/resources", resourceHandler.GetAllResources)
	api.Post("/resources", resourceHandler.CreateResource)
	api.Get("/resources/:id", resourceHandler.GetResourceByID)
	api.Delete("/resources/:id", resourceHandler.DeleteResource)

	// User Routes
	api.Get("/users", userHandler.GetAllUsers)
	api.Post("/users", userHandler.CreateUser)
	api.Get("/users/:id", userHandler.GetUserByID)

	// Handler Routes
	api.Get("/tasks", taskHandler.GetAllTasks)
	api.Get("/tasks/history", taskHandler.GetTaskHistory)
	api.Post("/tasks", taskHandler.CreateTask)
	api.Patch("tasks/:id/complete", taskHandler.CompleteTask)

	// Jalankan Server
	app.Listen(":3000")
}
