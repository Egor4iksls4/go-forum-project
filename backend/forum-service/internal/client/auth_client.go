package client

import (
	"context"
	"errors"
	"fmt"
	"go-forum-project/forum-service/internal/config"
	pb "go-forum-project/proto/gRPC"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(ctx context.Context, cfg *config.Config) (*AuthClient, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		cfg.AuthService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) GetUsername(ctx context.Context, token string) (string, error) {
	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return "", err
	}
	if !resp.Valid {
		return "", errors.New("invalid token")
	}
	return resp.Username, nil
}

func (c *AuthClient) Close() {
	if err := c.conn.Close(); err != nil {
		log.Fatalf("failed to close auth client connection: %v", err)
	}
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, bool, error) {
	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return "", false, fmt.Errorf("validate token error: %w", err)
	}
	return resp.Username, resp.Valid, nil
}

func (c *AuthClient) Refresh(ctx context.Context, refreshToken string) (*pb.TokenResponse, error) {
	return c.client.Refresh(ctx, &pb.RefreshRequest{
		RefreshToken: refreshToken,
	})
}
