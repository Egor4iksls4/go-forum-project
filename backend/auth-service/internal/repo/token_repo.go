package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-forum-project/auth-service/internal/entity"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, userID int, tokenHash string, expireAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type TokenRepo struct {
	Db *sql.DB
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{Db: db}
}

func (r *TokenRepo) CreateRefreshToken(ctx context.Context, userID int, tokenHash string, expireAt time.Time) error {
	_, err := r.Db.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expire_at) VALUES ($1, $2, $3)",
		userID, tokenHash, expireAt,
	)
	return err
}

func (r *TokenRepo) FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.Db.QueryRowContext(ctx,
		"SELECT user_id, token_hash, expire_at FROM refresh_tokens WHERE token_hash = $1",
		tokenHash,
	).Scan(&token.UserID, &token.TokenHash, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	result, err := r.Db.ExecContext(ctx,
		"DELETE FROM refresh_tokens WHERE token_hash = $1",
		tokenHash,
	)
	if err != nil {
		log.Printf("Delete query failed: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}
	return nil
}
