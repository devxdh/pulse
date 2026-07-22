BEGIN;

CREATE TABLE IF NOT EXISTS jobs (
  id SERIAL PRIMARY KEY,
  payload JSONB NOT NULL,
  status NOT NULL DEFAULT 'pending'
    CONSTRAINT chk_ticket_status
    CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
  created_at NOT NULL TIMESTAMPTZ DEFAULT NOW(),
  updated_at NOT NULL TIMESTAMPTZ DEFAULT NOW()
)

COMMIT;
