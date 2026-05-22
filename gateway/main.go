package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Membaca file .env jika tersedia
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ File .env tidak ditemukan, menggunakan variable lingkungan sistem")
	}

	// Mengambil port gateway (Default: 8000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New(fiber.Config{
		AppName: "FLASH-START Consilidated API Gateway v1.0",
	})

	// Endpoint Dasar Cek Kesehatan Server Gateway (Task 9.2)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "API Gateway terhubung dengan baik dan siap mengarahkan traffic",
		})
	})

	log.Printf("🚀 API Gateway berhasil mengudara di port: %s", port)
	log.Fatal(app.Listen(":" + port))
}
