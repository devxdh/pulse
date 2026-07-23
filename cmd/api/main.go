// Package main application's entrypoint.
package main

import (
	"fmt"
	"log"

	"github.com/devxdh/pulse/internal/cfg"
	"github.com/devxdh/pulse/pkg/db"
	apihandler "github.com/devxdh/pulse/pkg/handler"
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

	apiEnv := apihandler.New(pool, 100)
	apiEnv.StartWorkerPool(4)

	app := &application{
		db:  pool,
		api: apiEnv,
	}

	app.startServer()
}
