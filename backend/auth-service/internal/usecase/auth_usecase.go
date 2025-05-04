package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-forum-project/auth-service/internal/entity"
	"go-forum-project/auth-service/internal/repo"
	"golang.org/x/crypto/bcrypt"
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

	if user.ID == 0 {
		return "", time.Time{}, errors.New("invalid user ID")
	}

	if err := uc.tokenRepo.CreateRefreshToken(context.Background(), user.ID, token, expiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return token, expiresAt, nil
}

func (uc *AuthUseCase) Login(username, password string) (*entity.TokenPair, error) {
	user, err := uc.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, accessExp, err := uc.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExp, err := uc.generateRefreshToken(user)
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
	tokenHash := sha256.Sum256([]byte(refreshToken))
	log.Printf("Deleting token hash: %s", hex.EncodeToString(tokenHash[:]))
	return uc.tokenRepo.DeleteRefreshToken(context.Background(), hex.EncodeToString(tokenHash[:]))
}

func (uc *AuthUseCase) Register(username, password, role string) error {
	exists, err := uc.userRepo.UserExists(username)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists {
		return errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	return uc.userRepo.CreateUser(user)
}
