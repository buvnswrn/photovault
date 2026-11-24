package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection and exposes domain-level accessors.
type DB struct {
	conn *sql.DB
}

// File is a known file tracked by SHA-256.
type File struct {
	ID          int64
	SHA256      string
	Path        string
	Size        int64
	Source      string
	FirstSeenAt time.Time
	StagedAt    *time.Time
	PromotedAt  *time.Time
}

// Job is a single backup run, identified by what triggered it.
type Job struct {
	ID          int64
	Trigger     string // agent_start | drive_arrival | phone_request | manual
	DriveSerial string
	StartedAt   time.Time
	FinishedAt  *time.Time
	Status      string // running | done | failed | interrupted
}

// Step is one stage within a Job; its status persists across process restarts.
type Step struct {
	ID         int64
	JobID      int64
	Name       string
	Status     string // pending | running | done | failed
	StartedAt  *time.Time
	FinishedAt *time.Time
	Detail     string
}

// TierSuccess records the last verified backup for a storage tier.
type TierSuccess struct {
	Tier          string
	LastSuccessAt time.Time
	SnapshotID    string
}

var migrations = []string{
	`CREATE TABLE files (
		id           INTEGER PRIMARY KEY,
		sha256       TEXT    NOT NULL UNIQUE,
		path         TEXT    NOT NULL,
		size         INTEGER NOT NULL,
		source       TEXT    NOT NULL,
		first_seen_at TEXT   NOT NULL,
		staged_at    TEXT,
		promoted_at  TEXT
	)`,
	`CREATE TABLE jobs (
		id           INTEGER PRIMARY KEY,
		trigger      TEXT    NOT NULL,
		drive_serial TEXT    NOT NULL DEFAULT '',
		started_at   TEXT    NOT NULL,
		finished_at  TEXT,
		status       TEXT    NOT NULL DEFAULT 'running'
	)`,
	`CREATE TABLE steps (
		id          INTEGER PRIMARY KEY,
		job_id      INTEGER NOT NULL REFERENCES jobs(id),
		name        TEXT    NOT NULL,
		status      TEXT    NOT NULL DEFAULT 'pending',
		started_at  TEXT,
		finished_at TEXT,
		detail      TEXT    NOT NULL DEFAULT '',
		UNIQUE(job_id, name)
	)`,
	`CREATE TABLE tier_success (
		tier            TEXT PRIMARY KEY,
		last_success_at TEXT NOT NULL,
		snapshot_id     TEXT NOT NULL DEFAULT ''
	)`,
}

// Open opens (or creates) the SQLite database at path and runs any pending migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	conn.SetMaxOpenConns(1) // SQLite: single writer

	if _, err := conn.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}
	if _, err := conn.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("db: enable foreign keys: %w", err)
	}
	if _, err := conn.Exec(`PRAGMA busy_timeout=5000`); err != nil {
		return nil, fmt.Errorf("db: set busy timeout: %w", err)
	}

	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error { return d.conn.Close() }

func (d *DB) migrate() error {
	if _, err := d.conn.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT    NOT NULL
	)`); err != nil {
		return fmt.Errorf("db: bootstrap schema_version: %w", err)
	}

	var current int
	if err := d.conn.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current); err != nil {
		return fmt.Errorf("db: read schema version: %w", err)
	}

	for i, ddl := range migrations {
		v := i + 1
		if v <= current {
			continue
		}
		tx, err := d.conn.Begin()
		if err != nil {
			return fmt.Errorf("db: migration %d: begin: %w", v, err)
		}
		if _, err := tx.Exec(ddl); err != nil {
			tx.Rollback()
			return fmt.Errorf("db: migration %d: %w", v, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_version (version, applied_at) VALUES (?, ?)`, v, now()); err != nil {
			tx.Rollback()
			return fmt.Errorf("db: migration %d: record version: %w", v, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("db: migration %d: commit: %w", v, err)
		}
	}
	return nil
}

// UpsertFile inserts a file record or ignores if the SHA-256 already exists.
func (d *DB) UpsertFile(f *File) error {
	_, err := d.conn.Exec(
		`INSERT INTO files (sha256, path, size, source, first_seen_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(sha256) DO NOTHING`,
		f.SHA256, f.Path, f.Size, f.Source, now(),
	)
	return err
}

// FileByHash returns the file with the given SHA-256, or nil if not found.
func (d *DB) FileByHash(sha256 string) (*File, error) {
	row := d.conn.QueryRow(
		`SELECT id, sha256, path, size, source, first_seen_at, staged_at, promoted_at FROM files WHERE sha256 = ?`,
		sha256,
	)
	f := &File{}
	var firstSeen string
	var stagedAt, promotedAt sql.NullString
	if err := row.Scan(&f.ID, &f.SHA256, &f.Path, &f.Size, &f.Source, &firstSeen, &stagedAt, &promotedAt); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	f.FirstSeenAt, _ = time.Parse(time.RFC3339, firstSeen)
	if stagedAt.Valid {
		t, _ := time.Parse(time.RFC3339, stagedAt.String)
		f.StagedAt = &t
	}
	if promotedAt.Valid {
		t, _ := time.Parse(time.RFC3339, promotedAt.String)
		f.PromotedAt = &t
	}
	return f, nil
}

