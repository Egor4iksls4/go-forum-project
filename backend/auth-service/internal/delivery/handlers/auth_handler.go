package handlers

import (
	"context"
	"go-forum-project/auth-service/internal/usecase"
	grpc "go-forum-project/proto/gRPC"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type AuthHandler struct {
	grpc.UnimplementedAuthServiceServer
	uc usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(ctx context.Context, req *grpc.RegisterRequest) (*grpc.RegisterResponse, error) {
	err := h.uc.Register(req.Username, req.Password)
	if err != nil {
		log.Printf("Failed register: %v", err)
		return &grpc.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.AlreadyExists, "Failed register")
	}
	return &grpc.RegisterResponse{Success: true}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *grpc.LoginRequest) (*grpc.TokenResponse, error) {
	tokens, err := h.uc.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("User not found: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &grpc.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *grpc.LogoutRequest) (*grpc.LogoutResponse, error) {
	refreshToken := req.RefreshToken

	if refreshToken == "" {
		return &grpc.LogoutResponse{Success: false}, status.Error(codes.Unauthenticated, "refresh token required")
	}

	newCtx := context.WithValue(ctx, "refreshToken", refreshToken)
	if err := h.uc.Logout(newCtx); err != nil {
		return &grpc.LogoutResponse{Success: false}, status.Error(codes.Internal, "logout failed")
	}

	return &grpc.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *grpc.RefreshRequest) (*grpc.TokenResponse, error) {
	newTokens, err := h.uc.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "refresh failed")
	}

	return &grpc.TokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *grpc.ValidateTokenRequest) (*grpc.ValidateTokenResponse,
	error) {
	username, isValid, err := h.uc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if isValid {
		return &grpc.ValidateTokenResponse{
			Username: username,
			Valid:    true,
		}, nil
	}

	return &grpc.ValidateTokenResponse{
		Valid: false,
	}, nil
}
