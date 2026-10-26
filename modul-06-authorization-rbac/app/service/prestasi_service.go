package service

import (
	"api-students/app/repository"
	"api-students/helper"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type PrestasiService struct {
	r *repository.PrestasiRepository
}

func NewPrestasiService(repo *repository.PrestasiRepository) *PrestasiService {
	return &PrestasiService{r: repo}
}

func (s *PrestasiService) ListByStudentID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "invalid id")
	}

	data, err := s.r.ListByStudentID(c.UserContext(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "err fetch")
	}
	return helper.Success(c, fiber.StatusOK, "ok", data)
}
