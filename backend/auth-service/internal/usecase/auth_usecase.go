package usecase

import (
	"errors"
	"go-forum-project/auth-service/internal/entity"
	"go-forum-project/auth-service/internal/repo"
)

type AuthUseCase struct {
	Repo repo.AuthRepository
}

func (uc *AuthUseCase) Login(username, password string) (*entity.User, error) {
	user, err := uc.Repo.GetUserByUsername(username)
	if err != nil || user.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}
