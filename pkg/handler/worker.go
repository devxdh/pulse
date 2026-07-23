package apihandler

import (
	"context"
	"log"
	"time"
)

func (e *Env) StartWorkerPool(numWorkers int) {
	for i := 1; i <= numWorkers; i++ {
		workerID := i

		go func(id int) {
			log.Printf("[Worker %02d] Started and waiting for jobs...", id)

			for jobID := range e.JobQueue {
				e.processJob(id, jobID)
			}
		}(workerID)
	}
}

func (e *Env) processJob(workerID int, jobID int64) {
	query := `
		UPDATE jobs
		SET
			status = 'processing',
			updated_at = NOW()
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.DB.Exec(ctx, query, jobID)
	if err != nil {
		log.Printf("[Worker %02d] [DB ERROR] Failed to update Job #%d to PROCESSING: %v\n", workerID, jobID, err)
		return
	}

	time.Sleep(10 * time.Second)

	query = `
		UPDATE jobs
		SET
			status = 'completed',
			updated_at = NOW()
		WHERE id = $1
	`

	ctxComplete, cancelComplete := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelComplete()

	_, err = e.DB.Exec(ctxComplete, query, jobID)
	if err != nil {
		log.Printf("[Worker %02d] [DB ERROR] failed to Job #%d to COMPLETED: %v\n", workerID, jobID, err)
		return
	}

	log.Printf("[Worker %02d] Successfully completed Job #%d.\n", workerID, jobID)
}
