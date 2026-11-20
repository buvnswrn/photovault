# ADR-0015: Cloud: overflow and offsite are separate features

- **Status:** Proposed (deferred; not in the first version)
- **Date:** 2026-11-20
- **Related:** ADR-0006, ADR-0014, [blueprint v0 §4B](../ideas/photovault-blueprint-v0.md)

## Context
Blueprint v0 proposed "cloud overflow" with a hard GB quota (`max_cloud_gb`). The later plan proposed an "offsite copy" for the 3-2-1 rule. These are different needs that were easy to confuse.

## Decision
Treat them as two separate features:
| | **Overflow** (from blueprint v0) | **Offsite** |
|---|---|---|
| Purpose | Hold new files **while the hub is offline** or the phone is full | **Disaster copy** if the house loses both SSD and HDD |
| Data | Recent, not-yet-ingested files, temporary | Full restic snapshots, long-term |
| Direction | Phone → cloud → (later) hub, then deleted from the cloud | Hub → cloud via `restic copy` |
| Limits | Hard GB cap; only tagged or manual folders | Retention policy, cost cap |
| Encryption | Client-side before upload | restic repository encryption |

Both are **deferred** (LTR-1 offsite, LTR-2 overflow). Offsite has the higher priority because it closes the 3-2-1 gap.

## Alternatives considered
- **One cloud feature for both:** mixes temporary and archival data, and confuses retention and quota logic.
- **No cloud at all:** acceptable for now; leaves the single-site risk in ADR-0006.

## Consequences
- The provider choice (Backblaze B2, S3, Google Drive) is an open question.
- Overflow needs phone-side upload and encryption; that depends on the outcome of the phone-client spike (ING-5).
