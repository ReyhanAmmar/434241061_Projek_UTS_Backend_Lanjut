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

type categoryPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(
	pool *pgxpool.Pool,
) CategoryRepository {
	return &categoryPostgresRepository{pool: pool}
}

func (r *categoryPostgresRepository) FindAll(
	ctx context.Context,
	ownerID *int,
) ([]model.Category, error) {
	sqlText := `
		SELECT id, name, description, owner_id,
		       created_at, updated_at
		FROM categories`

	args := []any{}

	if ownerID != nil {
		sqlText += " WHERE owner_id = $1"
		args = append(args, *ownerID)
	}

	sqlText += " ORDER BY id ASC"

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kategori: %w", err)
	}
	defer rows.Close()

	result := []model.Category{}

	for rows.Next() {
		var category model.Category

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.OwnerID,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
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
	var category model.Category

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, owner_id,
		        created_at, updated_at
		 FROM categories
		 WHERE id = $1`,
		id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.OwnerID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

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
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		category.Name,
		category.Description,
		category.OwnerID,
	).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Category{}, ErrDuplicate
		}

		return model.Category{}, fmt.Errorf(
			"menyimpan kategori: %w", err,
		)
	}

	return category, nil
}

func (r *categoryPostgresRepository) Update(
	ctx context.Context,
	category model.Category,
) (model.Category, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE categories
		 SET name = $1,
		     description = $2,
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, name, description, owner_id,
		           created_at, updated_at`,
		category.Name,
		category.Description,
		category.ID,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.OwnerID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

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

	return category, nil
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