package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	db *pgxpool.Pool
}

func (app *application) startServer() {
	mux := http.NewServeMux()

	app.routes(mux)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
