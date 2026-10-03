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
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, role, is_active, created_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
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
	ctx context.Context, email string,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, password, role, is_active, created_at
		 FROM users
		 WHERE LOWER(email) = LOWER($1)`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
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
	ctx context.Context, user model.User,
) (model.User, error) {
	if user.Role == "" {
		user.Role = model.RoleUser
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, password, role, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		user.Email,
		user.Name,
		user.Password,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
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
