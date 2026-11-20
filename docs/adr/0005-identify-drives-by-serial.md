# ADR-0005: Identify backup drives by serial number

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0004, ADR-0009

## Context
The user has an SSD, an HDD and a pendrive that get plugged in by hand. Drive letters (`D:`, `E:`) and disk numbers change with port, plug order and reboots. Writing a backup to the wrong device (e.g. the pendrive) could fill it or overwrite data.

## Decision
Backup targets are configured by **disk serial number** (and optionally the volume's unique ID), never by drive letter or mount path:
- Windows: `MSFT_PhysicalDisk` / `MSFT_Disk` (`Get-PhysicalDisk`, `Get-Disk`) → partition → volume → current mount path.
- Linux: `lsblk -o SERIAL,…` / `/dev/disk/by-id/`.

Before every write, the agent works out the current mount path from the serial. If the serial isn't found, or more than one candidate matches, it **refuses** to write. Unknown drives are ignored.

## Alternatives considered
- **Mount path / drive letter (blueprint v0):** unstable; risk of writing to the wrong drive.
- **Volume label:** user-editable and not unique.
- **A marker file on the drive:** a useful *extra* check (e.g. the restic repository config), but it can be copied; the serial is the primary key.

## Consequences
- Setup includes a step where the user plugs in each drive and the agent records its serial.
- Replacing a drive needs a config update; the agent must make this easy (`photovault drives add`).
