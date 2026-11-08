package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type SupervisorService struct {
	repo repository.SupervisorRepository
	tx   *repository.TxManager
}

func NewSupervisorService(repo repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{repo: repo, tx: tx}
}

func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisors, err := s.repo.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", supervisors)
}

func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sup, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", sup)
}

func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var assigned model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.repo.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		sup, err := repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}
		assigned = sup

		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "berhasil menetapkan pembimbing", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if req.FromSupervisorID == req.ToSupervisorID {
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan from_supervisor_id",
		})
	}

	var joined model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.repo.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		if err := repo.LockSupervisors(ctx, req.FromSupervisorID, req.ToSupervisorID); err != nil {
			return err
		}

		if err := repo.DeleteSupervision(ctx, req.StudentID, req.FromSupervisorID); err != nil {
			return err
		}

		if _, err := repo.DecrementSupervised(ctx, req.FromSupervisorID); err != nil {
			return err
		}

		sup, err := repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}
		joined = sup

		return repo.InsertSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti pembimbing", joined)
}

func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.repo.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrStudentAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa tidak terdaftar pada pembimbing tersebut")
	case errors.Is(err, repository.ErrSupervisorVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT", "data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrSupervisorQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED", "kuota tidak boleh lebih kecil dari jumlah mahasiswa yang dibimbing saat ini")
	default:
		return helper.Internal(err)
	}
}
