# START (Simple Task & Resource Tracker) 🚀

**START** adalah aplikasi Fullstack yang dirancang untuk membantu tim kecil mendigitalisasi operasional mereka. Aplikasi ini memungkinkan pengguna untuk mendata aset (resource), mengelola pengguna, dan menugaskan aset tersebut ke dalam tugas (task) tertentu dengan pelacakan status secara real-time.

Aplikasi ini dibangun dengan fokus pada performa tinggi menggunakan **Golang** di backend dan antarmuka yang modern menggunakan **React + Tailwind CSS v4**.

## 🛠️ Tech Stack

- **Backend:** Golang 1.21, Fiber Framework, GORM (ORM).
- **Database:** PostgreSQL 15.
- **Frontend:** React (Vite), Tailwind CSS v4.
- **Infrastructure:** Docker & Docker Compose.

## ✨ Fitur Utama

- **Resource Management:** CRUD data aset perusahaan (hardware, software, akun).
- **Automated Status Logic:** Status aset otomatis berubah menjadi `in_use` saat ditugaskan ke sebuah task, dan divalidasi agar tidak terjadi duplikasi peminjaman.
- **Database Transactions:** Menjamin integritas data antara pembuatan task dan pembaruan status aset.
- **Responsive Dashboard:** Tampilan bersih dengan tema warna bumi (Earth Tones) yang nyaman digunakan untuk waktu lama.

## 🚀 Cara Menjalankan Aplikasi

Pastikan Anda sudah menginstal [Docker](https://www.docker.com/) di mesin Anda.

1. Clone repositori ini:
   ```bash
   git clone [https://github.com/codefootprints/start-app.git](https://github.com/codefootprints/start-app.git)
   cd start-app