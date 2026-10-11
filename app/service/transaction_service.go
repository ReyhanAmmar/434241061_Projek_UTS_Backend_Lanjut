package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
	"api-buku-kas/app/repository"
	"api-buku-kas/helper"
)

type TransactionService struct {
	repo       repository.TransactionRepository
	categories repository.CategoryRepository
	perms      *helper.PermissionSet
}

func NewTransactionService(
	repo repository.TransactionRepository,
	categories repository.CategoryRepository,
	perms *helper.PermissionSet,
) *TransactionService {
	return &TransactionService{
		repo:       repo,
		categories: categories,
		perms:      perms,
	}
}

func (s *TransactionService) List(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	q, err := helper.ParseTransactionQuery(c)
	if err != nil {
		return err
	}

	if !s.perms.Can(current.Role, "transaction:list:any") {
		ownerID := current.UserID
		q.OwnerID = &ownerID
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.SuccessList(
		c,
		"daftar transaksi berhasil diambil",
		items,
		&model.Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: CountTotalPages(total, q.Limit),
		},
	)
}

func (s *TransactionService) Get(c *fiber.Ctx) error {
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

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateTransactionError(err)
	}

	if !CanAccessTransaction(
		current, item.OwnerID, s.perms, "transaction:read:any",
	) {
		return helper.Forbidden("tidak berhak mengakses transaksi pengguna lain")
	}

	return helper.Success(c, fiber.StatusOK, "transaksi ditemukan", item)
}

func (s *TransactionService) Create(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateTransactionRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Note = strings.TrimSpace(req.Note)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	date := time.Now()
	if req.TransactionDate != nil {
		date = *req.TransactionDate
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.ensureCategoryUsable(
		ctx, req.CategoryID, current.UserID, req.Type,
	); err != nil {
		return err
	}

	created, err := s.repo.Create(ctx, model.Transaction{
		Type:            req.Type,
		Amount:          req.Amount,
		Note:            req.Note,
		TransactionDate: date,
		CategoryID:      req.CategoryID,
		OwnerID:         current.UserID,
	})
	if err != nil {
		return translateTransactionError(err)
	}

	return helper.Created(
		c,
		"transaksi berhasil dibuat",
		created,
		"/api/v1/transactions/"+strconv.Itoa(created.ID),
	)
}

func (s *TransactionService) Replace(c *fiber.Ctx) error {
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

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateTransactionError(err)
	}

	if !CanAccessTransaction(
		current, item.OwnerID, s.perms, "transaction:update:any",
	) {
		return helper.Forbidden("tidak berhak mengubah transaksi pengguna lain")
	}

	var req model.ReplaceTransactionRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Note = strings.TrimSpace(req.Note)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if err := s.ensureCategoryUsable(
		ctx, req.CategoryID, item.OwnerID, req.Type,
	); err != nil {
		return err
	}

	item.Type = req.Type
	item.Amount = req.Amount
	item.Note = req.Note
	item.TransactionDate = req.TransactionDate
	item.CategoryID = req.CategoryID

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return translateTransactionError(err)
	}

	return helper.Success(
		c, fiber.StatusOK, "transaksi berhasil diganti", updated,
	)
}

func (s *TransactionService) Patch(c *fiber.Ctx) error {
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

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateTransactionError(err)
	}

	if !CanAccessTransaction(
		current, item.OwnerID, s.perms, "transaction:update:any",
	) {
		return helper.Forbidden("tidak berhak mengubah transaksi pengguna lain")
	}

	var req model.PatchTransactionRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyTransactionPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if req.Note != nil {
		trimmed := strings.TrimSpace(*req.Note)
		req.Note = &trimmed
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated := ApplyTransactionPatch(item, req)

	if req.CategoryID != nil || req.Type != nil {
		if err := s.ensureCategoryUsable(
			ctx, updated.CategoryID, item.OwnerID, updated.Type,
		); err != nil {
			return err
		}
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateTransactionError(err)
	}

	return helper.Success(
		c, fiber.StatusOK, "transaksi berhasil diperbarui sebagian", result,
	)
}

func (s *TransactionService) Delete(c *fiber.Ctx) error {
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

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateTransactionError(err)
	}

	if !CanAccessTransaction(
		current, item.OwnerID, s.perms, "transaction:delete:any",
	) {
		return helper.Forbidden("tidak berhak menghapus transaksi pengguna lain")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateTransactionError(err)
	}

	return helper.NoContent(c)
}

func (s *TransactionService) ensureCategoryUsable(
	ctx context.Context,
	categoryID int,
	ownerID int,
	transactionType string,
) error {
	category, err := s.categories.FindByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return invalidCategoryError()
		}
		return helper.Internal(err)
	}

	if category.OwnerID != ownerID {
		return invalidCategoryError()
	}

	if !CategoryMatchesType(category.Type, transactionType) {
		return helper.Validation(map[string]string{
			"category_id": "jenis kategori tidak sesuai dengan jenis transaksi",
		})
	}

	return nil
}

func invalidCategoryError() error {
	return helper.Validation(map[string]string{
		"category_id": "kategori tidak ditemukan",
	})
}

func translateTransactionError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("transaksi tidak ditemukan")

	case errors.Is(err, repository.ErrInvalidReference):
		return invalidCategoryError()

	default:
		return helper.Internal(err)
	}
}