package app

import (
	"context"
	"fmt"
	auth "go-forum-project/proto/gRPC"
	"log"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-forum-project/auth-service/internal/config"
)

func RunAuthApp() {
	cfg, err := config.LoadConfig("auth-service/internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		authApp := NewApp(cfg)
		authApp.Run()
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		mux := runtime.NewServeMux()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		err := auth.RegisterAuthServiceHandlerFromEndpoint(
			ctx,
			mux,
			fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort),
			opts,
		)
		if err != nil {
			log.Fatalf("Failed to register gateway: %v", err)
		}

		corsHandler := allowCORS(mux)

		log.Printf("Starting gateway server on: %d", cfg.Server.GRPCGatewayPort)
		if err := http.ListenAndServe(":8080", corsHandler); err != nil {
			log.Fatalf("Failed to start gateway: %v", err)
		}
	}()

	wg.Wait()
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
