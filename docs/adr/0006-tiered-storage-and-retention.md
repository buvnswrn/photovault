# ADR-0006: Tiered storage (staging → SSD → HDD → offsite) and retention

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0002, ADR-0004, ADR-0015

## Context
The user owns an SSD (fast) and an HDD (large and slow), both plugged in by hand. Copies should follow the 3-2-1 rule: 3 copies, 2 kinds of media, 1 offsite. The SSD and HDD are in the same house, so a single event (fire, theft, ransomware while plugged in) could hit both.

## Decision
| Tier | Location | Role | Default retention |
|---|---|---|---|
| Staging | Hub internal disk | Landing area, not a backup | Deleted 7 days after verified in the SSD repository |
| Primary | SSD restic repository | Recent versioned backups | 14 daily, 8 weekly, 12 monthly |
| Archive | HDD restic repository | Second copy, long-term | 24 monthly, yearly forever |
| Offsite | Cloud restic repository (optional) | Disaster copy | Same as the archive |

- SSD → HDD (and HDD/SSD → offsite) uses `restic copy`, so snapshots stay identical across tiers.
- Retention is applied with `restic forget --prune`, separately for each repository, and only after the copy to the next tier succeeds. This prevents pruning a snapshot that hasn't been copied yet.

## Alternatives considered
- **A single backup drive:** one failure loses everything.
- **Mirroring SSD → HDD with robocopy:** loses versioning on the HDD and copies corruption along.
- **HDD as primary:** slower, and the SSD's speed helps short sessions on a PC that is rarely on.

## Consequences
- The offsite tier is needed for real 3-2-1; until it exists, a single disaster at home can still lose data (tracked in ADR-0015).
- Retention values live in config and are reviewed once real data volumes are known.
