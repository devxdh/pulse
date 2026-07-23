package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devxdh/pulse/internal/middleware"
	apihandler "github.com/devxdh/pulse/pkg/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	db  *pgxpool.Pool
	api *apihandler.Env
}

func (app *application) startServer() {
	mux := http.NewServeMux()

	app.routes(mux)

	rl := middleware.NewRateLimiter(5.0/60.0, 5.0)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        rl.Limit(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server is running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Critical server error: %v", err)
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("HTTP server shut down cleanly!")
}
