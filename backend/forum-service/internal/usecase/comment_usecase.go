package usecase

import (
	"context"
	"errors"
	"fmt"
	"go-forum-project/forum-service/internal/entity"
	"go-forum-project/forum-service/internal/repo"
)

var (
	ErrLengthComment = errors.New("content must be between 1 and 200 characters")
	ErrPostNotFound  = errors.New("post not found")
)

type CommentUseCase interface {
	Create(ctx context.Context, postID int, content, author string) error
	GetByPostID(ctx context.Context, postID int) ([]entity.Comment, error)
	DeleteComment(ctx context.Context, commentID int, currentUsername string) error
}

type commentUseCase struct {
	commentRepo repo.CommentRepository
	postRepo    repo.PostRepository
}

func NewCommentUseCase(cr repo.CommentRepository, pr repo.PostRepository) CommentUseCase {
	return &commentUseCase{
		commentRepo: cr,
		postRepo:    pr,
	}
}

func (c *commentUseCase) Create(ctx context.Context, postID int, content, author string) error {
	if len(content) == 0 || len(content) > 200 {
		return ErrLengthComment
	}

	if _, err := c.postRepo.GetPostByID(ctx, postID); err != nil {
		return ErrPostNotFound
	}

	err := c.commentRepo.CreateComm(ctx, postID, content, author)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}

	return nil
}

func (c *commentUseCase) GetByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	return c.commentRepo.GetByPostID(ctx, postID)
}

func (c *commentUseCase) DeleteComment(ctx context.Context, commentID int, currentUsername string) error {
	return c.commentRepo.Delete(ctx, commentID)
}
