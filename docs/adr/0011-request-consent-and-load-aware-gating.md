# ADR-0011: Request → consent protocol with load-aware gating

- **Status:** Proposed
- **Date:** 2026-11-20
- **Related:** ADR-0003, ADR-0009, ADR-0010

## Context
The phone (or a plugged-in drive) can start backups at any time. If the PC is already busy (CPU, memory or disk), a large backup makes everything slow, and the user may prefer to close something first, run it slowly, or postpone it.

## Decision
1. **Request before transfer.** Clients send `POST /request {device, files, bytes, reason}`. Nothing transfers until the hub returns a grant.
2. **Load sampling.** The hub samples, over a short window (default 30 s): CPU %, memory %, **per-disk active time and queue length** for the source and target disks, on battery or not, fullscreen app or game, and the **top processes** by CPU, memory and disk.
3. **Policy:**
   - Below thresholds → grant automatically (unless `always_ask: true`).
   - Above → interactive notification showing the load and the top processes, with:
     **Back up now · Back up slowly · When idle · In 1 hour · Skip**.
   - No answer within a timeout → treated as **When idle**.
4. **Slow mode:** below-normal CPU priority, low I/O priority, restic `--limit-*` / transfer bandwidth caps.
5. **When idle:** start once load stays below thresholds for N minutes (default 5).
6. **Reply to the client:** `granted | deferred(until) | slow | skipped`, so the phone can show *"PC deferred the backup. It will start when idle."*
7. **Applies to every automatic job:** drive-arrival backups, DB dumps and full verification checks, not just phone requests.

Default thresholds (configurable): CPU > 70%, memory > 85%, disk active > 80%, sustained across the window.

## Alternatives considered
- **Always start immediately:** bad experience while the PC is busy.
- **Always ask:** too many prompts; available as an option.
- **Only throttle, never ask:** the user explicitly wants to decide, and to see what to close.

## Consequences
- Toast **action buttons** on Windows need an AUMID and activation handling (spike AGT-6). Fallback: open a small dialog window from the tray.
- The client protocol has to handle deferred grants (polling or a callback when the hub is ready).