// MarkStaged records when a file was written to the staging area.
func (d *DB) MarkStaged(sha256 string) error {
	_, err := d.conn.Exec(`UPDATE files SET staged_at = ? WHERE sha256 = ?`, now(), sha256)
	return err
}

// MarkPromoted records when a file was verified in the SSD repository.
func (d *DB) MarkPromoted(sha256 string) error {
	_, err := d.conn.Exec(`UPDATE files SET promoted_at = ? WHERE sha256 = ?`, now(), sha256)
	return err
}

// CreateJob opens a new job record and returns its ID.
func (d *DB) CreateJob(trigger, driveSerial string) (int64, error) {
	res, err := d.conn.Exec(
		`INSERT INTO jobs (trigger, drive_serial, started_at, status) VALUES (?, ?, ?, 'running')`,
		trigger, driveSerial, now(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FinishJob sets the job's terminal status and finished_at time.
func (d *DB) FinishJob(id int64, status string) error {
	_, err := d.conn.Exec(
		`UPDATE jobs SET status = ?, finished_at = ? WHERE id = ?`,
		status, now(), id,
	)
	return err
}

// UpsertStep creates or updates a step within a job.
func (d *DB) UpsertStep(jobID int64, name, status, detail string) error {
	var startedAt, finishedAt interface{}
	if status == "running" {
		startedAt = now()
	}
	if status == "done" || status == "failed" {
		finishedAt = now()
	}
	_, err := d.conn.Exec(
		`INSERT INTO steps (job_id, name, status, started_at, finished_at, detail)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(job_id, name) DO UPDATE SET
		   status = excluded.status,
		   started_at = COALESCE(steps.started_at, excluded.started_at),
		   finished_at = excluded.finished_at,
		   detail = excluded.detail`,
		jobID, name, status, startedAt, finishedAt, detail,
	)
	return err
}

// InterruptedJob returns the most recent job still in 'running' state (interrupted by a
// process kill) along with its steps, so the caller can resume where it left off.
// Returns nil, nil, nil when no interrupted job exists.
func (d *DB) InterruptedJob() (*Job, []Step, error) {
	row := d.conn.QueryRow(
		`SELECT id, trigger, drive_serial, started_at FROM jobs WHERE status = 'running' ORDER BY id DESC LIMIT 1`,
	)
	j := &Job{Status: "running"}
	var startedAt string
	if err := row.Scan(&j.ID, &j.Trigger, &j.DriveSerial, &startedAt); err == sql.ErrNoRows {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	j.StartedAt, _ = time.Parse(time.RFC3339, startedAt)

	rows, err := d.conn.Query(
		`SELECT id, job_id, name, status, COALESCE(started_at,''), COALESCE(finished_at,''), detail
		 FROM steps WHERE job_id = ? ORDER BY id`,
		j.ID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var steps []Step
	for rows.Next() {
		var s Step
		var sa, fa string
		if err := rows.Scan(&s.ID, &s.JobID, &s.Name, &s.Status, &sa, &fa, &s.Detail); err != nil {
			return nil, nil, err
		}
		if sa != "" {
			t, _ := time.Parse(time.RFC3339, sa)
			s.StartedAt = &t
		}
		if fa != "" {
			t, _ := time.Parse(time.RFC3339, fa)
			s.FinishedAt = &t
		}
		steps = append(steps, s)
	}
	return j, steps, rows.Err()
}

// SetTierSuccess records a successful backup for the given tier.
func (d *DB) SetTierSuccess(tier, snapshotID string) error {
	_, err := d.conn.Exec(
		`INSERT INTO tier_success (tier, last_success_at, snapshot_id) VALUES (?, ?, ?)
		 ON CONFLICT(tier) DO UPDATE SET last_success_at = excluded.last_success_at, snapshot_id = excluded.snapshot_id`,
		tier, now(), snapshotID,
	)
	return err
}

// TierSuccesses returns the last verified backup time for each tier.
func (d *DB) TierSuccesses() (map[string]TierSuccess, error) {
	rows, err := d.conn.Query(`SELECT tier, last_success_at, snapshot_id FROM tier_success`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]TierSuccess{}
	for rows.Next() {
		var ts TierSuccess
		var lastAt string
		if err := rows.Scan(&ts.Tier, &lastAt, &ts.SnapshotID); err != nil {
			return nil, err
		}
		ts.LastSuccessAt, _ = time.Parse(time.RFC3339, lastAt)
		out[ts.Tier] = ts
	}
	return out, rows.Err()
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
