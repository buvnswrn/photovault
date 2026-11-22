# ADR-0002: Go for the program, restic for backup repositories

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0001, ADR-0006, ADR-0014

## Context
The system needs: an HTTP API for phone and laptop clients, a phone-side client, a tray process with interactive notifications, load monitoring, a state database, and versioned, deduplicated, encrypted backups. It must run on Windows (hub) and Android (client), with Linux as a nice-to-have.

## Decision
- We will write photovault in **Go**: single static binaries, easy cross-compilation (windows/amd64, linux/amd64, android/arm64), and pure-Go SQLite (`modernc.org/sqlite`).
- We will use **restic** as an *external binary* for the backup repositories (backup, copy, check, forget/prune, restore), rather than writing our own backup format.

## Alternatives considered
- **PowerShell scripts + restic** (from the first Backup Agent plan): fastest to start, but Windows-only. It doesn't fit an HTTP API, a phone client, or prompts with buttons. Rejected once those requirements arrived.
- **Python:** needs a runtime on every machine, and packaging for Android is poor.
- **Our own backup format** (mirror into `YYYY/MM/` only, as in blueprint v0): no versioning, encryption or dedupe across snapshots, and no protection against deletion or ransomware.
- **Kopia** instead of restic: a viable alternative. restic was chosen for `restic copy` between repositories (SSD → HDD) and its simple CLI. We may revisit this.
- **Embedding restic as a Go library:** restic doesn't provide a stable library API.

## Consequences
- restic must be installed or bundled alongside photovault; the release pipeline should fetch a pinned version.
- photovault parses restic's `--json` output, so we pin and test against a specific restic version.
