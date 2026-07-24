package apihandler

import (
	"context"
	"testing"
	"time"
)

func TestIntegration_StartWorkerPool_Deterministic(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	ctx := context.Background()

	_, _ = pool.Exec(ctx, "DELETE FROM jobs;")

	var seedJobID int64 = 555
	query := `
		INSERT INTO jobs (id, payload, status, created_at, updated_at)
		VALUES ($1, '{"test": true}', 'pending', NOW(), NOW());
	`
	_, err := pool.Exec(ctx, query, seedJobID)
	if err != nil {
		t.Fatalf("Failed to seed a test job row: %v", err)
	}

	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM jobs WHERE id = $1", seedJobID) }()

	env := &Env{
		DB:       pool,
		JobQueue: make(chan int64, 1),
	}

	env.StartWorkerPool(1)
	env.JobQueue <- seedJobID
	close(env.JobQueue)

	var finalStatus string
	success := false
	checkQuery := "SELECT status FROM jobs WHERE id = $1;"

	for range 50 {
		err = pool.QueryRow(ctx, checkQuery, seedJobID).Scan(&finalStatus)
		if err != nil {
			t.Fatalf("Failed to fetch state updates from DB: %v", err)
		}

		if finalStatus == "processing" || finalStatus == "completed" {
			success = true
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !success {
		t.Errorf("Worker failed to process job! Status remained stuck on: %s", finalStatus)
	}
}
