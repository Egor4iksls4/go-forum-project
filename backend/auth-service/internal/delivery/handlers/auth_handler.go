package handlers

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"

	pb "go-forum-project/auth-service/internal/delivery/gRPC"
	"go-forum-project/auth-service/internal/usecase"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := h.uc.Register(req.Username, req.Password, req.Role)
	if err != nil {
		log.Printf("Failed register: %v", err)
		return &pb.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.AlreadyExists, "Failed register")
	}
	return &pb.RegisterResponse{Success: true}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.TokenResponse, error) {
	tokens, err := h.uc.Login(req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &pb.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.uc.Logout(req.RefreshToken)
	if err != nil {
		log.Printf("Failed logout: %v", err)
		return &pb.LogoutResponse{
			Success: false,
		}, status.Error(codes.Internal, "Failed logout")
	}

	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.TokenResponse, error) {
	tokens, err := h.uc.RefreshTokens(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &pb.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
