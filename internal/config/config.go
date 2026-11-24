package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Staging      StagingConfig      `yaml:"staging"`
	Drives       []DriveConfig      `yaml:"drives"`
	Sources      []SourceConfig     `yaml:"sources"`
	Repositories ReposConfig        `yaml:"repositories"`
	Databases    []DatabaseConfig   `yaml:"databases"`
	Load         LoadConfig         `yaml:"load"`
	SlowMode     SlowModeConfig     `yaml:"slow_mode"`
	Healthchecks HealthchecksConfig `yaml:"healthchecks"`
}

type DriveConfig struct {
	Name   string `yaml:"name"`
	Serial string `yaml:"serial"`
	Role   string `yaml:"role"` // ssd | hdd | offsite
}

type SourceConfig struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type StagingConfig struct {
	Path       string `yaml:"path"`
	QuotaGB    int    `yaml:"quota_gb"`
	RetainDays int    `yaml:"retain_days"`
}

type RetentionConfig struct {
	Daily   int `yaml:"daily"`
	Weekly  int `yaml:"weekly"`
	Monthly int `yaml:"monthly"`
	Yearly  int `yaml:"yearly"` // -1 = keep forever
}

type RepoConfig struct {
	Drive       string          `yaml:"drive"`
	Healthcheck string          `yaml:"healthcheck"`
	Retention   RetentionConfig `yaml:"retention"`
}

type OffsiteRepoConfig struct {
	Enabled     bool            `yaml:"enabled"`
	Healthcheck string          `yaml:"healthcheck"`
	Retention   RetentionConfig `yaml:"retention"`
}

type ReposConfig struct {
	SSD     RepoConfig        `yaml:"ssd"`
	HDD     RepoConfig        `yaml:"hdd"`
	Offsite OffsiteRepoConfig `yaml:"offsite"`
}

type DatabaseConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // postgres | mysql | sqlite | mongo
}

type LoadConfig struct {
	CPUPercent          int  `yaml:"cpu_percent"`
	MemoryPercent       int  `yaml:"memory_percent"`
	DiskActivePercent   int  `yaml:"disk_active_percent"`
	SampleWindowSeconds int  `yaml:"sample_window_seconds"`
	IdleWaitMinutes     int  `yaml:"idle_wait_minutes"`
	AlwaysAsk           bool `yaml:"always_ask"`
}

type SlowModeConfig struct {
	BandwidthLimitKB int    `yaml:"bandwidth_limit_kb"`
	CPUPriority      string `yaml:"cpu_priority"` // low | normal
}

type HealthchecksConfig struct {
	BaseURL string `yaml:"base_url"`
}

var validRoles = map[string]bool{"ssd": true, "hdd": true, "offsite": true}
var validDBTypes = map[string]bool{"postgres": true, "mysql": true, "sqlite": true, "mongo": true}
var validCPUPriorities = map[string]bool{"low": true, "normal": true}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read %s: %w", path, err)
	}
	cfg := defaults()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: cannot parse %s: %w", path, err)
	}
	if errs := cfg.validate(); len(errs) > 0 {
		return nil, fmt.Errorf("config: invalid values:\n  %s", strings.Join(errs, "\n  "))
	}
	return cfg, nil
}

func WriteExample(path string) error {
	return os.WriteFile(path, []byte(exampleYAML), 0o644)
}

// KnownSerials reads the drives list from path without full validation.
// Returns a set of serials already registered.
func KnownSerials(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("config: cannot read %s: %w", path, err)
	}
	var partial struct {
		Drives []DriveConfig `yaml:"drives"`
	}
	if err := yaml.Unmarshal(data, &partial); err != nil {
		return nil, fmt.Errorf("config: cannot parse %s: %w", path, err)
	}
	out := make(map[string]bool, len(partial.Drives))
	for _, d := range partial.Drives {
		if d.Serial != "" {
			out[d.Serial] = true
		}
	}
	return out, nil
}

