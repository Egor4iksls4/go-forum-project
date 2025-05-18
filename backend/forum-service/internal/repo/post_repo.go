package repo

import (
	"context"
	"database/sql"
	"time"

	"go-forum-project/forum-service/internal/entity"
)

type PostRepository interface {
	CreatePost(ctx context.Context, title, content, author string) error
	UpdatePost(ctx context.Context, id int, title, content string) error
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	DeletePost(ctx context.Context, id int) error
}

type PostRepo struct {
	DB *sql.DB
}

func NewPostRepo(db *sql.DB) PostRepository {
	return &PostRepo{DB: db}
}

func (r *PostRepo) CreatePost(ctx context.Context, title, content, author string) error {
	query := "INSERT INTO posts (title, content, author) VALUES ($1, $2, $3)"
	_, err := r.DB.ExecContext(
		ctx,
		query,
		title, content, author,
	)
	return err
}

func (r *PostRepo) UpdatePost(ctx context.Context, id int, title, content string) error {
	query := "UPDATE posts SET title = $1, content = $2, updated_at = $3 WHERE id = $4"
	_, err := r.DB.ExecContext(
		ctx,
		query,
		title, content, time.Now(), id,
	)

	return err
}

func (r *PostRepo) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	query := "SELECT id, title, content, author, created_at, updated_at FROM posts ORDER BY created_at DESC"

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entity.Post
	for rows.Next() {
		p := &entity.Post{}
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.Author,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostRepo) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	query := "SELECT id, title, content, author, created_at, updated_at FROM posts WHERE id = $1"

	var p entity.Post
	err := r.DB.QueryRowContext(ctx,
		query,
		id,
	).Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PostRepo) DeletePost(ctx context.Context, id int) error {
	query := "DELETE FROM posts WHERE id = $1"
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}
