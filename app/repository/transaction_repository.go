package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-buku-kas/app/model"
)

var ErrInvalidReference = errors.New("referensi data tidak valid")

type TransactionRepository interface {
	FindAll(
		ctx context.Context,
		q model.TransactionQuery,
	) ([]model.Transaction, int, error)

	FindByID(ctx context.Context, id int) (model.Transaction, error)

	Create(
		ctx context.Context,
		t model.Transaction,
	) (model.Transaction, error)

	Update(
		ctx context.Context,
		t model.Transaction,
	) (model.Transaction, error)

	Delete(ctx context.Context, id int) error
}

const transactionColumns = `id, type, amount, note, transaction_date,
	category_id, owner_id, created_at, updated_at`

var transactionSortColumns = map[string]string{
	"transaction_date": "transaction_date",
	"amount":           "amount",
	"created_at":       "created_at",
	"id":               "id",
}

type transactionPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) TransactionRepository {
	return &transactionPostgresRepository{pool: pool}
}

func scanTransaction(row pgx.Row) (model.Transaction, error) {
	var t model.Transaction

	err := row.Scan(
		&t.ID,
		&t.Type,
		&t.Amount,
		&t.Note,
		&t.TransactionDate,
		&t.CategoryID,
		&t.OwnerID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	return t, err
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func buildTransactionFilter(q model.TransactionQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	add := func(condition string, value any) {
		args = append(args, value)
		where += " AND " + fmt.Sprintf(condition, len(args))
	}

	if q.OwnerID != nil {
		add("owner_id = $%d", *q.OwnerID)
	}
	if q.Type != "" {
		add("type = $%d", q.Type)
	}
	if q.CategoryID != nil {
		add("category_id = $%d", *q.CategoryID)
	}
	if q.From != nil {
		add("transaction_date >= $%d", *q.From)
	}
	if q.To != nil {
		add("transaction_date < $%d", *q.To)
	}
	if q.Search != "" {
		add("note ILIKE $%d", "%"+q.Search+"%")
	}

	return where, args
}

func (r *transactionPostgresRepository) FindAll(
	ctx context.Context,
	q model.TransactionQuery,
) ([]model.Transaction, int, error) {
	where, args := buildTransactionFilter(q)

	var total int
	err := r.pool.QueryRow(
		ctx, "SELECT COUNT(*) FROM transactions"+where, args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung transaksi: %w", err)
	}

	column, ok := transactionSortColumns[q.Sort]
	if !ok {
		column = "transaction_date"
	}

	direction := "DESC"
	if q.Order == "asc" {
		direction = "ASC"
	}

	sqlText := fmt.Sprintf(
		`SELECT %s FROM transactions%s
		 ORDER BY %s %s, id %s
		 LIMIT $%d OFFSET $%d`,
		transactionColumns, where,
		column, direction, direction,
		len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar transaksi: %w", err)
	}
	defer rows.Close()

	result := []model.Transaction{}

	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("membaca baris transaksi: %w", err)
		}
		result = append(result, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query transaksi: %w", err)
	}

	return result, total, nil
}

func (r *transactionPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Transaction, error) {
	t, err := scanTransaction(r.pool.QueryRow(ctx,
		"SELECT "+transactionColumns+" FROM transactions WHERE id = $1",
		id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Transaction{}, ErrNotFound
		}
		return model.Transaction{}, fmt.Errorf("mengambil transaksi: %w", err)
	}

	return t, nil
}

func (r *transactionPostgresRepository) Create(
	ctx context.Context,
	t model.Transaction,
) (model.Transaction, error) {
	created, err := scanTransaction(r.pool.QueryRow(ctx,
		`INSERT INTO transactions
		     (type, amount, note, transaction_date, category_id, owner_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+transactionColumns,
		t.Type, t.Amount, t.Note, t.TransactionDate, t.CategoryID, t.OwnerID,
	))
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Transaction{}, ErrInvalidReference
		}
		return model.Transaction{}, fmt.Errorf("menyimpan transaksi: %w", err)
	}

	return created, nil
}

func (r *transactionPostgresRepository) Update(
	ctx context.Context,
	t model.Transaction,
) (model.Transaction, error) {
	updated, err := scanTransaction(r.pool.QueryRow(ctx,
		`UPDATE transactions
		 SET type = $1, amount = $2, note = $3,
		     transaction_date = $4, category_id = $5,
		     updated_at = NOW()
		 WHERE id = $6
		 RETURNING `+transactionColumns,
		t.Type, t.Amount, t.Note, t.TransactionDate, t.CategoryID, t.ID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Transaction{}, ErrNotFound
		}
		if isForeignKeyViolation(err) {
			return model.Transaction{}, ErrInvalidReference
		}
		return model.Transaction{}, fmt.Errorf("memperbarui transaksi: %w", err)
	}

	return updated, nil
}

func (r *transactionPostgresRepository) Delete(
	ctx context.Context,
	id int,
) error {
	tag, err := r.pool.Exec(
		ctx, "DELETE FROM transactions WHERE id = $1", id,
	)
	if err != nil {
		return fmt.Errorf("menghapus transaksi: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}