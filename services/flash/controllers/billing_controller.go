package controllers

import (
	"context"
	"encoding/json"
	"flash-app/database"
	"flash-app/models"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Request payload untuk membuat transaksi baru
type CreateTransactionInput struct {
	CashierName string `json:"cashier_name" validate:"required"`
	Items       []struct {
		ProductID uint `json:"product_id" validate:"required"`
		Quantity  int  `json:"quantity" validate:"required,gt=0"`
	} `json:"items" validate:"required,dive"`
}

// CreateTransaction menangani pembuatan nota belanja dengan GORM Transaction
func CreateTransaction(c *fiber.Ctx) error {
	var input CreateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format input tidak valid"})
	}

	if input.CashierName == "" || len(input.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama kasir dan item belanja wajib diisi"})
	}

	// Memulai Database Transaction (ACID Compliance)
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var totalAmount float64
	var orderItems []models.Order

	// Loop untuk memvalidasi setiap produk dan menghitung subtotal
	for _, item := range input.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Produk dengan ID %d tidak ditemukan", item.ProductID),
			})
		}

		// Validasi kecukupan stok barang
		if product.Stock < item.Quantity {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Stok produk '%s' tidak mencukupi (Tersisa: %d)", product.Name, product.Stock),
			})
		}

		// Kurangi stok produk secara atomik
		product.Stock -= item.Quantity
		tx.Save(&product)

		// Hitung subtotal item belanja
		subTotal := product.Price * float64(item.Quantity)
		totalAmount += subTotal

		// Siapkan objek order untuk disimpan nanti
		orderItems = append(orderItems, models.Order{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			SubTotal:  subTotal,
		})
	}

	// Generate nomor invoice acak unik (Contoh: INV-20260518-8472)
	rand.Seed(time.Now().UnixNano())
	invoiceNum := fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))

	// Buat objek transaksi header utama
	transaction := models.Transaction{
		InvoiceNumber: invoiceNum,
		TotalAmount:   totalAmount,
		CashierName:   input.CashierName,
		Orders:        orderItems,
	}

	// Simpan transaksi beserta relasi detail orders-nya sekaligus
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan transaksi"})
	}

	// Commit transaksi jika semua proses berhasil tanpa cela
	tx.Commit()

	// 🚀 REFACTOR REAL-TIME STREAM DENGAN EAGER LOADING (Task 7.1)
	go func(transactionID uint) {
		// Buat variabel baru untuk menampung data transaksi yang sudah lengkap
		var completeTx models.Transaction

		// Ambil ulang data transaksi dari database, gabungkan dengan data detail order dan nama produknya
		err := database.DB.Preload("Orders.Product").First(&completeTx, transactionID).Error
		if err != nil {
			fmt.Println("⚠️ Gagal memuat data lengkap (Eager Loading) untuk Redis:", err)
			return
		}

		// Konversi objek transaksi yang sudah lengkap menjadi string JSON
		jsonBytes, err := json.Marshal(completeTx)
		if err != nil {
			fmt.Println("⚠️ Gagal mengonversi data transaksi lengkap ke JSON:", err)
			return
		}

		// Buat konteks singkat dengan batas waktu 5 detik untuk operasi Redis
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Publish string JSON lengkap tersebut ke channel "flash_billing_notifications"
		err = database.Rdb.Publish(redisCtx, "flash_billing_notifications", string(jsonBytes)).Err()
		if err != nil {
			fmt.Println("⚠️ Gagal mengirim pesan Publish ke Redis:", err)
		} else {
			fmt.Println("🚀 [REDIS] Sukses mempublikasikan event transaksi lengkap (Preloaded) ke channel!")
		}
	}(transaction.ID) // Kirim ID transaksi saja ke dalam Goroutine sebagai acuan query ulang

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Transaksi berhasil diproses",
		"transaction": transaction,
	})
}

// GetTransactions mengambil semua data transaksi beserta detail itemnya dari Postgres
func GetTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction

	// Menggunakan Preload("Orders.Product") agar data detail item dan nama produknya ikut terbawa (Eager Loading)
	err := database.DB.Preload("Orders.Product").Order("created_at desc").Find(&transactions).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data transaksi"})
	}

	return c.JSON(transactions)
}
