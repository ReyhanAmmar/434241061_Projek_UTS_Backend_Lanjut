package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-buku-kas/app/model"
)

type TokenRepository interface {
	Save(ctx context.Context, token model.RefreshToken) error

	FindActive(
		ctx context.Context,
		tokenHash string,
	) (model.RefreshToken, error)

	Rotate(
		ctx context.Context,
		oldHash string,
		next model.RefreshToken,
	) error

	Revoke(ctx context.Context, tokenHash string) error
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func (r *tokenPostgresRepository) Save(
	ctx context.Context,
	token model.RefreshToken,
) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("menyimpan refresh token: %w", err)
	}

	return nil
}

func (r *tokenPostgresRepository) FindActive(
	ctx context.Context,
	tokenHash string,
) (model.RefreshToken, error) {
	var token model.RefreshToken

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at,
		        revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		   AND expires_at > NOW()`,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, fmt.Errorf(
			"mengambil refresh token: %w", err,
		)
	}

	return token, nil
}

func (r *tokenPostgresRepository) Rotate(
	ctx context.Context,
	oldHash string,
	next model.RefreshToken,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai rotasi refresh token: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE token_hash = $1
		   AND user_id = $2
		   AND revoked_at IS NULL
		   AND expires_at > NOW()`,
		oldHash,
		next.UserID,
	)
	if err != nil {
		return fmt.Errorf("mencabut token lama: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		next.UserID,
		next.TokenHash,
		next.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("menyimpan token pengganti: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("menyelesaikan rotasi refresh token: %w", err)
	}

	return nil
}

func (r *tokenPostgresRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE token_hash = $1
		   AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}

	return nil
}