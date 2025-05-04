package grpcapp

import (
	"fmt"
	"log"
	"net"

	pb "go-forum-project/auth-service/internal/delivery/gRPC"
	"go-forum-project/auth-service/internal/delivery/handlers"
	"go-forum-project/auth-service/internal/usecase"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
}

func NewGRPCApp(port int, authUC *usecase.AuthUseCase) *App {
	gRPCServer := grpc.NewServer()

	authHandler := handlers.NewAuthHandler(authUC)
	pb.RegisterAuthServiceServer(gRPCServer, authHandler)

	return &App{
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (app *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.port))
	if err != nil {
		return fmt.Errorf("failed to listen %w", err)
	}

	log.Printf("Start gRPC server on port %d", app.port)

	if err := app.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to serve %w", err)
	}

	return nil
}

func (app *App) Stop() {
	app.gRPCServer.GracefulStop()
}
