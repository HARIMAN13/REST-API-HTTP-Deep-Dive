package service

import (
	"api-students/app/repository"
	"api-students/helper"
	"github.com/gofiber/fiber/v2"
	"strconv"
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
		return helper.BadRequest("invalid id")
	}

	data, err := s.r.ListByStudentID(c.UserContext(), id)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "ok", data)
}
