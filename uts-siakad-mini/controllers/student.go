package controllers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"siakad-mini/database"
	"siakad-mini/models"
	"siakad-mini/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetStudents(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if perPage > 50 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}

	search := c.Query("search", "")
	prodi := c.Query("prodi", "")
	angkatan := c.Query("angkatan", "")
	sort := c.Query("sort", "")

	query := database.DB.Model(&models.Student{})

	if search != "" {
		query = query.Where("nim ILIKE ? OR nama ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if prodi != "" {
		query = query.Where("prodi = ?", prodi)
	}
	if angkatan != "" {
		query = query.Where("angkatan = ?", angkatan)
	}

	if sort == "nama" {
		query = query.Order("nama asc")
	} else if sort == "-ipk_terakhir" {
		query = query.Order("ipk_terakhir desc")
	} else {
		query = query.Order("id desc")
	}

	var total int64
	query.Count(&total)

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage == 0 {
		lastPage = 1
	}

	var students []models.Student
	query.Offset((page - 1) * perPage).Limit(perPage).Find(&students)

	meta := models.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return utils.SendSuccessPaginated(c, "Data mahasiswa berhasil diambil", students, meta)
}

type CreateStudentInput struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

func CreateStudent(c *fiber.Ctx) error {
	var input CreateStudentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	errors := make(map[string][]string)

	if len(input.NIM) != 12 {
		errors["nim"] = []string{"NIM harus 12 digit"}
	}
	if input.Nama == "" {
		errors["nama"] = []string{"Nama wajib diisi"}
	}
	if input.Email == "" || !strings.Contains(input.Email, "@") {
		errors["email"] = []string{"Format email tidak valid"}
	}
	if input.Prodi == "" {
		errors["prodi"] = []string{"Prodi wajib diisi"}
	}
	currentYear := time.Now().Year()
	if input.Angkatan < 2000 || input.Angkatan > currentYear {
		errors["angkatan"] = []string{"Angkatan tidak valid"}
	}
	if input.IPKTerakhir < 0 || input.IPKTerakhir > 4.00 {
		errors["ipk_terakhir"] = []string{"IPK harus diantara 0.00 dan 4.00"}
	}

	if len(errors) > 0 {
		return utils.SendValidationError(c, "Validasi gagal", errors)
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var existingUser models.User
		if err := tx.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
			return fmt.Errorf("email_exists")
		}

		var existingStudent models.Student
		if err := tx.Where("nim = ?", input.NIM).First(&existingStudent).Error; err == nil {
			return fmt.Errorf("nim_exists")
		}

		hashedPassword, _ := utils.HashPassword(input.NIM)

		user := models.User{
			Email:    input.Email,
			Password: hashedPassword,
			Role:     "mahasiswa",
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		student := models.Student{
			UserID:      user.ID,
			NIM:         input.NIM,
			Nama:        input.Nama,
			Prodi:       input.Prodi,
			Angkatan:    input.Angkatan,
			IPKTerakhir: input.IPKTerakhir,
		}

		if err := tx.Create(&student).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "email_exists" {
			errors["email"] = []string{"Email sudah terdaftar"}
			return utils.SendValidationError(c, "Validasi gagal", errors)
		}
		if err.Error() == "nim_exists" {
			errors["nim"] = []string{"NIM sudah terdaftar"}
			return utils.SendValidationError(c, "Validasi gagal", errors)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menambahkan data mahasiswa")
	}

	return utils.SendSuccessCreated(c, "Berhasil menambahkan mahasiswa", nil)
}

func GetStudentDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	var student models.Student
	if err := database.DB.Preload("Enrollments.Course").First(&student, id).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	if role == "mahasiswa" && student.UserID != userID {
		return utils.SendError(c, fiber.StatusForbidden, "Tidak memiliki akses ke data ini")
	}

	var totalSKS int
	var courses []fiber.Map

	for _, enrollment := range student.Enrollments {
		if enrollment.Course != nil {
			totalSKS += enrollment.Course.SKS
			courses = append(courses, fiber.Map{
				"id":             enrollment.Course.ID,
				"kode_mk":        enrollment.Course.KodeMK,
				"nama_mk":        enrollment.Course.NamaMK,
				"sks":            enrollment.Course.SKS,
				"semester":       enrollment.Course.Semester,
				"tahun_akademik": enrollment.TahunAkademik,
			})
		}
	}

	batasSKS := 18
	if student.IPKTerakhir >= 3.00 {
		batasSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		batasSKS = 21
	}

	data := fiber.Map{
		"student":   student,
		"courses":   courses,
		"total_sks": totalSKS,
		"batas_sks": batasSKS,
	}

	return utils.SendSuccess(c, "Detail mahasiswa berhasil diambil", data)
}

type UpdateStudentInput struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

func UpdateStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var student models.Student
	if err := database.DB.First(&student, id).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	var input UpdateStudentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	errors := make(map[string][]string)

	if input.Nama == "" {
		errors["nama"] = []string{"Nama wajib diisi"}
	}
	if input.Prodi == "" {
		errors["prodi"] = []string{"Prodi wajib diisi"}
	}
	currentYear := time.Now().Year()
	if input.Angkatan < 2000 || input.Angkatan > currentYear {
		errors["angkatan"] = []string{"Angkatan tidak valid"}
	}
	if input.IPKTerakhir < 0 || input.IPKTerakhir > 4.00 {
		errors["ipk_terakhir"] = []string{"IPK harus diantara 0.00 dan 4.00"}
	}

	if len(errors) > 0 {
		return utils.SendValidationError(c, "Validasi gagal", errors)
	}

	student.Nama = input.Nama
	student.Prodi = input.Prodi
	student.Angkatan = input.Angkatan
	student.IPKTerakhir = input.IPKTerakhir

	if err := database.DB.Save(&student).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui data mahasiswa")
	}

	return utils.SendSuccess(c, "Berhasil memperbarui data mahasiswa", student)
}

func DeleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var student models.Student
	if err := database.DB.First(&student, id).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&student).Error; err != nil {
			return err
		}

		var user models.User
		if err := tx.First(&user, student.UserID).Error; err == nil {
			if err := tx.Delete(&user).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus mahasiswa")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
