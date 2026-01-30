package idempotency

import (
	"database/sql"
	"time"

	"github.com/project-atlas/atlas/pkg/models"
	_ "modernc.org/sqlite"
)

type Ledger struct {
	db *sql.DB
}

func NewLedger(dbPath string) (*Ledger, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS idempotency_ledger (
		request_id TEXT PRIMARY KEY,
		checksum TEXT NOT NULL,
		status TEXT NOT NULL,
		resource_id TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_checksum ON idempotency_ledger(checksum);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Ledger{db: db}, nil
}

func (l *Ledger) RecordPending(requestID, checksum string) error {
	now := time.Now()
	_, err := l.db.Exec(`
		INSERT INTO idempotency_ledger (request_id, checksum, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		requestID, checksum, models.StatusPending, now, now)
	return err
}

func (l *Ledger) Complete(requestID, resourceID string) error {
	_, err := l.db.Exec(`
		UPDATE idempotency_ledger 
		SET status = ?, resource_id = ?, updated_at = ?
		WHERE request_id = ?`,
		models.StatusCompleted, resourceID, time.Now(), requestID)
	return err
}

func (l *Ledger) GetPendingTasks() ([]models.ActionRecord, error) {
	rows, err := l.db.Query(`
		SELECT request_id, checksum, status, resource_id, created_at, updated_at
		FROM idempotency_ledger WHERE status = ?`, models.StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.ActionRecord
	for rows.Next() {
		var rec models.ActionRecord
		if err := rows.Scan(&rec.RequestID, &rec.Checksum, &rec.Status, &rec.ResourceID, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, rec)
	}
	return tasks, nil
}

func (l *Ledger) GetByChecksum(checksum string) (*models.ActionRecord, error) {
	var rec models.ActionRecord
	err := l.db.QueryRow(`
		SELECT request_id, checksum, status, resource_id, created_at, updated_at
		FROM idempotency_ledger WHERE checksum = ?`, checksum).
		Scan(&rec.RequestID, &rec.Checksum, &rec.Status, &rec.ResourceID, &rec.CreatedAt, &rec.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}
