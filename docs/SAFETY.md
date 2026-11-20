# Safety invariants

> **Status:** Draft. Derived from the Proposed ADRs; review it together with issue Q-6.

photovault is backup software. Its bugs are silent: you find out only when you try to restore. These rules state what must **never** happen. Every change is checked against them.

## Rules for this document
- Each invariant has **at least one automated test** in the fault-injection suite (issue TST-1). An invariant without a test is a wish.
- An issue or PR that touches an invariant ticks the **"Safety impact"** box in its template and names the invariant (e.g. `S2`).
- Changing or removing an invariant needs a **new ADR**.
- If an invariant can't be guaranteed in some situation, photovault must **stop and tell the user** rather than continue.

## Invariants

| ID | photovault must never… | Source | Fault-injection tests |
|---|---|---|---|
| **S1** | …write to a drive unless it matches a configured **serial number**, uniquely. | [ADR-0005](adr/0005-identify-drives-by-serial.md) | F1, F2 |
| **S2** | …delete a staging batch before its contents have been **restored from the SSD repository and compared** with the manifest, and the grace period has passed. | [ADR-0004](adr/0004-staging-first-ingest.md), [ADR-0007](adr/0007-end-to-end-integrity-verification.md) | F3, F4, F5 |
| **S3** | …prune (`forget --prune`) a snapshot from a tier before it exists, verified, on the **next** tier (when one is configured). | [ADR-0006](adr/0006-tiered-storage-and-retention.md) | F6 |
| **S4** | …delete or overwrite hub data because a file was deleted or changed **on a device** (append-only). | [ADR-0008](adr/0008-append-only-no-deletion-propagation.md) | F7 |
| **S5** | …delete anything **from the phone** automatically. | [ADR-0008](adr/0008-append-only-no-deletion-propagation.md) | F7 |
| **S6** | …let staging exceed its quota, or let the system drive drop below its free-space floor. | [ADR-0004](adr/0004-staging-first-ingest.md) | F8 |
| **S7** | …leave a repository inconsistent after an interruption (kill, shutdown, unplug); `restic check` must pass and the step must resume. | [ADR-0009](adr/0009-event-driven-triggers.md) | F4, F9, F10 |
| **S8** | …report **success** (green tray icon, Healthchecks success ping, "backed up") for a step that wasn't verified. | [ADR-0007](adr/0007-end-to-end-integrity-verification.md), [ADR-0012](adr/0012-monitoring-and-reminders.md) | F3, F5 |
| **S9** | …show **"safe to unplug"** before writes are flushed and the volume is ejected. | PRO-6 | F11 |
| **S10** | …store secrets (repository password, device tokens, ping URLs, upload credentials) in config, logs or git, or send paths, filenames, serials or DB names off the PC. | [ADR-0012](adr/0012-monitoring-and-reminders.md), [ADR-0014](adr/0014-encryption-and-secrets.md) | F12 |
| **S11** | …back up a running database by copying its data files; only native dumps. | [ADR-0013](adr/0013-database-backups-via-native-dumps.md) | (code review) |
| **S12** | …perform a destructive operation (delete, prune, forget, eject) without a `--dry-run` equivalent that shows exactly what would happen. | — | F13 |
| **S13** | …store data in a way that **plain restic can't restore** without photovault ([RECOVERY.md](RECOVERY.md)). | [ADR-0002](adr/0002-go-program-with-restic.md) | TST-2 (recovery drill) |

## Fault-injection scenarios (TST-1)

| ID | Scenario | Expected outcome |
|---|---|---|
| F1 | A pendrive or unknown drive is plugged in, even with the SSD's usual drive letter | Ignored; nothing written |
| F2 | Two drives report the same or a missing serial | Refused, with a notification |
| F3 | A file is corrupted in the repository after backup, before verification | Verification fails; staging kept; no success ping |
| F4 | Process killed between backup and verification | Next run re-verifies; staging kept |
| F5 | Restore-compare finds a missing file | Batch marked failed; staging kept; user notified |
| F6 | HDD unavailable for months while the SSD's retention would expire old snapshots | Snapshots not yet copied to the HDD are kept |
| F7 | A file is deleted or edited on the phone after it was backed up | Hub copy unchanged; edit stored as a new version; phone untouched |
| F8 | An upload would exceed the staging quota / system-drive floor | Refused or deferred with a notification; nothing partially written |
| F9 | The SSD is unplugged during `restic backup` / `restic copy` | `restic check` passes afterwards; the step resumes |
| F10 | Power loss (hard kill) during a staging move | No file lost or duplicated in staging; manifest still consistent |
| F11 | Eject fails because a file is in use | No "safe to unplug"; the user is told which process holds it |
| F12 | Logs, status.json, Healthchecks payloads and config scanned for secrets and paths | None found |
| F13 | Every destructive command run with `--dry-run` | Plan shown; file system unchanged |
