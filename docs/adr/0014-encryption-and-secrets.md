# ADR-0014: Encryption and secret storage

- **Status:** Proposed
- **Date:** 2026-11-20
- **Related:** ADR-0002, ADR-0006

## Context
The backups hold private photos and database contents. External drives can be lost or stolen, and an offsite copy would sit with a third party.

## Decision
- All restic repositories are **encrypted** (restic always encrypts).
- The repository password is stored in **Windows Credential Manager** (Linux: the Secret Service / keyring). It is never stored in `config.yaml` or the repo.
- The user keeps **one offline copy** of the password (e.g. on paper). Without it the backups can't be recovered; the setup flow says this clearly.
- The hub's HTTP API needs a **per-device token**, issued when a device is paired. API traffic stays on the LAN or Tailscale; it is never exposed to the internet.
- Healthchecks ping URLs and website upload credentials are treated as secrets and stored the same way.

## Alternatives considered
- **Password in config:** leaks via backups of the config or the repo.
- **No encryption on local drives:** a lost drive exposes everything.
- **BitLocker on the drives instead:** fine as an extra layer, but doesn't cover offsite copies; restic encryption works on every tier.

## Consequences
- Losing the password means losing the backups, so a password-recovery reminder belongs in the setup checklist.
- Device pairing needs a small UX flow (`photovault pair`).
