package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := WriteExample(path); err != nil {
		t.Fatal(err)
	}
	// Example has REPLACE_WITH_SERIAL placeholders; patch them for a valid load.
	raw, _ := os.ReadFile(path)
	patched := strings.ReplaceAll(string(raw), "REPLACE_WITH_SERIAL", "SN12345678")
	_ = os.WriteFile(path, []byte(patched), 0o644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("example config failed to load: %v", err)
	}
	if cfg.Load.CPUPercent != 70 {
		t.Errorf("load.cpu_percent: want 70, got %d", cfg.Load.CPUPercent)
	}
}

func TestValidateMissingFields(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing staging path",
			yaml:    "staging:\n  quota_gb: 10\n  retain_days: 7\n",
			wantErr: "staging.path: required",
		},
		{
			name:    "invalid drive role",
			yaml:    "staging:\n  path: /tmp\n  quota_gb: 10\n  retain_days: 7\ndrives:\n  - name: d\n    serial: s\n    role: flash\n",
			wantErr: "drives[0].role",
		},
		{
			name:    "unknown repo drive reference",
			yaml:    "staging:\n  path: /tmp\n  quota_gb: 10\n  retain_days: 7\nrepositories:\n  ssd:\n    drive: \"ghost\"\n",
			wantErr: "repositories.ssd.drive: unknown drive",
		},
		{
			name:    "invalid db type",
			yaml:    "staging:\n  path: /tmp\n  quota_gb: 10\n  retain_days: 7\ndatabases:\n  - name: db\n    type: redis\n",
			wantErr: "databases[0].type",
		},
		{
			name:    "cpu percent out of range",
			yaml:    "staging:\n  path: /tmp\n  quota_gb: 10\n  retain_days: 7\nload:\n  cpu_percent: 110\n  memory_percent: 85\n  disk_active_percent: 80\n  sample_window_seconds: 30\n  idle_wait_minutes: 5\n",
			wantErr: "load.cpu_percent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			_ = os.WriteFile(path, []byte(tc.yaml), 0o644)
			_, err := Load(path)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	cfg := defaults()
	if cfg.Load.CPUPercent != 70 {
		t.Errorf("default cpu_percent: want 70, got %d", cfg.Load.CPUPercent)
	}
	if cfg.Repositories.HDD.Retention.Yearly != -1 {
		t.Errorf("default hdd yearly retention: want -1, got %d", cfg.Repositories.HDD.Retention.Yearly)
	}
}
