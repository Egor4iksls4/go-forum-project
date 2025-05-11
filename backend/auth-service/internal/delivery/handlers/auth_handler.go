package handlers

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net/http"
	"strings"
	"time"

	grpc "go-forum-project/auth-service/internal/delivery/gRPC"
	"go-forum-project/auth-service/internal/usecase"
	g "google.golang.org/grpc"
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

func (h *AuthHandler) Refresh(ctx context.Context, _ *emptypb.Empty) (*grpc.TokenResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not found")
	}

	cookies := md.Get("cookie")
	var refreshToken string

	for _, c := range cookies {
		if strings.HasPrefix(c, "refresh_token=") {
			parts := strings.SplitN(c, "=", 2)
			if len(parts) == 2 {
				refreshToken = parts[1]
				break
			}
		}
	}

	if refreshToken == "" {
		return nil, status.Error(codes.Unauthenticated, "refresh token not provided")
	}

	tokens, err := h.uc.RefreshTokens(refreshToken)
	if err != nil {
		log.Printf("Refresh failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	if _, ok := metadata.FromIncomingContext(ctx); ok {
		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    tokens.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(time.Until(tokens.RefreshExpiresAt).Seconds()),
		}

		header := metadata.Pairs("Set-Cookie", cookie.String())
		if err := g.SetHeader(ctx, header); err != nil {
			log.Printf("Failed to set gRPC header: %v", err)
		}
	}

	return &grpc.TokenResponse{
		AccessToken: tokens.AccessToken,
	}, nil
}
