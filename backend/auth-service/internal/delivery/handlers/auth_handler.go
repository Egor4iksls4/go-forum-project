package handlers

import (
	"context"
	"log"

	"go-forum-project/auth-service/internal/delivery/gRPC"
	"go-forum-project/auth-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	grpc.UnimplementedAuthServiceServer
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(ctx context.Context, req *grpc.RegisterRequest) (*grpc.RegisterResponse, error) {
	err := h.uc.Register(req.Username, req.Password, req.Role)
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
	err := h.uc.Logout(req.RefreshToken)
	if err != nil {
		log.Printf("Failed logout: %v", err)
		return &grpc.LogoutResponse{
			Success: false,
		}, status.Error(codes.Internal, "Failed logout")
	}

	return &grpc.LogoutResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *grpc.RefreshRequest) (*grpc.TokenResponse, error) {
	tokens, err := h.uc.RefreshTokens(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &grpc.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
