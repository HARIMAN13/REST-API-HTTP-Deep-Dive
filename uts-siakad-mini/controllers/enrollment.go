package controllers

import (
	"fmt"
	"regexp"

	"siakad-mini/database"
	"siakad-mini/models"
	"siakad-mini/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateEnrollmentInput struct {
	CourseID      uint   `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

func CreateEnrollment(c *fiber.Ctx) error {
	var input CreateEnrollmentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.CourseID == 0 || input.TahunAkademik == "" {
		errors := make(map[string][]string)
		if input.CourseID == 0 {
			errors["course_id"] = []string{"Course ID wajib diisi"}
		}
		if input.TahunAkademik == "" {
			errors["tahun_akademik"] = []string{"Tahun akademik wajib diisi"}
		}
		return utils.SendValidationError(c, "Validasi gagal", errors)
	}

	match, _ := regexp.MatchString(`^\d{4}/\d{4}-(Ganjil|Genap)$`, input.TahunAkademik)
	if !match {
		errors := map[string][]string{
			"tahun_akademik": {"Format harus seperti 2026/2027-Ganjil"},
		}
		return utils.SendValidationError(c, "Validasi gagal", errors)
	}

	userID := c.Locals("user_id").(uint)

	var student models.Student
	if err := database.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat mengambil KRS")
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var course models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&course, input.CourseID).Error; err != nil {
			return fmt.Errorf("course_not_found")
		}

		var existingEnrollment models.Enrollment
		errExist := tx.Where("student_id = ? AND course_id = ? AND tahun_akademik = ?", student.ID, input.CourseID, input.TahunAkademik).First(&existingEnrollment).Error
		if errExist == nil {
			return fmt.Errorf("conflict")
		}

		var currentEnrollments []models.Enrollment
		tx.Preload("Course").Where("student_id = ? AND tahun_akademik = ?", student.ID, input.TahunAkademik).Find(&currentEnrollments)

		totalSKS := 0
		for _, e := range currentEnrollments {
			if e.Course != nil {
				totalSKS += e.Course.SKS
			}
		}

		batasSKS := 18
		if student.IPKTerakhir >= 3.00 {
			batasSKS = 24
		} else if student.IPKTerakhir >= 2.50 {
			batasSKS = 21
		}

		if totalSKS+course.SKS > batasSKS {
			sisaSKS := batasSKS - totalSKS
			return fmt.Errorf("sks_limit|%d", sisaSKS)
		}

		var enrolledCount int64
		tx.Model(&models.Enrollment{}).Where("course_id = ? AND tahun_akademik = ?", course.ID, input.TahunAkademik).Count(&enrolledCount)

		if int(enrolledCount) >= course.Kuota {
			return fmt.Errorf("quota_full")
		}

		newEnrollment := models.Enrollment{
			StudentID:     student.ID,
			CourseID:      course.ID,
			TahunAkademik: input.TahunAkademik,
		}

		if err := tx.Create(&newEnrollment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "course_not_found" {
			return utils.SendError(c, fiber.StatusNotFound, "Mata kuliah tidak ditemukan")
		}
		if err.Error() == "conflict" {
			return utils.SendError(c, fiber.StatusConflict, "Mata kuliah sudah diambil pada tahun akademik ini")
		}
		if err.Error() == "quota_full" {
			return utils.SendError(c, fiber.StatusUnprocessableEntity, "Kuota kelas sudah penuh")
		}
		if len(err.Error()) > 10 && err.Error()[:10] == "sks_limit|" {
			sisa := err.Error()[10:]
			return utils.SendError(c, fiber.StatusUnprocessableEntity, "SKS tidak mencukupi, sisa SKS Anda: "+sisa)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menambahkan mata kuliah")
	}

	return utils.SendSuccessCreated(c, "Berhasil mengambil mata kuliah", nil)
}

func DeleteEnrollment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var student models.Student
	if err := database.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		return utils.SendError(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat menghapus KRS")
	}

	var enrollment models.Enrollment
	if err := database.DB.First(&enrollment, id).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Data KRS tidak ditemukan")
	}

	if enrollment.StudentID != student.ID {
		return utils.SendError(c, fiber.StatusForbidden, "Tidak dapat membatalkan KRS milik orang lain")
	}

	if err := database.DB.Delete(&enrollment).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membatalkan mata kuliah")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
