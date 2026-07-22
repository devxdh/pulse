// Package db handles Database init and DDL injection
package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/001-init.sql
var initSchema string

func InitDB(connStr string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("[DB] Failed to connect to database: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("[DB] error while pinging database: %w", err)
	}

	return pool, nil
}

func InjectDDL(pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("[DB] Injecting DDL...")

	_, err := pool.Exec(ctx, initSchema)
	if err != nil {
		log.Fatalf("[DB] failed to Inject DDL: %v", err)
	}

	fmt.Println("[DB] DDL Injection succeeded!")
}
