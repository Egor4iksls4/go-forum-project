package app

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"go-forum-project/auth-service/cmd/app/grpcapp"
	"go-forum-project/auth-service/internal/config"
	"go-forum-project/auth-service/internal/repo"
	"go-forum-project/auth-service/internal/usecase"
)

type App struct {
	GRPCApp *grpcapp.App
}

func NewApp(cfg *config.Config) *App {
	db, err := sql.Open("postgres", cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	userRepo := repo.NewUserRepo(db)
	tokenRepo := repo.NewTokenRepo(db)

	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, cfg.Security.SecretKey)

	gRPCApp := grpcapp.NewGRPCApp(cfg.Server.GRPCPort, authUC)

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
