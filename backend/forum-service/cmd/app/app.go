package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-forum-project/forum-service/internal/client"
	"go-forum-project/forum-service/internal/config"
	"go-forum-project/forum-service/internal/delivery/http/router"
	"go-forum-project/forum-service/internal/middleware"
	"go-forum-project/forum-service/internal/repo"
	"go-forum-project/forum-service/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunForumApp() {
	cfg, err := config.LoadConfig("forum-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authClient, err := client.NewAuthClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to create auth client: %w", err)
	}
	defer authClient.Close()

	postRepo := repo.NewPostRepo(db)
	commentRepo := repo.NewCommentRepo(db)

	postUseCase := usecase.NewPostUseCase(postRepo)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo)

	authMiddleware := middleware.AuthMiddleware(authClient)

	r := router.NewRouter(postUseCase, commentUseCase, authMiddleware)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan error, 1)
	go func() {
		log.Printf("Listening on port: %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			done <- fmt.Errorf("failed to listen on port %d: %v", cfg.Server.Port, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-done:
		log.Fatalf("Failed to done: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}

	log.Println("Server gracefully stopped")
}
