package repository

import (
	"context"
	"fmt"
	"time"

	"api-buku-kas/app/model"
)

func (r *transactionPostgresRepository) FindAfterCursor(
	ctx context.Context,
	q model.TransactionQuery,
	cursor *model.Cursor,
	limit int,
) ([]model.Transaction, error) {
	where, args := buildTransactionFilter(q)

	if cursor != nil {
		args = append(args, cursor.Date, cursor.ID)
		where += fmt.Sprintf(
			" AND (transaction_date, id) < ($%d, $%d)",
			len(args)-1, len(args),
		)
	}

	args = append(args, limit+1)

	sqlText := fmt.Sprintf(
		`SELECT %s FROM transactions%s
		 ORDER BY transaction_date DESC, id DESC
		 LIMIT $%d`,
		transactionColumns, where, len(args),
	)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil feed transaksi: %w", err)
	}
	defer rows.Close()

	result := []model.Transaction{}

	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris feed transaksi: %w", err)
		}
		result = append(result, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil feed transaksi: %w", err)
	}

	return result, nil
}

func (r *transactionPostgresRepository) Summarize(
	ctx context.Context,
	q model.TransactionQuery,
) (model.TransactionSummary, error) {
	where, args := buildTransactionFilter(q)

	summary := model.TransactionSummary{
		ByCategory: []model.CategoryTotal{},
		ByDay:      []model.DayTotal{},
	}

	err := r.pool.QueryRow(ctx,
		`SELECT
		     COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::BIGINT,
		     COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::BIGINT,
		     COUNT(*)
		 FROM transactions`+where,
		args...,
	).Scan(
		&summary.TotalIncome,
		&summary.TotalExpense,
		&summary.TransactionCount,
	)
	if err != nil {
		return summary, fmt.Errorf("menghitung total transaksi: %w", err)
	}

	catRows, err := r.pool.Query(ctx,
		`SELECT t.category_id, c.name, t.type,
		        SUM(t.amount)::BIGINT, COUNT(*)
		 FROM (
		     SELECT category_id, type, amount FROM transactions`+where+`
		 ) t
		 JOIN categories c ON c.id = t.category_id
		 GROUP BY t.category_id, c.name, t.type
		 ORDER BY SUM(t.amount) DESC, c.name ASC`,
		args...,
	)
	if err != nil {
		return summary, fmt.Errorf("menghitung total per kategori: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var item model.CategoryTotal

		if err := catRows.Scan(
			&item.CategoryID,
			&item.Category,
			&item.Type,
			&item.Total,
			&item.Count,
		); err != nil {
			return summary, fmt.Errorf("membaca total per kategori: %w", err)
		}

		summary.ByCategory = append(summary.ByCategory, item)
	}

	if err := catRows.Err(); err != nil {
		return summary, fmt.Errorf("membaca hasil total per kategori: %w", err)
	}

	dayRows, err := r.pool.Query(ctx,
		`SELECT (transaction_date AT TIME ZONE 'Asia/Jakarta')::date AS day,
		        COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::BIGINT,
		        COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::BIGINT,
		        COUNT(*)
		 FROM transactions`+where+`
		 GROUP BY day
		 ORDER BY day ASC`,
		args...,
	)
	if err != nil {
		return summary, fmt.Errorf("menghitung total per hari: %w", err)
	}
	defer dayRows.Close()

	for dayRows.Next() {
		var (
			day  time.Time
			item model.DayTotal
		)

		if err := dayRows.Scan(
			&day, &item.Income, &item.Expense, &item.Count,
		); err != nil {
			return summary, fmt.Errorf("membaca total per hari: %w", err)
		}

		item.Date = day.Format("2006-01-02")
		summary.ByDay = append(summary.ByDay, item)
	}

	if err := dayRows.Err(); err != nil {
		return summary, fmt.Errorf("membaca hasil total per hari: %w", err)
	}

	return summary, nil
}