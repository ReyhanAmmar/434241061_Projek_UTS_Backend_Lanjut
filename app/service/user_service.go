package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
	"api-buku-kas/app/repository"
	"api-buku-kas/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{
		repo:  repo,
		perms: perms,
	}
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"daftar user berhasil diambil",
		users,
	)
}

func (s *UserService) Get(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden(
			"tidak berhak mengakses data user lain",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("user tidak ditemukan")
		}

		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"user ditemukan",
		user,
	)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRoleRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(
		current, id, req, s.perms,
	); len(errs) > 0 {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.UpdateRole(
		ctx, id, strings.TrimSpace(req.Role),
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("user tidak ditemukan")
		}

		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"role user berhasil diubah",
		user,
	)
}