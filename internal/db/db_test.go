package db

import (
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	d, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestOpenCreatesSchema(t *testing.T) {
	d := openTestDB(t)
	tables := []string{"schema_version", "files", "jobs", "steps", "tier_success"}
	for _, tbl := range tables {
		var name string
		err := d.conn.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tbl).Scan(&name)
		if err != nil {
			t.Errorf("table %q missing: %v", tbl, err)
		}
	}
}

func TestMigrationIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	d, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()

	d2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	d2.Close()
}

func TestJobResume(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	// First "process": create a job and some steps, then exit without finishing.
	func() {
		d, _ := Open(path)
		defer d.Close()

		jobID, err := d.CreateJob("drive_arrival", "SN12345")
		if err != nil {
			t.Fatal(err)
		}
		if err := d.UpsertStep(jobID, "ingest", "done", ""); err != nil {
			t.Fatal(err)
		}
		if err := d.UpsertStep(jobID, "ssd_backup", "running", ""); err != nil {
			t.Fatal(err)
		}
		// process "killed" here — FinishJob never called
	}()

	// Second "process": should find the interrupted job.
	d, _ := Open(path)
	defer d.Close()

	job, steps, err := d.InterruptedJob()
	if err != nil {
		t.Fatal(err)
	}
	if job == nil {
		t.Fatal("expected interrupted job, got nil")
	}
	if job.Trigger != "drive_arrival" {
		t.Errorf("trigger: want drive_arrival, got %q", job.Trigger)
	}
	if len(steps) != 2 {
		t.Fatalf("steps: want 2, got %d", len(steps))
	}
	if steps[0].Status != "done" {
		t.Errorf("step[0] status: want done, got %q", steps[0].Status)
	}
	if steps[1].Status != "running" {
		t.Errorf("step[1] status: want running, got %q", steps[1].Status)
	}
}

func TestFileDedup(t *testing.T) {
	d := openTestDB(t)
	f := &File{SHA256: "abc123", Path: "/photos/a.jpg", Size: 1024, Source: "photos"}
	if err := d.UpsertFile(f); err != nil {
		t.Fatal(err)
	}
	if err := d.UpsertFile(f); err != nil {
		t.Fatal("second upsert should not error:", err)
	}

	var count int
	d.conn.QueryRow(`SELECT COUNT(*) FROM files WHERE sha256 = 'abc123'`).Scan(&count)
	if count != 1 {
		t.Errorf("want 1 row, got %d", count)
	}
}

func TestFileByHash(t *testing.T) {
	d := openTestDB(t)
	f := &File{SHA256: "deadbeef", Path: "/photos/b.jpg", Size: 2048, Source: "photos"}
	if err := d.UpsertFile(f); err != nil {
		t.Fatal(err)
	}
	got, err := d.FileByHash("deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected file, got nil")
	}
	if got.Path != "/photos/b.jpg" {
		t.Errorf("path: want /photos/b.jpg, got %q", got.Path)
	}

	missing, err := d.FileByHash("nothere")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for missing hash")
	}
}

func TestTierSuccess(t *testing.T) {
	d := openTestDB(t)
	if err := d.SetTierSuccess("ssd", "snap-001"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetTierSuccess("hdd", "snap-002"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetTierSuccess("ssd", "snap-003"); err != nil { // update
		t.Fatal(err)
	}

	tiers, err := d.TierSuccesses()
	if err != nil {
		t.Fatal(err)
	}
	if tiers["ssd"].SnapshotID != "snap-003" {
		t.Errorf("ssd snapshot: want snap-003, got %q", tiers["ssd"].SnapshotID)
	}
	if tiers["hdd"].SnapshotID != "snap-002" {
		t.Errorf("hdd snapshot: want snap-002, got %q", tiers["hdd"].SnapshotID)
	}
}

func TestNoInterruptedJob(t *testing.T) {
	d := openTestDB(t)
	job, steps, err := d.InterruptedJob()
	if err != nil {
		t.Fatal(err)
	}
	if job != nil || steps != nil {
		t.Error("expected nil job and steps on empty DB")
	}
}
