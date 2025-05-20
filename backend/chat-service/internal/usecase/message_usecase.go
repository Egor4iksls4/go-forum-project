package usecase

import (
	"context"
	"errors"
	"go-forum-project/chat-service/internal/entity"
	"go-forum-project/chat-service/internal/repo"
	"time"
)

var (
	ErrLengthText = errors.New("text must be between 1 and 150 characters")
)

type MessageUseCase interface {
	CreateMessage(ctx context.Context, author, text string) error
	GetAllMessages(ctx context.Context) ([]*entity.Message, error)
	DeleteMessage(ctx context.Context, id int) error
	CleanupOldMessages(ctx context.Context) error
}

type messageUseCase struct {
	repo repo.MessageRepository
}

func NewMessageUseCase(repo repo.MessageRepository) MessageUseCase {
	return &messageUseCase{repo: repo}
}

func (c *messageUseCase) CreateMessage(ctx context.Context, author, text string) error {
	if len(text) == 0 || len(author) > 150 {
		return ErrLengthText
	}

	if err := c.repo.CreateMessage(ctx, author, text); err != nil {
		return err
	}

	return nil
}

func (c *messageUseCase) GetAllMessages(ctx context.Context) ([]*entity.Message, error) {
	return c.repo.GetAllMessages(ctx)
}

func (c *messageUseCase) DeleteMessage(ctx context.Context, id int) error {
	return c.repo.DeleteMessage(ctx, id)
}

func (c *messageUseCase) CleanupOldMessages(ctx context.Context) error {
	messages, err := c.repo.GetAllMessages(ctx)
	if err != nil {
		return err
	}

	for _, message := range messages {
		if !message.CreatedAt.IsZero() && time.Since(message.CreatedAt) >= time.Hour*24 {
			if err := c.repo.DeleteMessage(ctx, message.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
