package main

import (
	"flash-app/controllers"
	"flash-app/database"
	"flash-app/models"
	"flash-app/ws"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Muat file .env jika berjalan di luar Docker (lokal)
	if os.Getenv("DB_HOST") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel lingkungan sistem.")
		}
	}

	fmt.Println("🚀 Memulai Inisialisasi Engine FLASH...")

	// Jalankan Koneksi Database & Redis
	db := database.InitPostgres()
	database.InitRedis()

	// Jalankan Auto-Migration untuk skema tabel FLASH (Task 2.1)
	fmt.Println("⏳ Menjalankan Auto-Migration Struktur Tabel...")
	err := db.AutoMigrate(&models.Product{}, &models.Transaction{}, &models.Order{})
	if err != nil {
		log.Fatal("❌ Gagal melakukan migrasi database:", err)
	}
	fmt.Println("✅ Database Schema Berhasil Dimigrasikan!")

	// Inisialisasi dan jalankan WebSocket Hub di background (Goroutine)
	ws.HubGlobal = ws.NewHub()
	go ws.HubGlobal.Run()

	// 🚀 Pemicu Background Worker Redis Pub/Sub (Task 4.2)
	go listenRedisPubSub()

	// Inisialisasi Fiber App
	app := fiber.New(fiber.Config{
		AppName: "FLASH - Realtime Billing Engine v1.0",
	})

	// 🚀 PASANG MIDDLEWARE CORS DI SINI (Task Pengaman API)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // Mengizinkan alamat frontend React Anda
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// Route cek kesehatan server dasar
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"message": "FLASH Backend is running smoothly",
		})
	})

	// Billing REST Route (Task 2.2)
	app.Post("/api/transactions", controllers.CreateTransaction)
	app.Get("/api/transactions", controllers.GetTransactions)

	// Middleware Upgrade HTTP ke WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Endpoint WebSocket Route (Task 3.1 & 3.2)
	app.Get("/ws/notifications", websocket.New(func(c *websocket.Conn) {
		client := &ws.Client{Conn: c}

		// Daftarkan client ke pusat Hub
		ws.HubGlobal.RegisterClient(client)

		// Bersihkan koneksi saat browser ditutup atau disconnect
		defer func() {
			ws.HubGlobal.UnregisterClient(client)
		}()

		// Read loop: Menjaga koneksi tetap hidup dan mendengarkan jika ada pesan masuk dari client
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				// Keluar dari loop jika koneksi terputus
				break
			}

			// Jika client iseng mengirim teks, kita print di log server (tidak wajib)
			if messageType == websocket.TextMessage {
				fmt.Printf("📩 Pesan masuk dari browser client: %s\n", string(message))
			}
		}
	}))

	// Jalankan Server HTTP
	log.Fatal(app.Listen(":3000"))
}

// listenRedisPubSub berjalan di background untuk menangkap pesan dari Redis 8 dan meneruskannya ke WebSocket Hub
func listenRedisPubSub() {
	// Membuat subscriber ke channel bernama "flash_billing_notifications"
	pubsub := database.Rdb.Subscribe(database.Ctx, "flash_billing_notifications")
	defer pubsub.Close()

	fmt.Println("📢 [REDIS] Background Worker Subscribed ke Channel 'flash_billing_notifications'!")

	// Mendengarkan pesan masuk secara terus-menerus (blocking loop)
	for {
		msg, err := pubsub.ReceiveMessage(database.Ctx)
		if err != nil {
			fmt.Println("⚠️ Redis Pub/Sub mengalami error saat menerima pesan:", err)
			time.Sleep(2 * time.Second) // Delay sejenak sebelum mencoba ulang jika koneksi terputus
			continue
		}

		// Ketika ada pesan masuk dari Redis, langsung tembak ke channel broadcast milik WebSocket Hub!
		fmt.Printf("🎯 [REDIS PUB/SUB] Menangkap event transaksi baru! Meneruskan ke WebSocket...\n")
		ws.HubGlobal.BroadcastObject(map[string]interface{}{
			"event":   "NEW_TRANSACTION",
			"payload": msg.Payload, // Isi pesan dalam bentuk string JSON transaksi
		})
	}
}
