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

var ErrCategoryInUse = errors.New("kategori masih digunakan")

type CategoryRepository interface {
	FindAll(
		ctx context.Context,
		ownerID *int,
		categoryType string,
	) ([]model.Category, error)

	FindByID(
		ctx context.Context,
		id int,
	) (model.Category, error)

	Create(
		ctx context.Context,
		category model.Category,
	) (model.Category, error)

	Update(
		ctx context.Context,
		category model.Category,
	) (model.Category, error)

	Delete(ctx context.Context, id int) error
}

const categoryColumns = `id, name, description, type, owner_id,
	created_at, updated_at`

type categoryPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(
	pool *pgxpool.Pool,
) CategoryRepository {
	return &categoryPostgresRepository{pool: pool}
}

func scanCategory(row pgx.Row) (model.Category, error) {
	var category model.Category

	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Type,
		&category.OwnerID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	return category, err
}

func (r *categoryPostgresRepository) FindAll(
	ctx context.Context,
	ownerID *int,
	categoryType string,
) ([]model.Category, error) {
	sqlText := "SELECT " + categoryColumns + " FROM categories WHERE 1 = 1"
	args := []any{}

	if ownerID != nil {
		args = append(args, *ownerID)
		sqlText += fmt.Sprintf(" AND owner_id = $%d", len(args))
	}

	if categoryType != "" {
		args = append(args, categoryType)
		sqlText += fmt.Sprintf(" AND type = $%d", len(args))
	}

	sqlText += " ORDER BY type ASC, LOWER(name) ASC, id ASC"

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kategori: %w", err)
	}
	defer rows.Close()

	result := []model.Category{}

	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris kategori: %w", err)
		}

		result = append(result, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query kategori: %w", err)
	}

	return result, nil
}

func (r *categoryPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Category, error) {
	category, err := scanCategory(r.pool.QueryRow(ctx,
		"SELECT "+categoryColumns+" FROM categories WHERE id = $1",
		id,
	))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrNotFound
		}

		return model.Category{}, fmt.Errorf(
			"mengambil kategori: %w", err,
		)
	}

	return category, nil
}

func (r *categoryPostgresRepository) Create(
	ctx context.Context,
	category model.Category,
) (model.Category, error) {
	created, err := scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO categories (name, description, type, owner_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+categoryColumns,
		category.Name,
		category.Description,
		category.Type,
		category.OwnerID,
	))

	if err != nil {
		if isUniqueViolation(err) {
			return model.Category{}, ErrDuplicate
		}

		return model.Category{}, fmt.Errorf(
			"menyimpan kategori: %w", err,
		)
	}

	return created, nil
}

func (r *categoryPostgresRepository) Update(
	ctx context.Context,
	category model.Category,
) (model.Category, error) {
	updated, err := scanCategory(r.pool.QueryRow(ctx,
		`UPDATE categories
		 SET name = $1,
		     description = $2,
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING `+categoryColumns,
		category.Name,
		category.Description,
		category.ID,
	))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrNotFound
		}

		if isUniqueViolation(err) {
			return model.Category{}, ErrDuplicate
		}

		return model.Category{}, fmt.Errorf(
			"memperbarui kategori: %w", err,
		)
	}

	return updated, nil
}

func (r *categoryPostgresRepository) Delete(
	ctx context.Context,
	id int,
) error {
	tag, err := r.pool.Exec(
		ctx,
		"DELETE FROM categories WHERE id = $1",
		id,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrCategoryInUse
		}

		return fmt.Errorf("menghapus kategori: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
