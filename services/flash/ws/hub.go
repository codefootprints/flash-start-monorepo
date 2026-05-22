package ws

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Client merepresentasikan satu koneksi browser aktif
type Client struct {
	Conn *websocket.Conn
}

// Hub mengelola semua koneksi client WebSocket secara thread-safe
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.Mutex
}

// HubGlobal adalah instance tunggal yang akan diakses oleh controller dan main
var HubGlobal *Hub

// NewHub menginisialisasi Hub baru
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run menjalankan background loop untuk mengelola registrasi dan broadcast event
func (h *Hub) Run() {
	fmt.Println("🌐 [WEBSOCKET] Hub Manager Engine Berhasil Dijalankan!")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Printf("🔌 Client Terkoneksi! Total Client Aktif: %d\n", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mu.Unlock()
			fmt.Printf("❌ Client Terputus! Sisa Client Aktif: %d\n", len(h.clients))

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				// Kirim pesan ke setiap browser dalam goroutine terpisah agar tidak memblokir client lain
				go func(c *Client, msg []byte) {
					err := c.Conn.WriteMessage(websocket.TextMessage, msg)
					if err != nil {
						fmt.Println("⚠️ Gagal mengirim pesan ke client, memutuskan koneksi...")
						h.unregister <- c
					}
				}(client, message)
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastObject adalah fungsi pembantu untuk melempar struct apa saja ke seluruh client dalam bentuk JSON
func (h *Hub) BroadcastObject(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("⚠️ Gagal mengonversi objek ke JSON untuk broadcast:", err)
		return
	}
	h.broadcast <- data
}

// RegisterClient mendaftarkan client secara eksternal ke dalam channel
func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

// UnregisterClient mengeluarkan client dari channel
func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}
