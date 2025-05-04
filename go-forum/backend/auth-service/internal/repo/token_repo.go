package repo

import (
	"context"
	"database/sql"
	"go-forum-project/auth-service/internal/entity"
	"time"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, userID int, tokenHash string, expireAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type TokenRepo struct {
	db *sql.DB
}

func (r *TokenRepo) CreateRefreshToken(ctx context.Context, userID int, tokenHash string, expireAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expire_at) VALUES ($1, $2, $3)",
		userID, tokenHash, expireAt,
	)
	return err
}

func (r *TokenRepo) FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.QueryRowContext(ctx,
		"SELECT user_id, token_hash, expire_at FROM refresh_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&token.UserID, &token.TokenHash, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM refresh_tokens WHERE token_hash = $1",
		tokenHash)
	return err
}