// AppendDrive adds a drive entry to the drives: list in the YAML file at path,
// preserving all existing content and comments.
func AppendDrive(path string, d DriveConfig) error {
	entry := fmt.Sprintf("  - name: %q\n    serial: %q\n    role: %s\n", d.Name, d.Serial, d.Role)

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("config: cannot read %s: %w", path, err)
		}
		// No file yet — create a minimal one.
		return os.WriteFile(path, []byte("drives:\n"+entry), 0o644)
	}

	// Find insertion point: after the last `role:` line inside the drives: block.
	lines := strings.Split(string(data), "\n")
	insertAt := -1
	inDrives := false
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if t == "drives:" {
			inDrives = true
			insertAt = i + 1 // default: insert right after drives: if list is empty
			continue
		}
		if inDrives {
			if strings.HasPrefix(t, "role:") {
				insertAt = i + 1
			} else if t != "" && !strings.HasPrefix(t, "-") &&
				!strings.HasPrefix(t, "name:") && !strings.HasPrefix(t, "serial:") &&
				!strings.HasPrefix(t, "#") {
				break // left the drives block
			}
		}
	}

	if insertAt < 0 {
		// drives: section not found — just append
		return os.WriteFile(path, append(data, []byte("\ndrives:\n"+entry)...), 0o644)
	}

	var buf strings.Builder
	for i, line := range lines {
		if i == insertAt {
			buf.WriteString(entry)
		}
		buf.WriteString(line)
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}
	return os.WriteFile(path, []byte(buf.String()), 0o644)
}

func defaults() *Config {
	return &Config{
		Staging: StagingConfig{
			QuotaGB:    50,
			RetainDays: 7,
		},
		Repositories: ReposConfig{
			SSD: RepoConfig{
				Healthcheck: "ssd-backup",
				Retention:   RetentionConfig{Daily: 14, Weekly: 8, Monthly: 12},
			},
			HDD: RepoConfig{
				Healthcheck: "hdd-backup",
				Retention:   RetentionConfig{Monthly: 24, Yearly: -1},
			},
			Offsite: OffsiteRepoConfig{
				Enabled:     false,
				Healthcheck: "offsite-backup",
				Retention:   RetentionConfig{Monthly: 24, Yearly: -1},
			},
		},
		Load: LoadConfig{
			CPUPercent:          70,
			MemoryPercent:       85,
			DiskActivePercent:   80,
			SampleWindowSeconds: 30,
			IdleWaitMinutes:     5,
		},
		SlowMode: SlowModeConfig{
			BandwidthLimitKB: 1024,
			CPUPriority:      "low",
		},
		Healthchecks: HealthchecksConfig{
			BaseURL: "https://hc-ping.com",
		},
	}
}

