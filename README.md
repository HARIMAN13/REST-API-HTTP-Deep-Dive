# Praktikum Pemrograman Backend Lanjut (PBL)

Repositori ini memuat seluruh kode implementasi praktikum dan tugas mandiri mata kuliah **Praktikum Pemrograman Backend Lanjut (2 SKS) - Semester 5**, Program Studi **D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga**.

Seluruh materi perkuliahan telah ditata secara modular ke dalam direktori masing-masing per pertemuan/topik praktikum agar rapi, terisolasi, dan mudah ditinjau.

---

## 📁 Daftar Modul & Struktur Proyek

| Folder | Topik Pertemuan | Deskripsi & Fitur Utama |
| :--- | :--- | :--- |
| [📁 `modul-01-persiapan-go-fiber/`](./modul-01-persiapan-go-fiber) | **Pertemuan 1: Persiapan & Sintaks Go Fiber** | Pengenalan sintaks Go (variabel, slice, map, pointer, struct, channel) & HTTP server Hello World dengan Fiber v2. |
| [📁 `modul-02-rest-api/`](./modul-02-rest-api) | **Pertemuan 2: REST API & HTTP Deep Dive** | Standarisasi RESTful API, pembedaan PUT vs PATCH, standard response envelope (`WebResponse`), parser query & pagination in-memory. |
| [📁 `modul-03-database-repository/`](./modul-03-database-repository) | **Pertemuan 3: Database & Repository Pattern** | Integrasi PostgreSQL menggunakan connection pool (`pgxpool`), skema migrasi tabel `students`, dan abstraksi lapisan database via Repository Pattern. |
| [📁 `modul-04-clean-architecture/`](./modul-04-clean-architecture) | **Pertemuan 4: Clean Architecture** | Restrukturisasi proyek ke arsitektur berlapis (Model, Repository, Service / Business Rules murni, Route, Middleware, Config) & unit test rules. |
| [📁 `modul-05-auth-security/`](./modul-05-auth-security) | **Pertemuan 5: Authentication & Security** | Hashing password dengan Bcrypt, otentikasi JWT (Access Token & Refresh Token rotasi), endpoint auth (`/login`, `/register`, `/refresh`, `/logout`, `/me`), dan auth middleware. |
| [📁 `modul-06-authorization-rbac/`](./modul-06-authorization-rbac) | **Pertemuan 6: Authorization & RBAC** | Role-Based Access Control (RBAC) relasional (`roles`, `permissions`, `role_permissions`), otorisasi berbasis hak akses dinamis, dan aturan kepemilikan data (data ownership). |
| [📁 `modul-07-advanced-api/`](./modul-07-advanced-api) | **Pertemuan 7: Advanced API Design** | Keyset / Cursor-based Pagination, validasi deklaratif via `go-playground/validator`, diferensiasi patch pointer, dan Content Negotiation (`Accept: application/json` / `text/csv`). |
| [📁 `uts-siakad-mini/`](./uts-siakad-mini) | **Pertemuan 8: UTS - Siakad Mini** | Implementasi sistem akademik mini: autentikasi JWT, manajemen mahasiswa, matakuliah, perwalian, dan pendaftaran KRS (Enrollment) dengan validasi batas SKS. |
| [📁 `modul-09-transactions-concurrency/`](./modul-09-transactions-concurrency) | **Pertemuan 9: Transactions & Concurrency** | Pengendalian transaksi ACID via `pgx`, mitigasi race condition (*check-then-act*), *atomic update*, *pessimistic locking*, fitur *supervision*, serta dokumentasi pengujian konkurensi. |
| [📁 `modul-10-api-mongodb/`](./modul-10-api-mongodb) | **Pertemuan 10: Polyglot Persistence (MongoDB)** | Integrasi basis data dokumen NoSQL (MongoDB) mendampingi PostgreSQL untuk menyimpan data portofolio dan review mahasiswa. |
| [📁 `modul-11-background-jobs/`](./modul-11-background-jobs) | **Pertemuan 11: Background Jobs & Async Processing** | Antrean tugas asinkron dengan Asynq & Redis, Transactional Outbox Pattern, notifikasi kuota matakuliah, outbox relay worker, dan mekanisme redeliver. |

---

## 🚀 Panduan Menjalankan Tiap Modul

Setiap modul dirancang independen dengan modul Go (`go.mod`) masing-masing.

### 1. Masuk ke Folder Modul
Pilih modul yang ingin diuji, misalnya Modul 4:
```bash
cd modul-04-clean-architecture
```

### 2. Konfigurasi Environment (`.env`)
Salin file `.env.example` (jika tersedia pada modul terkait) menjadi `.env`, lalu sesuaikan kredensial basis data Anda:
```bash
cp .env.example .env
```

### 3. Migrasi Basis Data (Bila Menggunakan Database)
Jalankan berkas migrasi SQL yang terdapat pada subfolder `migrations/` ke database PostgreSQL lokal Anda:
```bash
psql -U postgres -d nama_db -f migrations/001_create_students.sql
```

### 4. Menjalankan Aplikasi
Unduh dependensi dan jalankan server:
```bash
go mod tidy
go run main.go
```

### 5. Menjalankan Pengujian Unit (*Unit Test*)
Untuk modul yang memuat unit test (seperti Modul 4, 5, 6, 7):
```bash
go test -v ./app/service/...
```

---

## 👤 Identitas Mahasiswa

- **Nama**: Rafi Hariman Priambodo
- **Program Studi**: D4 Teknik Informatika
- **Fakultas**: Fakultas Vokasi, Universitas Airlangga
- **GitHub**: [@HARIMAN13](https://github.com/HARIMAN13)
