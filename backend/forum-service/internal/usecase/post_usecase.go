package usecase

import (
	"context"
	"errors"
	"go-forum-project/forum-service/internal/entity"
	"go-forum-project/forum-service/internal/repo"
)

type PostUseCase interface {
	CreatePost(ctx context.Context, title, content, author string) error
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	GetPostById(ctx context.Context, id int) (*entity.Post, error)
	UpdatePost(ctx context.Context, id int, title, content string) error
	DeletePost(ctx context.Context, id int) error
}

type postUseCase struct {
	repo repo.PostRepository
}

func NewPostUseCase(repo repo.PostRepository) PostUseCase {
	return &postUseCase{repo: repo}
}

func (uc *postUseCase) CreatePost(ctx context.Context, title, content, author string) error {
	if len(title) == 0 || len(title) > 100 {
		return errors.New("title must be between 1 and 100 characters")
	}
	if len(content) == 0 || len(content) > 250 {
		return errors.New("content must be between 1 and 250 characters")
	}

	if err := uc.repo.CreatePost(ctx, title, content, author); err != nil {
		return err
	}

	return nil
}

func (uc *postUseCase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	return uc.repo.GetAllPosts(ctx)
}

func (uc *postUseCase) GetPostById(ctx context.Context, id int) (*entity.Post, error) {
	return uc.repo.GetPostByID(ctx, id)
}

func (uc *postUseCase) UpdatePost(ctx context.Context, id int, title, content string) error {
	if len(title) == 0 || len(title) > 100 {
		return errors.New("title must be between 1 and 100 characters")
	}
	if len(content) == 0 || len(content) > 250 {
		return errors.New("content must be between 1 and 250 characters")
	}

	return uc.repo.UpdatePost(ctx, id, title, content)
}

func (uc *postUseCase) DeletePost(ctx context.Context, id int) error {
	return uc.repo.DeletePost(ctx, id)
}
