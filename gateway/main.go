package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/golang-jwt/jwt/v5"
	gorilla "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// MIDDLEWARE SECURITY GUARD (Task 10.1)
func AuthGuard() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil header Authorization (Format standar: Bearer <token>)
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Akses ditolak, token otentikasi tidak ditemukan",
			})
		}

		// Potong teks "Bearer " untuk mengambil string token murninya
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // Jika tidak ada kata Bearer
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Format token salah, harus menggunakan format 'Bearer <token>'",
			})
		}

		// Validasi tanda tangan token menggunakan JWT_SECRET
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Pastikan metode signing yang digunakan adalah HMAC (HS256)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing tidak didukung: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Jika token rusak, kedaluwarsa, atau tanda tangan tidak cocok
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Sesi Anda telah berakhir atau token anda tidak valid, silakan login kembali",
			})
		}

		// Jika token lolos verifikasi, izinkan request melaku ke handler proxy berikutnya
		return c.Next()
	}
}

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

	// 🚗 RUTE PROXY START SERVICE - JALUR BEBAS (Tanpa Login)
	// Membuka akses untuk register dan login agar user bisa mendapatkan token awal
	app.Post("/api/v1/start/register", func(c *fiber.Ctx) error {
		return proxy.Do(c, startURL+"/register")
	})
	app.Post("/api/v1/start/login", func(c *fiber.Ctx) error {
		return proxy.Do(c, startURL+"/login")
	})

	// 🔒 RUTE PROXY START SERVICE - JALUR AMAN (Diproteksi AuthGuard)
	// Semua rute selain login/register harus membawa token valid melewati AuthGuard()
	app.All("/api/v1/start/*", AuthGuard(), func(c *fiber.Ctx) error {
		remainingPath := strings.TrimPrefix(c.Path(), "/api/v1/start")
		targetURL := startURL + remainingPath

		// Ikut sertakan query parameter (?page=1&limit=10) jika ada
		if len(c.Request().URI().QueryString()) > 0 {
			targetURL += "?" + string(c.Request().URI().QueryString())
		}
		return proxy.Do(c, targetURL)
	})

	// ⚡️ RUTE PROXY FLASH SERVICE (HTTP REST)
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
