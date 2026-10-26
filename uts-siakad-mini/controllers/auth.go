package controllers

import (
	"strings"

	"siakad-mini/database"
	"siakad-mini/models"
	"siakad-mini/utils"

	"github.com/gofiber/fiber/v2"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.Email == "" || !strings.Contains(input.Email, "@") || len(input.Password) < 8 {
		errors := make(map[string][]string)
		if input.Email == "" || !strings.Contains(input.Email, "@") {
			errors["email"] = []string{"Format email tidak valid"}
		}
		if len(input.Password) < 8 {
			errors["password"] = []string{"Password minimum 8 characters"}
		}
		return utils.SendValidationError(c, "Validasi gagal", errors)
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Kredensial salah")
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return utils.SendError(c, fiber.StatusUnauthorized, "Kredensial salah")
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	data := fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   72 * 3600,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	}

	return c.Status(fiber.StatusOK).JSON(data)
}

func GetMe(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	var user models.User
	if err := database.DB.Preload("Student").First(&user, userID).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "User not found")
	}

	userData := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if role == "mahasiswa" && user.Student != nil {
		userData["student"] = fiber.Map{
			"nim":      user.Student.NIM,
			"nama":     user.Student.Nama,
			"prodi":    user.Student.Prodi,
			"angkatan": user.Student.Angkatan,
		}
	}

	return utils.SendSuccess(c, "Berhasil mengambil profil", userData)
}
