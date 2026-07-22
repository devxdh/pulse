package main

import (
	"fmt"
	"log"

	"github.com/devxdh/pulse/internal/cfg"
	"github.com/devxdh/pulse/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool := StartDB()
}

func StartDB() *pgxpool.Pool {
	fmt.Println("Starting Pulse API...")

	cfg.LoadEnv()

	dbURL, err := cfg.GetEnv("DATABASE_URL")
	if err != nil {
		log.Fatalf("Startup Failed: %v", err)
	}

	pool, err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = db.InjectDDL(pool)
	if err != nil {
		log.Fatalf("%v", err)
	}

	return pool
}
