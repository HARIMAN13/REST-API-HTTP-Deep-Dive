package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	portfolio, err := s.portfolios.FindByStudentID(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", portfolio)
}

func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		req.Projects[i].Title = strings.TrimSpace(req.Projects[i].Title)
		req.Projects[i].Technologies = NormalizeTags(req.Projects[i].Technologies)
		req.Projects[i].Link = strings.TrimSpace(req.Projects[i].Link)
	}
	for i := range req.Certificates {
		req.Certificates[i].Name = strings.TrimSpace(req.Certificates[i].Name)
		req.Certificates[i].Issuer = strings.TrimSpace(req.Certificates[i].Issuer)
		req.Certificates[i].Date = strings.TrimSpace(req.Certificates[i].Date)
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.Upsert(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	var req model.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Technologies = NormalizeTags(req.Technologies)
	req.Link = strings.TrimSpace(req.Link)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.AddProject(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", portfolio)
}

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("id proyek tidak boleh kosong")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	if err := s.portfolios.DeleteProject(ctx, studentID, projectID); err != nil {
		return translatePortfolioError(err)
	}

	return helper.NoContent(c)
}

func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 50 {
		limit = 10
	}

	stats, err := s.portfolios.TopTechnologies(ctx, limit)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas")
	default:
		return helper.Internal(err)
	}
}
