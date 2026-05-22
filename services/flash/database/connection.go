package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Rdb *redis.Client
	Ctx = context.Background()
)

// InitPostgres mengonfigurasi koneksi ke PostgreSQL 18 dengan Connection Pool
func InitPostgres() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Menggunakan logger silent/error saja agar log terminal tidak penuh saat production build
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal("❌ Gagal terhubung ke PostgreSQL 18:", err)
	}

	// Optimasi Connection Pool untuk performa tinggi
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)           // Jumlah koneksi idle maksimum
		sqlDB.SetMaxOpenConns(100)          // Jumlah koneksi terbuka maksimum
		sqlDB.SetConnMaxLifetime(time.Hour) // Batas waktu hidup koneksi
	}

	fmt.Println("✅ [POSTGRES] Sukses Terhubung ke PostgreSQL 18 (Pool Optimized)!")
	DB = db
	return db
}

// InitRedis mengonfigurasi koneksi ke Redis 8 menggunakan go-redis v9
func InitRedis() *redis.Client {
	redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     "",               // Kosong secara default pada image alpine
		DB:           0,                // Default DB 0
		DialTimeout:  10 * time.Second, // Batas waktu koneksi awal
		ReadTimeout:  30 * time.Second, // Batas waktu pembacaan data
		WriteTimeout: 30 * time.Second, // Batas waktu penulisan data
	})

	// Melakukan uji ketuk (Ping) ke Redis Server dengan batas waktu context
	pingCtx, cancel := context.WithTimeout(Ctx, 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(pingCtx).Result()
	if err != nil {
		log.Fatal("❌ Gagal terhubung ke Redis 8 Broker:", err)
	}

	fmt.Println("✅ [REDIS] Sukses Terhubung ke Redis 8 Broker (Pub/Sub Engine Ready)!")
	Rdb = rdb
	return rdb
}
