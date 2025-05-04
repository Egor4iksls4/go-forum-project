package app

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-forum-project/auth-service/cmd/app/grpcapp"
	"go-forum-project/auth-service/internal/repo"
	"go-forum-project/auth-service/internal/usecase"
)

type App struct {
	GRPCApp *grpcapp.App
}

func NewApp(db *sql.DB) *App {
	userRepo := repo.NewUserRepo(db)
	tokenRepo := repo.NewTokenRepo(db)

	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, "secret-key")

	gRPCApp := grpcapp.NewGRPCApp(50051, authUC)

	return &App{GRPCApp: gRPCApp}
}

func (app *App) Run() {
	go func() {
		if err := app.GRPCApp.Run(); err != nil {
			log.Fatalf("Failed to start GRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	app.GRPCApp.Stop()
	log.Println("Shutting down gracefully")
}
