# Recovery runbook

> **Use this when photovault itself is gone or broken:** dead PC, lost config, a bug.
> It needs only **restic** and your **repository password**. photovault is *not* required (safety invariant [S13](SAFETY.md)).
>
> **Status:** Draft. The exact repository paths are filled in when PRO-1 is implemented. Keep a printed copy, and a copy on each backup drive (`RECOVERY.md` next to the repository folder).

## What you need
1. **One backup drive:** the SSD (newest data) or the HDD (older; use it if the SSD is dead), or the offsite copy if it exists.
2. **The repository password:** from Windows Credential Manager on the old PC (entry `photovault/restic`), or from your **offline paper copy**.
   ⚠️ Without the password the data **cannot be recovered**. No one can reset it.
3. **restic** for your OS: <https://github.com/restic/restic/releases>. Check the download against `SHA256SUMS` from the same release.

## Steps (Windows, any PC)

```powershell
# 1. Point restic at the repository on the backup drive (drive letter may differ)
$env:RESTIC_REPOSITORY = 'E:\photovault\repo'      # TODO(PRO-1): confirm the final folder name
$env:RESTIC_PASSWORD   = Read-Host 'Repository password'   # or set RESTIC_PASSWORD_FILE

# 2. Check that the repository is readable and intact
restic snapshots
restic check                  # add --read-data for a full check (slow on an HDD)

# 3. Look inside a snapshot
restic ls latest                                   # list files
restic find "IMG_2026*"                            # search by name

# 4. Restore: everything, or just a part
restic restore latest --target D:\restore
restic restore latest --target D:\restore --include "/phone/2026/11"
restic restore <snapshot-id> --target D:\restore   # an older version
```
On Linux or macOS, the same commands work with `export RESTIC_REPOSITORY=/media/…`. There you can also browse a snapshot like a folder with `restic mount /mnt/restic`; this isn't available on Windows.

## Restoring databases
Dumps are restored to `D:\restore\db\<date>\`. Load them with the database's own tool:

| Engine | Restore command |
|---|---|
| PostgreSQL (`.dump`, `-Fc`) | `pg_restore -d <db> --clean --if-exists file.dump` |
| MySQL / MariaDB (`.sql`) | `mysql <db> < file.sql` |
| SQLite (`.db`) | Copy the file into place while the app is stopped |
| MongoDB (`.archive`) | `mongorestore --archive=file.archive` |
| SQL Server (`.bak`) | `RESTORE DATABASE <db> FROM DISK = 'file.bak'` |
| In Docker | `docker exec -i <container> <command above>` |

## Which copy to use
| Situation | Use |
|---|---|
| PC dead, drives fine | **SSD** (newest) |
| SSD dead or corrupted | **HDD** (may be a few weeks older) |
| Both lost (fire, theft) | **Offsite** copy, if set up (LTR-1); otherwise the data is gone. This is why 3-2-1 matters |
| A file deleted or damaged long ago | An **older snapshot** (`restic snapshots`, then restore by ID) |
| `restic check` reports errors | Try the other drive first. Don't run `restic repair` on the only copy until you've made a copy of the repository folder |

## Yearly recovery drill (TST-2)
Once a year, on a PC **other than the hub**, using only this document:
- [ ] Install restic from the official release and verify the checksum
- [ ] Unlock the HDD repository with the **paper** password
- [ ] `restic check --read-data-subset=10%` passes
- [ ] Restore 20 random photos and 1 database dump; open them and load the dump
- [ ] Fix anything in this document that was wrong or unclear
- [ ] Record the date and result in issue TST-2
