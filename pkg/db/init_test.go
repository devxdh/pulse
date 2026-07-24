package db

import (
	"context"
	"testing"

	"github.com/devxdh/pulse/internal/cfg"
)

func TestDB_Init_and_InjectDDL(t *testing.T) {
	cfg.LoadEnv()
	dbURL, err := cfg.GetEnv("TEST_DATABASE_URL")
	if err != nil {
		t.Fatalf("failed to get db url env: %v", err)
	}

	ctx := context.Background()

	pool, err := InitDB(dbURL)
	if err != nil {
		t.Fatalf("InitDB failed to connect: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS jobs CASCADE")
	if err != nil {
		t.Fatalf("failed to drop residue tables: %v", err)
	}

	err = InjectDDL(pool)
	if err != nil {
		t.Fatalf("InjectDDL failed to inject: %v", err)
	}

	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS jobs CASCADE")
}
