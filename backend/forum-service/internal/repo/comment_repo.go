package repo

import (
	"context"
	"database/sql"
	"errors"
	"go-forum-project/forum-service/internal/entity"
	"time"
)

type CommentRepository interface {
	CreateComm(ctx context.Context, postId int, content, author string) error
	GetByPostID(ctx context.Context, postID int) ([]entity.Comment, error)
	GetCommentByID(ctx context.Context, id int) (entity.Comment, error)
	Delete(ctx context.Context, postID int) error
}

type CommentRepo struct {
	db *sql.DB
}

func NewCommentRepo(db *sql.DB) CommentRepository {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComm(ctx context.Context, postId int, content, author string) error {
	query := `INSERT INTO comments (post_id, content, author, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, postId, content, author, time.Now())
	return err
}

func (r *CommentRepo) GetByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	query := `
        SELECT id, post_id, content, author, created_at 
        FROM comments 
        WHERE post_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []entity.Comment
	for rows.Next() {
		var c entity.Comment
		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.Content,
			&c.Author,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (r *CommentRepo) GetCommentByID(ctx context.Context, id int) (entity.Comment, error) {
	query := `
		SELECT id, post_id, content, author, created_at 
		FROM comments 
		WHERE id = $1
	`

	var c entity.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.PostID,
		&c.Content,
		&c.Author,
		&c.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Comment{}, err
		}
		return entity.Comment{}, err
	}

	return c, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
