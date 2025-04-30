package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-forum-project/auth-service/internal/entity"
	"go-forum-project/auth-service/internal/repo"
)

type AuthUseCase struct {
	userRepo  repo.AuthRepository
	tokenRepo repo.TokenRepository
	secretKey string
}

func NewAuthUseCase(ur repo.AuthRepository, tr repo.TokenRepository, secretKey string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  ur,
		tokenRepo: tr,
		secretKey: secretKey,
	}
}

func (uc *AuthUseCase) generateAccessToken(user *entity.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(uc.secretKey))
	return signedToken, expiresAt, err
}

func (uc *AuthUseCase) generateRefreshToken(user *entity.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	rawToken := sha256.Sum256([]byte(user.Username + time.Now().String() + uc.secretKey))
	token := hex.EncodeToString(rawToken[:])

	if err := uc.tokenRepo.CreateRefreshToken(context.Background(), user.ID, token, expiresAt); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (uc *AuthUseCase) Login(username, password string) (*entity.TokenPair, error) {
	user, err := uc.userRepo.GetUserByUsername(username)
	if err != nil || user.Password != password {
		return nil, errors.New("invalid credentials")
	}

	accessToken, accessExp, err := uc.generateAccessToken(&user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExp, err := uc.generateRefreshToken(&user)
	if err != nil {
		return nil, err
	}

	return &entity.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (uc *AuthUseCase) RefreshTokens(refreshToken string) (*entity.TokenPair, error) {
	storedToken, err := uc.tokenRepo.FindRefreshToken(context.Background(), refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := uc.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := uc.tokenRepo.DeleteRefreshToken(context.Background(), refreshToken); err != nil {
		return nil, err
	}

	return uc.Login(user.Username, user.Password)
}

func (uc *AuthUseCase) Logout(refreshToken string) error {
	return uc.tokenRepo.DeleteRefreshToken(context.Background(), refreshToken)
}
