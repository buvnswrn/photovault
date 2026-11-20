# ADR-0008: Append-only: device deletions never propagate

- **Status:** Proposed
- **Date:** 2026-11-20
- **Related:** ADR-0004, ADR-0006

## Context
Blueprint v0 doesn't say what happens when a photo is deleted on the phone. Two-way sync tools (e.g. Syncthing in send/receive mode) would delete it on the hub too. That makes accidental deletion, or a compromised or wiped phone, a data-loss event.

## Decision
- Ingest is **append-only**. Deleting or editing a file on a device never removes or overwrites data on the hub.
- An edited file with the same name but a new hash is stored as a new version (restic handles versions; staging keeps both, with a suffix).
- photovault **never deletes from the phone automatically**. At most it reports which files are safely in **both** the SSD and HDD repositories, so the user can free up space by hand.
- Sync tools, if used, must run as **send-only** on the device and **receive-only** on the hub.

## Alternatives considered
- **Mirror deletions:** turns backups into a sync; rejected.
- **Auto-delete from the phone after backup:** convenient but risky; a later opt-in feature at most, and only after both tiers are verified.

## Consequences
- The hub's storage only grows; retention (ADR-0006) removes old *snapshots*, never the last copy of a file.
- Removing data on purpose (e.g. a private photo) needs an explicit `photovault forget` command, out of scope for now.
