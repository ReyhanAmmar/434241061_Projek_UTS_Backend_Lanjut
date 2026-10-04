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

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type UserRepository interface {
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) FindAll(
	ctx context.Context,
) ([]model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, role, is_active,
		        created_at, updated_at
		 FROM users
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	result := []model.User{}

	for rows.Next() {
		var user model.User

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca baris user: %w", err)
		}

		result = append(result, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query user: %w", err)
	}

	return result, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, role, is_active,
		        created_at, updated_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}

	return user, nil
}

func (r *userPostgresRepository) FindByEmail(
	ctx context.Context,
	email string,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password, role, is_active,
		        created_at, updated_at
		 FROM users
		 WHERE LOWER(email) = LOWER($1)`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf(
			"mengambil user berdasarkan email: %w", err,
		)
	}

	return user, nil
}

func (r *userPostgresRepository) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {
	if user.Role == "" {
		user.Role = model.RoleUser
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password, role, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}

	return user, nil
}

func (r *userPostgresRepository) UpdateRole(
	ctx context.Context,
	id int,
	role string,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx,
		`UPDATE users
		 SET role = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, name, email, role, is_active,
		           created_at, updated_at`,
		role,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}