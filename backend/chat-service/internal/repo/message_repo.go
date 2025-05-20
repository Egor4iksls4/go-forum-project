package repo

import (
	"context"
	"database/sql"
	"go-forum-project/chat-service/internal/entity"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, author, text string) error
	DeleteMessage(ctx context.Context, id int) error
	GetAllMessages(ctx context.Context) ([]*entity.Message, error)
}

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) MessageRepository {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) CreateMessage(ctx context.Context, author, text string) error {
	query := `INSERT INTO messages (author, text) VALUES ($1, $2)`
	_, err := r.db.ExecContext(
		ctx,
		query,
		author,
		text,
	)
	return err
}

func (r *MessageRepo) DeleteMessage(ctx context.Context, id int) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MessageRepo) GetAllMessages(ctx context.Context) ([]*entity.Message, error) {
	query := `SELECT id, author, text, created_at FROM messages ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*entity.Message
	for rows.Next() {
		message := &entity.Message{}
		err := rows.Scan(
			&message.ID,
			&message.Author,
			&message.Text,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}