func (c *Config) validate() []string {
	var errs []string

	if c.Staging.Path == "" {
		errs = append(errs, "staging.path: required")
	}
	if c.Staging.QuotaGB <= 0 {
		errs = append(errs, "staging.quota_gb: must be > 0")
	}
	if c.Staging.RetainDays <= 0 {
		errs = append(errs, "staging.retain_days: must be > 0")
	}

	driveNames := map[string]bool{}
	for i, d := range c.Drives {
		if d.Name == "" {
			errs = append(errs, fmt.Sprintf("drives[%d].name: required", i))
		}
		if d.Serial == "" {
			errs = append(errs, fmt.Sprintf("drives[%d].serial: required", i))
		}
		if !validRoles[d.Role] {
			errs = append(errs, fmt.Sprintf("drives[%d].role: must be one of ssd, hdd, offsite (got %q)", i, d.Role))
		}
		if d.Name != "" {
			driveNames[d.Name] = true
		}
	}

	for i, s := range c.Sources {
		if s.Name == "" {
			errs = append(errs, fmt.Sprintf("sources[%d].name: required", i))
		}
		if s.Path == "" {
			errs = append(errs, fmt.Sprintf("sources[%d].path: required", i))
		}
	}

	if c.Repositories.SSD.Drive != "" && !driveNames[c.Repositories.SSD.Drive] {
		errs = append(errs, fmt.Sprintf("repositories.ssd.drive: unknown drive %q", c.Repositories.SSD.Drive))
	}
	if c.Repositories.HDD.Drive != "" && !driveNames[c.Repositories.HDD.Drive] {
		errs = append(errs, fmt.Sprintf("repositories.hdd.drive: unknown drive %q", c.Repositories.HDD.Drive))
	}

	for i, db := range c.Databases {
		if db.Name == "" {
			errs = append(errs, fmt.Sprintf("databases[%d].name: required", i))
		}
		if !validDBTypes[db.Type] {
			errs = append(errs, fmt.Sprintf("databases[%d].type: must be one of postgres, mysql, sqlite, mongo (got %q)", i, db.Type))
		}
	}

	errs = append(errs, validatePercent("load.cpu_percent", c.Load.CPUPercent)...)
	errs = append(errs, validatePercent("load.memory_percent", c.Load.MemoryPercent)...)
	errs = append(errs, validatePercent("load.disk_active_percent", c.Load.DiskActivePercent)...)
	if c.Load.SampleWindowSeconds <= 0 {
		errs = append(errs, "load.sample_window_seconds: must be > 0")
	}
	if c.Load.IdleWaitMinutes <= 0 {
		errs = append(errs, "load.idle_wait_minutes: must be > 0")
	}

	if !validCPUPriorities[c.SlowMode.CPUPriority] {
		errs = append(errs, fmt.Sprintf("slow_mode.cpu_priority: must be low or normal (got %q)", c.SlowMode.CPUPriority))
	}
	if c.SlowMode.BandwidthLimitKB <= 0 {
		errs = append(errs, "slow_mode.bandwidth_limit_kb: must be > 0")
	}

	return errs
}

func validatePercent(field string, v int) []string {
	if v < 1 || v > 100 {
		return []string{fmt.Sprintf("%s: must be 1–100 (got %d)", field, v)}
	}
	return nil
}

const exampleYAML = `# photovault configuration
# Secrets (restic passwords, Healthchecks ping URLs, API tokens) are stored
# in Windows Credential Manager / the system keyring — never here.

staging:
  path: "C:/photovault/staging"
  quota_gb: 50
  retain_days: 7       # remove staged files this many days after SSD verification

drives:
  - name: "My SSD"
    serial: "REPLACE_WITH_SERIAL"
    role: ssd
  - name: "My HDD"
    serial: "REPLACE_WITH_SERIAL"
    role: hdd

sources:
  - name: "photos"
    path: "C:/Users/YourName/Pictures"

repositories:
  ssd:
    drive: "My SSD"
    healthcheck: "ssd-backup"   # Healthchecks.io check name; ping URL in Credential Manager
    retention:
      daily: 14
      weekly: 8
      monthly: 12
      yearly: 0
  hdd:
    drive: "My HDD"
    healthcheck: "hdd-backup"
    retention:
      daily: 0
      weekly: 0
      monthly: 24
      yearly: -1    # -1 = keep forever
  offsite:
    enabled: false
    healthcheck: "offsite-backup"
    retention:
      daily: 0
      weekly: 0
      monthly: 24
      yearly: -1

databases: []
# databases:
#   - name: "myapp"
#     type: postgres    # postgres | mysql | sqlite | mongo
#     # connection string stored in Credential Manager

load:
  cpu_percent: 70
  memory_percent: 85
  disk_active_percent: 80
  sample_window_seconds: 30
  idle_wait_minutes: 5
  always_ask: false

slow_mode:
  bandwidth_limit_kb: 1024
  cpu_priority: low     # low | normal

healthchecks:
  base_url: "https://hc-ping.com"
`
