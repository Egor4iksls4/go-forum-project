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
	cfg, err := config.LoadConfig("auth-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	expiresAt := time.Now().Add(cfg.Security.AccessTokenTTL)

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

func (uc *AuthUseCase) Logout(ctx context.Context) error {
	refreshToken, ok := ctx.Value("refreshToken").(string)
	if !ok || refreshToken == "" {
		return errors.New("refresh token not found in context")
	}

	log.Printf("Deleting refresh token: %s", refreshToken)
	if err := uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (uc *AuthUseCase) Register(username, password string) error {
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

func (uc *AuthUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	storedToken, err := uc.tokenRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		_ = uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, errors.New("refresh token expired")
	}

	user, err := uc.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return uc.Login(user.Username, user.Password)
}

func (uc *AuthUseCase) ValidateToken(ctx context.Context, accessToken string) (string, bool, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return "", false, errors.New("invalid token format")
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

	// Проверяем expiration вручную
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", false, errors.New("expiration not found in token")
	}

	expirationTime := time.Unix(int64(exp), 0)
	isExpired := time.Now().After(expirationTime)

	return username, !isExpired, nil
}
