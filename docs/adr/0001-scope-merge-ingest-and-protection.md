# ADR-0001: Scope: one system for ingest and protection

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** [ROADMAP](../ROADMAP.md), [blueprint v0](../ideas/photovault-blueprint-v0.md)

## Context
Two designs existed:
- **Blueprint v0 (photovault):** a hub/spoke tool that moves photos from an Android phone (Samsung S24 Ultra) and laptops onto a PC.
- **Backup Agent:** turns local data (phone folders, local databases) into versioned, verified backups on an SSD, then an HDD, triggered when the drives are plugged in.

Real-world constraints:
1. The hub PC is rarely switched on.
2. The SSD and HDD are connected by hand, not every session.
3. There is no always-on server.

## Decision
We will build **one system, named `photovault`**, with three layers:
1. **Ingest:** get data off devices (phone, laptops) and dump local databases into a staging area.
2. **Protect:** turn staging into versioned, encrypted, verified backups on the SSD, then the HDD, then optionally offsite.
3. **Operate:** pending-work status, consent prompts, reminders and monitoring.

Scope covers photos and other phone folders, local databases, and laptops (later).

## Alternatives considered
- **Two separate tools** (blueprint for ingest, scripts for backup): duplicated config, state and notifications, and two things to keep running on a PC that is rarely on.
- **Photos only (blueprint scope):** leaves databases and versioning out, and a plain mirror copy is not a backup.

## Consequences
- The name `photovault` now covers more than photos; we keep it anyway.
- The design must work under all three constraints, which drives ADR-0003, -0004, -0009 and -0012.
