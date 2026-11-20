# ADR-0004: Staging-first ingest with a quota

- **Status:** Proposed
- **Date:** 2026-11-20
- **Related:** ADR-0005, ADR-0006, ADR-0007

## Context
Blueprint v0 writes incoming files **directly to the external drive**, and a "mount safety guard" refuses uploads if the drive isn't mounted. But the SSD and HDD are plugged in by hand and are usually absent, so most uploads would be refused. Separately, the original plan wanted "some space in between" on the desktop while copying.

## Decision
All ingested data goes first into a **staging area on the hub's internal disk** (`<root>/staging/`):
- `inbox/` → (hash, dedupe) → `phone/YYYY/MM/…`, `db/<date>/…`, each batch with a SHA-256 manifest.
- Photos are organized by **EXIF capture date**, falling back to file modification time.
- Staging has a **size quota** (`staging.max_gb`) and a **free-space floor** for the system drive. Requests that would exceed either are refused or deferred, with a notification.
- Data leaves staging only after it is verified in the SSD repository (ADR-0007), and is then kept for a grace period (default 7 days) before deletion.

The blueprint's mount guard is redefined as: "staging has room" for ingest, plus "the target drive is matched by serial" (ADR-0005) for promotion to SSD/HDD.

## Alternatives considered
- **Direct-to-drive writes (blueprint v0):** fails under the plugged-in-by-hand constraint.
- **Leave data on the phone until the drives are present:** forces the phone, the PC and both drives to be available at the same time; too restrictive.

## Consequences
- The internal disk needs spare space for staging; the quota protects the OS drive.
- Staging is a single, unversioned copy until it reaches the SSD, so the agent must push it to the SSD at the next opportunity and remind the user (ADR-0012).
