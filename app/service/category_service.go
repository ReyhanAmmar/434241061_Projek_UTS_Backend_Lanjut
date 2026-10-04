package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
	"api-buku-kas/app/repository"
	"api-buku-kas/helper"
)

type CategoryService struct {
	repo  repository.CategoryRepository
	perms *helper.PermissionSet
}

func NewCategoryService(
	repo repository.CategoryRepository,
	perms *helper.PermissionSet,
) *CategoryService {
	return &CategoryService{
		repo:  repo,
		perms: perms,
	}
}

func (s *CategoryService) List(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var ownerID *int

	if !s.perms.Can(current.Role, "category:list:any") {
		id := current.UserID
		ownerID = &id
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	categories, err := s.repo.FindAll(ctx, ownerID)
	if err != nil {
		return translateCategoryError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"daftar kategori berhasil diambil",
		categories,
	)
}

func (s *CategoryService) Get(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateCategoryError(err)
	}

	if !CanAccessCategory(
		current, category.OwnerID, s.perms, "category:read:any",
	) {
		return helper.Forbidden(
			"tidak berhak mengakses kategori pengguna lain",
		)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"kategori ditemukan",
		category,
	)
}

func (s *CategoryService) Create(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateCategoryRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	category, err := s.repo.Create(ctx, model.Category{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     current.UserID,
	})
	if err != nil {
		return translateCategoryError(err)
	}

	return helper.Created(
		c,
		"kategori berhasil dibuat",
		category,
		"/api/v1/categories/"+strconv.Itoa(category.ID),
	)
}

func (s *CategoryService) Replace(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateCategoryError(err)
	}

	if !CanAccessCategory(
		current, category.OwnerID, s.perms, "category:update:any",
	) {
		return helper.Forbidden(
			"tidak berhak mengubah kategori pengguna lain",
		)
	}

	var req model.ReplaceCategoryRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	category.Name = req.Name
	category.Description = req.Description

	updated, err := s.repo.Update(ctx, category)
	if err != nil {
		return translateCategoryError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"kategori berhasil diganti",
		updated,
	)
}

func (s *CategoryService) Patch(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateCategoryError(err)
	}

	if !CanAccessCategory(
		current, category.OwnerID, s.perms, "category:update:any",
	) {
		return helper.Forbidden(
			"tidak berhak mengubah kategori pengguna lain",
		)
	}

	var req model.PatchCategoryRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyCategoryPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	updated, errs := ApplyCategoryPatch(category, req)
	if errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateCategoryError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"kategori berhasil diperbarui sebagian",
		result,
	)
}

func (s *CategoryService) Delete(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateCategoryError(err)
	}

	if !CanAccessCategory(
		current, category.OwnerID, s.perms, "category:delete:any",
	) {
		return helper.Forbidden(
			"tidak berhak menghapus kategori pengguna lain",
		)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateCategoryError(err)
	}

	return helper.NoContent(c)
}

func translateCategoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("kategori tidak ditemukan")

	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict(
			"nama kategori sudah digunakan oleh pemilik yang sama",
		)

	case errors.Is(err, repository.ErrCategoryInUse):
		return helper.Conflict(
			"kategori masih digunakan dan tidak dapat dihapus",
		)

	default:
		return helper.Internal(err)
	}
}