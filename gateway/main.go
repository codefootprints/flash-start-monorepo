package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	gorilla "github.com/gorilla/websocket"
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

	// Mengambil alamat target microservices dari .env
	startURL := os.Getenv("START_SERVICE_URL")
	flashURL := os.Getenv("FLASH_SERVICE_URL")

	app := fiber.New(fiber.Config{
		AppName: "FLASH-START Consilidated API Gateway v1.0",
	})

	// Pasang CORS di level Gateway
	// agar frontend (:5173) cukup mengetuk pintu port :8000 tanpa terkena blokir browser
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// Endpoint Dasar Cek Kesehatan Server Gateway (Task 9.2)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "API Gateway terhubung dengan baik dan siap mengarahkan traffic",
		})
	})

	// Reverse Proxy route untuk websocket live stream (FLASH)
	// harus diletakkan di atas rute HTTP biasa agar rute jabat tangan (Handshake Upgrade) terdeteksi duluan
	app.Get("/api/v1/flash/ws/notifications", websocket.New(func(c *websocket.Conn) {
		// Mengubah protokol target http://localhost:3000 menjadi ws://localhost:3000/ws/notifications
		targetWS := strings.Replace(flashURL, "http", "ws", 1) + "/ws/notifications"

		// Membuka pipa koneksi internal dari Gateway ke Backend FLASH
		backendConn, _, err := gorilla.DefaultDialer.Dial(targetWS, nil)
		if err != nil {
			log.Println("⚠️ Gateway gagal menjangkau WebSocket Backend FLASH", err)
			return
		}
		defer backendConn.Close()

		errChan := make(chan error, 2)

		// Sub-Worker A: Mengalirkan data secara real-time dari Backend FLASH -> Browser Client
		go func() {
			for {
				msgType, msg, err := backendConn.ReadMessage()
				if err != nil {
					errChan <- err
					return
				}
				if err := c.WriteMessage(msgType, msg); err != nil {
					errChan <- err
					return
				}
			}
		}()

		// Sub-Worker B: Mengalirkan data dari Browser Client -> Backend FLASH (untuk sinyal Ping/Keep-Alive)
		go func() {
			for {
				msgType, msg, err := c.ReadMessage()
				if err != nil {
					errChan <- err
					return
				}
				if err := backendConn.WriteMessage(msgType, msg); err != nil {
					errChan <- err
					return
				}
			}
		}()

		// Menahan goroutine tetap hidup selama pipa komunikasi berjalan
		<-errChan
		log.Println("🔌 Jalur koneksi WebSocket Proxy ditutup dengan aman")
	}))

	// 🚗 Reverse Proxy Route untuk START Service (HTTP REST)
	// mengarahkan /api/v1/start/* ke http://localhost:3001/*
	app.All("/api/v1/start/*", func(c *fiber.Ctx) error {
		remainingPath := strings.TrimPrefix(c.Path(), "/api/v1/start")
		targetURL := startURL + remainingPath

		// Ikut sertakan query parameter (?page=1&limit=10) jika ada
		if len(c.Request().URI().QueryString()) > 0 {
			targetURL += "?" + string(c.Request().URI().QueryString())
		}
		return proxy.Do(c, targetURL)
	})

	// 🚗 Reverse Proxy Route untuk FLASH Service (HTTP REST)
	// mengarahkan /api/v1/flash/* ke http://localhost:3000/*
	app.All("/api/v1/flash/*", func(c *fiber.Ctx) error {
		remainingPath := strings.TrimPrefix(c.Path(), "/api/v1/flash")
		targetURL := flashURL + remainingPath

		// Ikut sertakan query parameter (?page=1&limit=10) jika ada
		if len(c.Request().URI().QueryString()) > 0 {
			targetURL += "?" + string(c.Request().URI().QueryString())
		}
		return proxy.Do(c, targetURL)
	})

	log.Printf("🚀 API Gateway berhasil mengudara di port: %s", port)

	// Untuk mengintip isi perut router Fiber
	for _, route := range app.Stack() {
		for _, r := range route {
			log.Printf("Route Terdaftar: [%s] -> %s", r.Method, r.Path)
		}
	}

	log.Fatal(app.Listen(":" + port))
}
