package app

import (
	"context"
	"database/sql"
	"fmt"
	"go-forum-project/chat-service/internal/client"
	"go-forum-project/chat-service/internal/config"
	"go-forum-project/chat-service/internal/delivery/handler"
	"go-forum-project/chat-service/internal/repo"
	"go-forum-project/chat-service/internal/usecase"
	"log"
	"net/http"
	"time"
)

func RunChat() {
	cfg, err := config.LoadConfig("chat-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	messageRepo := repo.NewMessageRepo(db)
	messageUC := usecase.NewMessageUseCase(messageRepo)

	authClient, err := client.NewAuthClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}
	defer authClient.Close()

	hub := handler.NewHub(messageUC)
	go hub.Run()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			if err := messageUC.CleanupOldMessages(context.Background()); err != nil {
				log.Printf("Failed to cleanup old messages: %v", err)
			}
		}
	}()

	http.Handle("/ws", enableCORS(handler.ServeWs(hub, authClient)))
	http.Handle("/api/messages", enableCORS(handler.GetMessageHandler(messageUC)))

	log.Printf("Server started on : %d", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := []string{"http://localhost:3000"}
		origin := r.Header.Get("Origin")

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Refresh-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
