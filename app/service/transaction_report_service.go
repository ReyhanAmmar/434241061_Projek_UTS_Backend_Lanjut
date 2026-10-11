package service

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
	"api-buku-kas/helper"
)

const exportMaxRows = 10000

var exportHeader = []string{
	"id", "transaction_date", "type", "category", "amount", "note",
}

func (s *TransactionService) ownerScope(
	current model.AuthUser,
	q *model.TransactionQuery,
) {
	if !s.perms.Can(current.Role, "transaction:list:any") {
		ownerID := current.UserID
		q.OwnerID = &ownerID
	}
}

func (s *TransactionService) Feed(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	q, err := helper.ParseTransactionQuery(c)
	if err != nil {
		return err
	}

	cursor, err := helper.DecodeCursor(c.Query("cursor"))
	if err != nil {
		return err
	}

	s.ownerScope(current, &q)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	items, err := s.repo.FindAfterCursor(ctx, q, cursor, q.Limit)
	if err != nil {
		return helper.Internal(err)
	}

	page, hasMore := TrimPage(items, q.Limit)

	var next *string
	if hasMore {
		last := page[len(page)-1]
		token := helper.EncodeCursor(model.Cursor{
			Date: last.TransactionDate,
			ID:   last.ID,
		})
		next = &token
	}

	return helper.SuccessCursor(
		c,
		"feed transaksi berhasil diambil",
		page,
		model.CursorMeta{
			Limit:      q.Limit,
			HasMore:    hasMore,
			NextCursor: next,
		},
	)
}

func (s *TransactionService) Summary(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	q, err := helper.ParseTransactionQuery(c)
	if err != nil {
		return err
	}

	if q.From == nil && q.To == nil {
		from, to := CurrentMonthRange(time.Now(), helper.AppLocation)
		q.From = &from
		q.To = &to
	}

	s.ownerScope(current, &q)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	summary, err := s.repo.Summarize(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	summary.Period = model.Period{From: q.From, To: q.To}
	summary.Balance = ComputeBalance(summary.TotalIncome, summary.TotalExpense)

	return helper.Success(
		c, fiber.StatusOK, "ringkasan transaksi berhasil dihitung", summary,
	)
}

func (s *TransactionService) Export(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	format, err := helper.NegotiateFormat(c)
	if err != nil {
		return err
	}

	q, err := helper.ParseTransactionQuery(c)
	if err != nil {
		return err
	}

	q.Page = 1
	q.Limit = exportMaxRows

	s.ownerScope(current, &q)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	if total > exportMaxRows {
		c.Set("X-Truncated", "true")
	}

	if format == "json" {
		return helper.Success(
			c, fiber.StatusOK, "data transaksi berhasil diekspor", items,
		)
	}

	names, err := s.categoryNames(ctx, q.OwnerID)
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, ExportRow(
			item, names[item.CategoryID],
			helper.AppLocation, helper.SanitizeCSVCell,
		))
	}

	body, err := helper.BuildCSV(exportHeader, rows)
	if err != nil {
		return helper.Internal(err)
	}

	filename := "transaksi-" +
		time.Now().In(helper.AppLocation).Format("20060102") + ".csv"

	c.Set(fiber.HeaderContentType, helper.MIMETextCSV+"; charset=utf-8")
	c.Set(
		fiber.HeaderContentDisposition,
		`attachment; filename="`+filename+`"`,
	)

	return c.Status(fiber.StatusOK).Send(body)
}

func (s *TransactionService) categoryNames(
	ctx context.Context,
	ownerID *int,
) (map[int]string, error) {
	categories, err := s.categories.FindAll(ctx, ownerID, "")
	if err != nil {
		return nil, helper.Internal(err)
	}

	names := make(map[int]string, len(categories))
	for _, category := range categories {
		names[category.ID] = category.Name
	}

	return names, nil
}