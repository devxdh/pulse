// Package main application's entrypoint.
package main

import (
	"fmt"
	"log"

	"github.com/devxdh/pulse/internal/cfg"
	"github.com/devxdh/pulse/pkg/db"
)

func main() {
	fmt.Println("Starting Pulse API...")

	cfg.LoadEnv()
	dbURL, err := cfg.GetEnv("DATABASE_URL")
	if err != nil {
		log.Fatalf("Configuration Error: %v", err)
	}

	pool, err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer pool.Close()

	db.InjectDDL(pool)

	app := &application{
		db: pool,
	}

	app.startServer()
}
