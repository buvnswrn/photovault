<!-- Original blueprint (Idea A), kept verbatim. See ../ROADMAP.md for the merged plan. -->

# Project Blueprint: `photovault` (Cross-Platform Personal Photo Backup System)

## 1. Overview & Vision
`photovault` is a lightweight, self-hosted, single-binary backup tool designed to automatically sync photos from an Android phone (Samsung S24 Ultra) and secondary laptops to a primary PC's local storage (HDDs/SSDs), with optional quota-capped cloud overflow. 

Maintained as a versioned package in GitHub, it can be deployed onto any new machine in seconds without runtime dependencies.

---

## 2. Technology Stack
* **Core Language:** **Go (Golang)** — Compiles to standalone, zero-dependency binaries (`linux/amd64`, `windows/amd64`, `android/arm64`), ideal for AI-agent-assisted development and cross-compilation.
* **Networking:** **Local LAN + Tailscale (Peer-to-Peer WireGuard Mesh)** — Provides static private IPs (`100.x.y.z`) and MagicDNS hostnames (e.g., `http://my-desktop-pc:8080`) so devices communicate securely over home Wi-Fi or cellular data without opening router ports.
* **Local State Tracking:** **SQLite (`modernc.org/sqlite`)** — Pure-Go embedded database to track SHA-256 file hashes and prevent duplicate transfers.
* **Automation & Triggers:**
  * **Windows / Linux:** `kardianos/service` (background daemon), `fsnotify` (real-time folder watching), `robfig/cron` (scheduled checks), `gen2brain/beeep` (OS notifications).
  * **Android (S24 Ultra):** **Termux** (runs the native `android/arm64` Go binary) + **MacroDroid / Tasker** (triggers syncs on new camera photos, Wi-Fi connection, or daily schedules).
* **Distribution:** **GitHub Releases + GoReleaser** — Automatically compiles and publishes binaries for all platforms on every git tag.

---

## 3. System Architecture

```text
+-----------------------+         LAN / Tailscale Mesh         +-------------------------+

|  Clients (Spokes)     | -----------------------------------> |  Primary PC (Hub)       |
| - Samsung S24 Ultra   |   1. Ping / Health Check             | - Runs `photovault serve`|
| - Windows/Linux Laptop|   2. Stream Chunked Files + SHA-256  | - Verifies HDD/SSD Mount|
+-----------------------+                                      | - Writes to Local Drive |
           |                                                   +-------------------------+

           | (If Hub PC is Offline)                                         |
           v                                                                | (Optional Overflow)
+-----------------------+                                                   v

| Smart Fallback Logic  |                                      +-------------------------+
| - Check Snooze/Sched  |                                      | Cloud Storage (S3/Drive)|
| - Notify: "Turn on PC"|                                      | - Hard GB Quota Cap     |
| - Queue in SQLite     |                                      | - Manual/Tagged Only    |
+-----------------------+                                      +-------------------------+
```

---

## 4. Core Features & Workflow

### A. Hub Mode (`photovault serve`)
* Runs continuously as a background service (`systemd` on Linux, Windows Service on Windows).
* **Mount Safety Guard:** Before accepting or writing files, verifies that the target external HDD/SSD path (e.g., `D:\PhotoBackup` or `/mnt/backup_hdd`) is actively mounted so it never accidentally fills up the OS boot drive.
* **Integrity Verification:** Validates incoming SHA-256 checksums before committing files to disk and organizes photos into `YYYY/MM/` directories.

### B. Client / Watcher Mode (`photovault watch` & `photovault sync`)
* **Event-Driven & Scheduled Triggers:** Watches target folders (`DCIM/Camera` on Android, local photo folders on laptops) in real time and runs periodic scheduled sweeps.
* **Offline Hub Detection & Snooze:**
  * Pings the Hub PC via LAN or Tailscale MagicDNS.
  * If unreachable, checks `notify_offline` settings and active snooze timers (`photovault snooze 4h`).
  * Triggers a non-intrusive system notification (*"Backup PC is offline — turn it on to sync 24 new photos"*) unless snoozed or outside configured active hours.
* **Cloud Overflow Guard:** Optional secondary cloud target (S3, Backblaze B2, Google Drive) restricted by a strict storage ceiling (`max_cloud_gb`) or explicit folder tags to prevent accidental quota exhaustion.

