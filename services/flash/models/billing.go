package models

import (
	"time"

	"gorm.io/gorm"
)

// Product merepresentasikan data master barang dagangan
type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(150);not null" json:"name"`
	Price     float64        `gorm:"type:numeric(15,2);not null" json:"price"`
	Stock     int            `gorm:"not null;default:0" json:"stock"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Transaction merepresentasikan rangkuman nota pembayaran (Header)
type Transaction struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	InvoiceNumber string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"invoice_number"`
	TotalAmount   float64        `gorm:"type:numeric(15,2);not null" json:"total_amount"`
	CashierName   string         `gorm:"type:varchar(100);not null" json:"cashier_name"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi: Satu Transaksi bisa memiliki banyak detail Order Item
	Orders []Order `gorm:"constraint:OnDelete:CASCADE;" json:"orders,omitempty"`
}

// Order merepresentasikan baris item detail yang dibeli di dalam satu transaksi
type Order struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `gorm:"not null" json:"transaction_id"` // Foreign Key ke Transaction
	ProductID     uint    `gorm:"not null" json:"product_id"`     // Foreign Key ke Product
	Quantity      int     `gorm:"not null" json:"quantity"`
	SubTotal      float64 `gorm:"type:numeric(15,2);not null" json:"sub_total"`

	// Eager Loading structs (Opsional saat query data)
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
