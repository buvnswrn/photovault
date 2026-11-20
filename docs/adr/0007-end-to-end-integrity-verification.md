# ADR-0007: End-to-end integrity verification

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0004, ADR-0006

## Context
A copy that hasn't been verified isn't a reliable backup. Data can be corrupted in transfer, by failing media (silent "bit rot", especially on long-stored HDDs), or by bugs. Staging is deleted after promotion, so verification must happen *before* that.

## Decision
Use SHA-256 at every step:
1. **Device → hub:** the client sends SHA-256 per file; the hub re-hashes on receipt and rejects mismatches. Over adb/MTP, the hub hashes after the copy and compares size and timestamp with the source listing.
2. **Staging:** each batch gets a manifest (`relative path, size, sha256`).
3. **Staging → SSD:** after `restic backup`, restore that snapshot's new files to a temporary folder and compare them with the manifest. Only a passing check marks staging eligible for deletion.
4. **SSD → HDD:** after `restic copy`, run `restic check --read-data-subset=5%` each session, and a full `restic check --read-data` at least weekly (when the drive is present and the PC is idle).
5. **Restore test:** monthly, restore random files from the HDD and compare them with the stored manifests.

Results are written to the state DB and shown in the status/dashboard.

## Alternatives considered
- **Trust the copy tool's exit code:** doesn't detect corruption.
- **Compare size/timestamp only:** misses silent corruption.
- **Full read-data check every session:** too slow on an HDD for short sessions; the subset plus weekly full check is the compromise.

## Consequences
- Verification adds time to sessions; the load gate (ADR-0011) can defer the heavy checks.
- Manifests are kept (small), so restore tests can be checked against the original hashes.
