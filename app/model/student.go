package model

import "time"

// Student merepresentasikan entitas mahasiswa di sistem.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     string    `json:"grade"`
	IsActive bool      `json:"is_active"`
	OwnerID  int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

// POST — semua field wajib diisi saat pembuatan data baru
type CreateStudentRequest struct {
	NIM   string `json:"nim" validate:"required,nim"`
	Name  string `json:"name" validate:"required,min=3,max=100"`
	Grade string `json:"grade" validate:"required,min=1,max=2"`
}

// PUT — mengganti seluruh isi, jadi semua field wajib
type ReplaceStudentRequest struct {
	NIM      string `json:"nim" validate:"required,nim"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Grade    string `json:"grade" validate:"required,min=1,max=2"`
	IsActive bool   `json:"is_active"`
}

// PATCH — ubah sebagian, field bertipe pointer agar bisa membedakan "tidak dikirim" (nil) dan "kosong"
type PatchStudentRequest struct {
	NIM      *string `json:"nim,omitempty" validate:"omitnil,nim"`
	Name     *string `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
	Grade    *string `json:"grade,omitempty" validate:"omitnil,min=1,max=2"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// WebResponse adalah amplop baku untuk semua respons REST API
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
