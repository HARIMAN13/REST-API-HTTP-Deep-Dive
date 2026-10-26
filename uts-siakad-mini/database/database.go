package database

import (
	"fmt"
	"log"
	"os"

	"siakad-mini/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}

	log.Println("Database connected")
}

func MigrateDB() {
	err := DB.AutoMigrate(&models.User{}, &models.Student{}, &models.Course{}, &models.Enrollment{})
	if err != nil {
		log.Fatal("Failed to migrate database. \n", err)
	}
	log.Println("Database migrated")
}

func SeedDB() {
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.User{
		Email:    "admin@siakad.com",
		Password: string(adminPassword),
		Role:     "admin",
	}
	DB.Create(&admin)

	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("18722100%04d", i)
		studentPassword, _ := bcrypt.GenerateFromPassword([]byte(nim), bcrypt.DefaultCost)
		user := models.User{
			Email:    fmt.Sprintf("student%d@siakad.com", i),
			Password: string(studentPassword),
			Role:     "mahasiswa",
		}
		DB.Create(&user)

		student := models.Student{
			UserID:      user.ID,
			NIM:         nim,
			Nama:        fmt.Sprintf("Student %d", i),
			Prodi:       "Sistem Informasi",
			Angkatan:    2022,
			IPKTerakhir: 3.00 + (float64(i%10) * 0.1),
		}
		DB.Create(&student)
	}

	courses := []models.Course{
		{KodeMK: "SI101", NamaMK: "Algoritma", SKS: 3, Semester: 1, Kuota: 30},
		{KodeMK: "SI102", NamaMK: "Basis Data", SKS: 3, Semester: 2, Kuota: 30},
		{KodeMK: "SI103", NamaMK: "Pemrograman Web", SKS: 3, Semester: 3, Kuota: 30},
		{KodeMK: "SI104", NamaMK: "Jaringan Komputer", SKS: 3, Semester: 4, Kuota: 30},
		{KodeMK: "SI105", NamaMK: "Rekayasa Perangkat Lunak", SKS: 3, Semester: 5, Kuota: 30},
		{KodeMK: "SI106", NamaMK: "Kecerdasan Buatan", SKS: 3, Semester: 6, Kuota: 30},
		{KodeMK: "SI107", NamaMK: "Keamanan Sistem", SKS: 3, Semester: 7, Kuota: 30},
		{KodeMK: "SI108", NamaMK: "E-Business", SKS: 3, Semester: 8, Kuota: 30},
		{KodeMK: "SI109", NamaMK: "Sistem Operasi", SKS: 3, Semester: 2, Kuota: 30},
		{KodeMK: "SI110", NamaMK: "Struktur Data", SKS: 3, Semester: 3, Kuota: 30},
	}
	for _, course := range courses {
		DB.Create(&course)
	}

	log.Println("Database seeded")
}
