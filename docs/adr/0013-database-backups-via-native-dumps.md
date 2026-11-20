# ADR-0013: Database backups via native dump tools

- **Status:** Accepted (the list of engines is still open)
- **Date:** 2026-11-20
- **Related:** ADR-0004, ADR-0011

## Context
The user wants their local databases backed up as well. Copying a running database's data files can produce an inconsistent, unrecoverable copy.

## Decision
- Back up databases **only through their native consistent-export tools**, written into `staging/db/<date>/<name>`:
  `pg_dump -Fc` (PostgreSQL), `mysqldump --single-transaction` (MySQL/MariaDB), `sqlite3 .backup` (SQLite), `mongodump --archive` (MongoDB), `BACKUP DATABASE … TO DISK` (SQL Server).
- **Docker-aware:** if the database runs in a container, run the dump via `docker exec`.
- Dump files are hashed into the manifest and then follow the normal staging → SSD → HDD path.
- DB dumps are a local-only step, so they can run at agent start without any drive present (subject to the load gate).

## Alternatives considered
- **Copying data directories:** inconsistent while the DB is running; rejected.
- **Stopping the DB, then copying files:** consistent, but disruptive and engine-specific; dumps are simpler.
- **Filesystem snapshots (VSS):** possible later for large DBs; more complex.

## Consequences
- The required client tools (`pg_dump`, etc.) must be available on the PC or in the container.
- Restore testing for dumps is a separate task (DB-2).
- Which engines to support is an open question.
