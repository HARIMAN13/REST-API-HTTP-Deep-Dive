package controllers

import (
	"siakad-mini/database"
	"siakad-mini/models"
	"siakad-mini/utils"

	"github.com/gofiber/fiber/v2"
)

func GetCourses(c *fiber.Ctx) error {
	semester := c.Query("semester", "")
	search := c.Query("search", "")
	available := c.Query("available", "")

	query := database.DB.Model(&models.Course{}).Preload("Enrollments")

	if semester != "" {
		query = query.Where("semester = ?", semester)
	}

	if search != "" {
		query = query.Where("kode_mk ILIKE ? OR nama_mk ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var dbCourses []models.Course
	if err := query.Find(&dbCourses).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	var result []fiber.Map

	for _, course := range dbCourses {
		terisi := len(course.Enrollments)
		sisaKuota := course.Kuota - terisi

		if available == "true" && sisaKuota <= 0 {
			continue
		}

		result = append(result, fiber.Map{
			"id":         course.ID,
			"kode_mk":    course.KodeMK,
			"nama_mk":    course.NamaMK,
			"sks":        course.SKS,
			"semester":   course.Semester,
			"kuota":      course.Kuota,
			"terisi":     terisi,
			"sisa_kuota": sisaKuota,
		})
	}

	if result == nil {
		result = []fiber.Map{}
	}

	return utils.SendSuccess(c, "Daftar mata kuliah berhasil diambil", result)
}
