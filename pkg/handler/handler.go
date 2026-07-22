// Package apihandler contains all handler functions
package apihandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Env struct {
	DB *pgxpool.Pool
}

type Job struct {
	ID        int64           `json:"id"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func New(db *pgxpool.Pool) *Env {
	return &Env{DB: db}
}

func (e *Env) HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, World!\n")
}

func (e *Env) CreateJob(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Payload json.RawMessage `json:"payload"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid json payload", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received payload: %s\n", string(req.Payload))

	query := `
		INSERT INTO jobs (payload)
		VALUES($1)
		RETURNING ID, status
	`

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var resID int
	var resStatus string

	err = e.DB.QueryRow(ctx, query, req.Payload).Scan(&resID, &resStatus)
	if err != nil {
		log.Printf("[DB ERROR] Database insert failed: %v", err)
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(struct {
		ID      int             `json:"id"`
		Status  string          `json:"status"`
		Payload json.RawMessage `json:"payload"`
	}{
		ID:      resID,
		Status:  resStatus,
		Payload: req.Payload,
	})
}

func (e *Env) GetJobByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid job ID: id must be an integer", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	query := `
		SELECT id, status, payload, created_at, updated_at
		FROM jobs
		WHERE id=$1
	`

	var job Job
	err = e.DB.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.Status,
		&job.Payload,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		log.Printf("[DB ERROR] job query failed: %v", err)
		http.Error(w, fmt.Sprintf("Job not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}
