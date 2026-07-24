package apihandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/devxdh/pulse/internal/cfg"
	"github.com/devxdh/pulse/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreateJob(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	env := New(pool, 10)

	payload := []byte(`{"payload":"test_data"}`)
	req := httptest.NewRequest("POST", "/api/job", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	env.CreateJob(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp struct {
		ID      int             `json:"id"`
		Status  string          `json:"status"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.ID != 1 {
		t.Errorf("Expected job ID 1, got %d", resp.ID)
	}

	if resp.Status != "pending" && resp.Status != "PENDING" {
		t.Errorf("Expected status 'pending' got %q", resp.Status)
	}

	select {
	case queuedID := <-env.JobQueue:
		if queuedID != int64(resp.ID) {
			t.Errorf("Expected channel to receive ID %d, got %d", resp.ID, queuedID)
		}
	default:
		t.Error("Expected job ID to be pushed to JobQueue, but the channel was empty")
	}

}

func TestGetJobByID(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	env := New(pool, 10)

	payload := struct {
		Name     string `json:"name"`
		Task     string `json:"task"`
		Quantity int    `json:"quantity"`
	}{
		Name:     "payload_1",
		Task:     "test payload 1",
		Quantity: 1,
	}

	var seedJob Job

	jsonPayload, _ := json.Marshal(payload)

	query := `
		INSERT INTO jobs (payload)
		VALUES ($1)
		RETURNING id, status, payload, created_at, updated_at
	`

	err := pool.QueryRow(context.Background(), query, jsonPayload).Scan(
		&seedJob.ID,
		&seedJob.Status,
		&seedJob.Payload,
		&seedJob.CreatedAt,
		&seedJob.UpdatedAt,
	)

	if err != nil {
		t.Fatalf("failed to seed job in DB: %v", err)
	}

	path := fmt.Sprintf("/api/job/%d", seedJob.ID)
	req := httptest.NewRequest("GET", path, nil)
	req.SetPathValue("id", fmt.Sprintf("%d", seedJob.ID))

	rr := httptest.NewRecorder()

	env.GetJobByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d.", http.StatusOK, rr.Code)
	}

	var responseJob Job
	if err := json.NewDecoder(rr.Body).Decode(&responseJob); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if responseJob.ID != seedJob.ID {
		t.Errorf("Expected job ID %d, got %d", seedJob.ID, responseJob.ID)
	}

	if responseJob.Status != seedJob.Status {
		t.Errorf("Expected job status %q, got %q", seedJob.Status, responseJob.Status)
	}

	var expectedMap, actualMap map[string]any

	if err := json.Unmarshal(seedJob.Payload, &expectedMap); err != nil {
		t.Fatalf("failed to unmarshal expected payload: %v", err)
	}

	if err := json.Unmarshal(responseJob.Payload, &actualMap); err != nil {
		t.Fatalf("failed to unmarshal actual payload: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Expected payload object %v, got %v", expectedMap, actualMap)
	}

	if !responseJob.CreatedAt.Equal(seedJob.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", seedJob.CreatedAt, responseJob.CreatedAt)
	}

	if !responseJob.UpdatedAt.Equal(seedJob.UpdatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", seedJob.UpdatedAt, responseJob.UpdatedAt)
	}
}

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	cfg.LoadEnv()
	dbURL, err := cfg.GetEnv("TEST_DATABASE_URL")
	if err != nil {
		t.Fatalf("[DB_TEST_HELPER] failed to get env: %v", err)
	}

	pool, err := db.InitDB(dbURL)
	if err != nil {
		t.Fatalf("[DB_TEST_HELPER] failed to connect to db: %v", err)
	}

	if err = db.InjectDDL(pool); err != nil {
		t.Fatalf("[DB] failed to Inject DDL: %v", err)
	}

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE jobs RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("[DB_TEST_HELPER] failed to clean up db: %v", err)
	}

	return pool
}
