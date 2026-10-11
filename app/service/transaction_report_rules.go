package service

import (
	"strconv"
	"time"

	"api-buku-kas/app/model"
)

func TrimPage(
	items []model.Transaction, limit int,
) ([]model.Transaction, bool) {
	if limit < 1 || len(items) <= limit {
		return items, false
	}

	return items[:limit], true
}

func ComputeBalance(income, expense int64) int64 {
	return income - expense
}

func CurrentMonthRange(now time.Time, loc *time.Location) (time.Time, time.Time) {
	local := now.In(loc)
	from := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)

	return from, from.AddDate(0, 1, 0)
}

func ExportRow(
	t model.Transaction,
	categoryName string,
	loc *time.Location,
	sanitize func(string) string,
) []string {
	return []string{
		strconv.Itoa(t.ID),
		t.TransactionDate.In(loc).Format("2006-01-02 15:04:05"),
		t.Type,
		sanitize(categoryName),
		strconv.FormatInt(t.Amount, 10),
		sanitize(t.Note),
	}
}