### C. Tailscale Capabilities (Current & Future Scope)
* **Seamless Roaming:** Uses the same MagicDNS hostname whether on home Wi-Fi or traveling.
* **Node Online/Offline Monitoring:** Can optionally embed `://tailscale.com` or use Tailscale Webhooks to automatically trigger a backup the exact moment the Hub PC powers on.
* **Wake-on-LAN (WoL):** Can pair with an always-on home node (router/Raspberry Pi) to remotely wake the Hub PC before initiating a scheduled backup.

---

## 5. Repository Structure

```text
photovault/
├── .github/
│   └── workflows/
│       └── release.yml          # GitHub Actions workflow running GoReleaser on tags
├── .goreleaser.yaml             # Cross-compilation config (Linux, Windows, Android ARM64)
├── cmd/
│   └── photovault/
│       └── main.go              # CLI entrypoint (Cobra commands)
├── internal/
│   ├── config/
│   │   └── config.go            # YAML configuration & snooze state management
│   ├── server/
│   │   └── server.go            # HTTP API, mount verification, SHA-256 validation
│   ├── client/
│   │   ├── scanner.go           # Directory walker & SQLite deduplication tracker
│   │   └── uploader.go          # Chunked HTTP uploader & offline retry queue
│   ├── watcher/
│   │   └── watcher.go           # fsnotify real-time listener & cron scheduler
│   ├── notify/
│   │   └── notify.go            # Desktop (beeep) & Termux notification wrapper
│   └── daemon/
│       └── service.go           # Windows Service & Linux systemd installer
├── scripts/
│   ├── install.sh               # One-line Linux / macOS / Termux installer
│   └── install.ps1              # One-line Windows PowerShell installer
├── config.example.yaml          # Sample configuration template
├── go.mod
└── README.md
```

---

## 6. CLI Command Interface

* `photovault init` — Generates default `config.yaml` and initializes the local SQLite tracking database.
* `photovault serve` — Starts the Hub HTTP server to receive and store backups.
* `photovault sync` — Performs an immediate one-off sync check (used by Android MacroDroid triggers).
* `photovault watch` — Runs continuously in Client mode (file watcher + cron scheduler).
* `photovault snooze <duration>` — Suppresses "PC Offline" notifications for a set period (e.g., `2h`, `1d`) to align with calendar schedules.
* `photovault status` — Displays pending files, last successful sync timestamp, and Hub reachability.
* `photovault install-service` — Registers `photovault` as an auto-starting background service on Windows or Linux.

---

## 7. AI Agent Starter Prompt

Copy and paste the prompt below into your AI coding assistant (Cursor, Claude Code, GitHub Copilot) to scaffold the initial repository:

> **Prompt:**
> "Initialize a Go module named `://github.com` following the exact directory layout above. Use `spf13/cobra` for the CLI, `modernc.org/sqlite` for tracking synced file hashes, `fsnotify/fsnotify` for folder watching, `robfig/cron/v3` for scheduling, `gen2brain/beeep` for desktop notifications, and `kardianos/service` for background daemon registration.
>
> Implement:
> 1. `serve` mode: An HTTP server with `/health` and `/upload` endpoints that verifies the destination drive mount exists before writing, checks SHA-256 hashes, and saves files into `YYYY/MM/` subfolders.
> 2. `sync` mode: Checks `/health` on the configured Hub URL (supporting LAN IPs and Tailscale MagicDNS hostnames). If offline, checks if notifications are enabled and not currently snoozed in `config.yaml`; if active, triggers a system notification and exits gracefully. If online, scans configured source directories, skips files already recorded in SQLite, and uploads new files concurrently using a worker pool.
> 3. A `.goreleaser.yaml` file configured with `CGO_ENABLED=0` to cross-compile standalone binaries for `linux/amd64`, `windows/amd64`, and `android/arm64`."

---

## 8. Setup Checklist for New Devices

1. **Primary PC (Windows/Linux):**
   * Download latest binary from GitHub Releases (or run `install.sh` / `install.ps1`).
   * Configure external HDD/SSD path in `config.yaml`.
   * Run `photovault install-service --mode serve`.
2. **Samsung S24 Ultra (Android):**
   * Install **Termux** (from F-Droid/GitHub) and **MacroDroid**.
   * Set Battery usage to **Unrestricted** for both Termux and MacroDroid in One UI Settings.
   * Download the `android/arm64` binary into Termux (`~/bin/photovault`) and run `photovault init`.
   * Create a MacroDroid rule to execute `photovault sync` when a new photo is added to `/storage/emulated/0/DCIM/Camera` or on a daily schedule.
