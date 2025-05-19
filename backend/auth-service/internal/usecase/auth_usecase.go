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
	"go-forum-project/auth-service/internal/config"
	"go-forum-project/auth-service/internal/entity"
	"go-forum-project/auth-service/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Login(username, password string) (*entity.TokenPair, error)
	Logout(ctx context.Context) error
	Register(username, password string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
	ValidateToken(ctx context.Context, accessToken string) (string, bool, error)
}

type authUseCase struct {
	userRepo  repo.AuthRepository
	tokenRepo repo.TokenRepository
	secretKey string
}

func NewAuthUseCase(ur repo.AuthRepository, tr repo.TokenRepository, secretKey string) AuthUseCase {
	return &authUseCase{
		userRepo:  ur,
		tokenRepo: tr,
		secretKey: secretKey,
	}
}

func (uc *authUseCase) generateAccessToken(user *entity.User) (string, time.Time, error) {
	cfg, err := config.LoadConfig("auth-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	expiresAt := time.Now().Add(cfg.Security.AccessTokenTTL)

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"role":     user.Role,
		"username": user.Username,
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(uc.secretKey))
	return signedToken, expiresAt, err
}

func (uc *authUseCase) generateRefreshToken(user *entity.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, errors.New("user cannot be nil")
	}

	if user.ID == 0 {
		return "", time.Time{}, fmt.Errorf("invalid user ID: %d", user.ID)
	}

	if user.Username == "" {
		return "", time.Time{}, errors.New("username cannot be empty")
	}

	cfg, err := config.LoadConfig("auth-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	expiresAt := time.Now().Add(cfg.Security.RefreshTokenTTL)

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

func (uc *authUseCase) Login(username, password string) (*entity.TokenPair, error) {
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

func (uc *authUseCase) Logout(ctx context.Context) error {
	refreshToken, ok := ctx.Value("refreshToken").(string)
	if !ok || refreshToken == "" {
		return errors.New("refresh token not found in context")
	}

	if _, err := uc.tokenRepo.FindRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("refresh token not found: %w", err)
	}

	log.Printf("Deleting refresh token: %s", refreshToken)
	if err := uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (uc *authUseCase) Register(username, password string) error {
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
		Role:     "user",
	}

	return uc.userRepo.CreateUser(user)
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	storedToken, err := uc.tokenRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	// 2. Проверяем срок действия
	if time.Now().After(storedToken.ExpiresAt) {
		_ = uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// 3. Получаем пользователя
	user, err := uc.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		_ = uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, errors.New("user not found")
	}

	// 4. Генерируем новые токены
	newAccessToken, accessExp, err := uc.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshExp, err := uc.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// 5. Удаляем старый токен
	if err := uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		log.Printf("Failed to delete old refresh token: %v", err)
	}

	return &entity.TokenPair{
		AccessToken:      newAccessToken,
		RefreshToken:     newRefreshToken,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (uc *authUseCase) ValidateToken(ctx context.Context, accessToken string) (string, bool, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.secretKey), nil
	})
	if err != nil {
		return "", false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false, errors.New("invalid token claims")
	}

	// Проверяем username в claims
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return "", false, errors.New("username not found in token")
	}

	return username, true, nil
}
