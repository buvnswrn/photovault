# Architecture Decision Records

This folder records the important design decisions for photovault: what we decided, why, and which alternatives we rejected. The point is that ideas don't get lost and settled questions don't get re-argued.

Format: [Michael Nygard's ADR template](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions), extended with an "Alternatives considered" section. Copy [`template.md`](template.md) for new records.

## Rules
- One decision per ADR, numbered in sequence and never renumbered.
- Once an ADR is **Accepted**, don't rewrite its decision. Write a new ADR that **supersedes** it, and update the old one's status line.
- Statuses: `Proposed` → `Accepted` | `Rejected` | `Deprecated` | `Superseded by ADR-NNNN`.
- **Only the project owner moves an ADR to `Accepted`.** Proposals (including AI-drafted ones) stay `Proposed` until explicitly reviewed (see issue Q-6).

## Index

| ADR | Title | Status |
|---|---|---|
| [0000](0000-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0001](0001-scope-merge-ingest-and-protection.md) | Scope: one system for ingest and protection | Proposed |
| [0002](0002-go-program-with-restic.md) | Go for the program, restic for backup repositories | Proposed |
| [0003](0003-hub-runs-as-logon-tray-process.md) | Hub runs as a logon-time tray process, not a Windows Service | Proposed |
| [0004](0004-staging-first-ingest.md) | Staging-first ingest with a quota | Proposed |
| [0005](0005-identify-drives-by-serial.md) | Identify backup drives by serial number | Proposed |
| [0006](0006-tiered-storage-and-retention.md) | Tiered storage (staging → SSD → HDD → offsite) and retention | Proposed |
| [0007](0007-end-to-end-integrity-verification.md) | End-to-end integrity verification | Proposed |
| [0008](0008-append-only-no-deletion-propagation.md) | Append-only: device deletions never propagate | Proposed |
| [0009](0009-event-driven-triggers.md) | Event-driven triggers instead of fixed schedules | Proposed |
| [0010](0010-phone-connectivity-wired-and-wireless.md) | Phone connectivity: wired and wireless | Accepted |
| [0011](0011-request-consent-and-load-aware-gating.md) | Request → consent protocol with load-aware gating | Proposed |
| [0012](0012-monitoring-and-reminders.md) | Monitoring and reminders | Proposed |
| [0013](0013-database-backups-via-native-dumps.md) | Database backups via native dump tools | Proposed |
| [0014](0014-encryption-and-secrets.md) | Encryption and secret storage | Proposed |
| [0015](0015-cloud-overflow-vs-offsite.md) | Cloud: overflow and offsite are separate features | Proposed (deferred) |
| [0016](0016-deferred-remote-wake-and-roaming.md) | Deferred: Tailscale roaming, Wake-on-LAN, node-online triggers | Proposed (deferred) |
