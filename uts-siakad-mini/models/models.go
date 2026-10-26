package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string         `gorm:"not null" json:"role"`
	Student   *Student       `gorm:"foreignKey:UserID" json:"student,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Student struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	NIM         string         `gorm:"uniqueIndex;type:varchar(12);not null" json:"nim"`
	Nama        string         `gorm:"not null" json:"nama"`
	Prodi       string         `gorm:"not null" json:"prodi"`
	Angkatan    int            `gorm:"not null" json:"angkatan"`
	IPKTerakhir float64        `gorm:"type:numeric(3,2)" json:"ipk_terakhir"`
	Enrollments []Enrollment   `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Course struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	KodeMK      string       `gorm:"uniqueIndex;not null" json:"kode_mk"`
	NamaMK      string       `gorm:"not null" json:"nama_mk"`
	SKS         int          `gorm:"not null" json:"sks"`
	Semester    int          `gorm:"not null" json:"semester"`
	Kuota       int          `gorm:"not null" json:"kuota"`
	Enrollments []Enrollment `gorm:"foreignKey:CourseID" json:"enrollments,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Enrollment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentID     uint      `gorm:"uniqueIndex:idx_student_course_year;not null" json:"student_id"`
	CourseID      uint      `gorm:"uniqueIndex:idx_student_course_year;not null" json:"course_id"`
	TahunAkademik string    `gorm:"uniqueIndex:idx_student_course_year;not null" json:"tahun_akademik"`
	Student       *Student  `json:"student,omitempty"`
	Course        *Course   `json:"course,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}
