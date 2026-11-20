# photovault: ideas, comparison & backlog

_Last updated: 2026-11-20. Status: planning, nothing built yet._

This document merges two idea sets into one plan and turns it into a backlog that can be copied into GitHub issues.

- **Idea A: `photovault` blueprint.** The earlier Go hub/spoke design. Kept verbatim in [`ideas/photovault-blueprint-v0.md`](ideas/photovault-blueprint-v0.md).
- **Idea B: "Backup Agent".** Designed in a later chat: tiered staging → SSD → HDD backups, an event-driven tray agent and Healthchecks reminders.
- **New requirements** added on 2026-11-20 (see §3).

---

## 1. Verdict

The two ideas **complement each other; they don't compete**:

| Layer | Covered by | Job |
|---|---|---|
| **Ingest** | Idea A | Get photos *off* devices (phone, laptops) and onto the hub PC |
| **Protect** | Idea B | Turn what landed into versioned, verified, encrypted backups on SSD → HDD (→ cloud) |
| **Operate** | Idea B, plus Idea A's snooze/notify | Tell the user what's pending, remind them while the PC is off, ask before heavy work |

**Merged design:** one Go program (`photovault`) does ingest, the agent, prompts and status. It calls **restic** to manage the backup repositories on the SSD and HDD. The project keeps the name `photovault`, even though it now also covers databases.

### Real-world constraints (these shape every decision)
1. The hub PC is **rarely switched on**.
2. The SSD and HDD are **connected by hand**, and not every session.
3. There is **no always-on server**.
4. Phone: **Samsung S24 Ultra (Android)**.

---

## 2. Side-by-side comparison

| Topic | Idea A: photovault blueprint | Idea B: Backup Agent | Merged decision |
|---|---|---|---|
| Scope | Photos from phone and laptops | Phone folders, local DBs, full backup lifecycle | **Both** |
| Language | Go single binary, cross-platform | PowerShell + restic, Windows-only | **Go** for the program, **restic** as an external tool. The new requirements (HTTP API, phone client, load monitoring, interactive prompts) outgrow PowerShell |
| Hub runtime | Windows Service / systemd, always running | Tray app that starts at logon | **Tray app at user logon.** A Windows Service runs in *session 0* and **cannot show tray icons or interactive notifications**, which the consent prompts (§3.2) need |
| Where incoming files go | Straight to an external drive; refuses if not mounted | Staging folder on the internal disk, promoted later | **Staging first.** Under Idea A's mount guard, uploads would be refused almost every time, because the drives are usually unplugged. The guard becomes "staging has room (quota)" plus "the target drive is identified by its serial before promoting" |
| Drive identification | Mount path exists | Drive **serial number**, so the wrong drive is never written to | **Serial number.** A path such as `D:\` can belong to the pendrive on another day |
| Copy model | Mirror copy into `YYYY/MM/` | Versioned, deduplicated, encrypted snapshots (restic) | **Both:** organize into `YYYY/MM/` in staging, then take restic snapshots. Plain mirrors don't protect against accidental deletion or ransomware |
| Integrity | SHA-256 on upload | SHA-256 manifests, restore-and-compare, `restic check --read-data` | **Both,** end to end |
| Deduplication | SQLite of file hashes on the client | restic deduplicates storage | **Both:** SQLite stops repeat *transfers*; restic stops repeat *storage* |
| Phone transport | Termux + Go binary + MacroDroid, over Wi-Fi | Syncthing (Wi-Fi) | **Wired and wireless** (§3.1). The phone-side client still needs a spike (#ING-5) |
| Hub offline | Phone notifies "turn on PC", with **snooze** and **active hours** | Healthchecks.io reminder after N days | **Both:** the phone prompt for immediate needs, Healthchecks as a fallback that works even if the phone app dies |
| Remote access | Tailscale (MagicDNS, roaming) | Tailscale dropped (PC rarely on); local dashboard instead | **LAN first, Tailscale optional** (useful for laptops or when away, while the PC is on) |
| Monitoring | `photovault status` CLI | Tray status, local dashboard, Healthchecks, optional push to the personal website | **All of them** |
| DB backups | none | Native dump tools → staging → restic | **Keep** |
| Retention / restore tests / SMART / safe eject | none | Yes | **Keep** |
| Cloud | "Overflow" with a hard GB quota | "Offsite copy" (3-2-1 rule) | **Two different features,** both kept: *overflow* holds files while the hub is offline; *offsite* is the disaster copy |
| Distribution | GoReleaser, GitHub Releases, install scripts | none | **Keep** |
| Extra clients | Windows/Linux laptops | none | **Keep,** later milestone |
| Wake-on-LAN, Tailscale webhooks | Future scope | none | **Low priority.** The drives still have to be plugged in by hand, so waking the PC remotely gains little |

### New ideas from Idea A that Idea B didn't have
- Phone-side **"hub offline"** notifications with **snooze** and **active hours**
- `YYYY/MM` organization (use the **EXIF capture date**, falling back to the file's modified time)
- SQLite transfer tracking, so files are never re-sent
- Laptops as clients
- Cloud overflow with a quota cap
- GoReleaser cross-platform releases and one-line installers
- A CLI (`init`, `serve`, `sync`, `watch`, `snooze`, `status`)

### What Idea B adds that Idea A was missing
- Tiered staging → SSD → HDD design for drives that are connected by hand
- Drive identification by serial number
- Versioning, retention policy, encryption
- Restore tests, full data verification, SMART health checks
- Database dumps
- Safe flush and eject, plus a "safe to unplug" notification
- Healthchecks "dead man's switch" reminders while the PC is off
- Pushing status to the personal website (outbound only)

### Problems in Idea A to fix
1. A **Windows Service can't show prompts or a tray icon.** Run the hub as a logon-time user process instead.
2. The **mount guard blocks nearly every upload** under the "drives plugged in by hand" constraint. Use staging.
3. There's **no deletion rule.** If a photo is deleted on the phone, it must **never** be deleted on the hub (append-only).
4. **Termux + MacroDroid is fragile** on One UI: battery optimization, Android's background-process killer, storage permissions. Validate it in a spike before committing.
5. The `://github.com` and `://tailscale.com` strings in the blueprint are broken links or module paths. Replace them with the real repo path.

---

## 3. New requirements (2026-11-20)

### 3.1 Phone connection: wired AND wireless
| Mode | How | Notes |
|---|---|---|
| **Wired (USB)** | `adb pull` when USB debugging is on; **MTP** (Windows Portable Devices API) as the no-setup fallback | MTP phones get **no drive letter**, so ordinary file copying won't work. Detect the phone by its USB serial when it's plugged in |
| **Wireless (LAN)** | The phone client pushes to the hub's HTTP API (Idea A), *or* Syncthing-style sync into the inbox | Discover the hub via mDNS on the LAN or a fixed IP; Tailscale is optional |

Both modes feed the **same** pipeline: inbox → hash and deduplicate against SQLite → staging `YYYY/MM` → restic.

### 3.2 Phone-initiated backups with a load-aware consent prompt on the PC
1. The phone sends a **backup request**: `{device, file count, total bytes, reason}`. It does **not** start transferring immediately.
2. The hub measures current load: CPU %, memory %, disk activity on the source and target disks (active time / queue length), on battery or not, and whether a fullscreen app or game is running.
3. **Below the thresholds** (configurable): start automatically, or still ask if `always_ask: true`.
   **Above the thresholds:** show a notification on the PC:
   > 📱 *S24 Ultra wants to back up 1,240 files (8.1 GB). Your PC is busy: CPU 87%, Disk 95%.*
   > *Top: chrome.exe (CPU 41%), docker (Mem 6.2 GB), OneDrive (Disk 60%)*
   > **[Back up now]  [Back up slowly]  [When idle]  [In 1 hour]  [Skip]**
4. **"Back up slowly"** runs at low CPU and I/O priority with a bandwidth limit, so it doesn't slow down whatever else you're doing.
5. **"When idle"** waits until load stays below the thresholds for N minutes, then starts.
6. The hub replies to the phone, which shows *"PC deferred backup. It will start when idle,"* or *"Started."*
7. The **same load check** applies to every automatic job, such as drive-plug-in backups and DB dumps, not only phone requests.

Technical risk: Windows toast **action buttons** need an App User Model ID (AUMID) and protocol or COM activation. Prototype this early (#AGT-6).

---

## 4. Merged architecture

```text
 S24 Ultra ──USB (adb/MTP)──┐
 S24 Ultra ──Wi-Fi (HTTP)───┤    ┌──────────────── Hub PC: photovault agent (tray) ───────────────┐
 Laptops ───Wi-Fi/Tailscale─┤    │  request → load check → [consent prompt] → transfer            │
                            └───►│  inbox → SHA-256 + SQLite dedupe → staging/YYYY/MM (+manifest) │
 Local DBs ──native dump────────►│  staging/db/<date>                                              │
                                 │        │ SSD plugged in (serial matched)                         │
                                 │        ▼                                                         │
                                 │  restic backup → SSD repo → restore & compare → staging cleanup │
                                 │        │ HDD plugged in (serial matched)                         │
                                 │        ▼                                                         │
                                 │  restic copy → HDD repo → check → flush + eject → "safe to unplug"│
                                 └───────┬─────────────────────────┬──────────────────────────────┘
                                         │ pings                   │ status.json (outbound only)
                                         ▼                         ▼
                                  Healthchecks.io            Personal website
                            ("no backup in 14 days" alerts)   (public badges / login-only page)
```

---

## 5. Decisions

The full reasoning, including rejected alternatives, is in the **[Architecture Decision Records](adr/README.md)**.

| Decision | ADR | Status |
|---|---|---|
| One system for ingest and protection | [0001](adr/0001-scope-merge-ingest-and-protection.md) | Proposed |
| One Go program; restic for the backup repositories | [0002](adr/0002-go-program-with-restic.md) | Proposed: confirm (Q-5) |
| Hub runs as a logon-time tray process, not a Windows Service | [0003](adr/0003-hub-runs-as-logon-tray-process.md) | Proposed: confirm (Q-5) |
| Staging on the internal disk with a size quota | [0004](adr/0004-staging-first-ingest.md) | Proposed |
| Drives identified by serial number | [0005](adr/0005-identify-drives-by-serial.md) | Proposed |
| Tiered storage and retention | [0006](adr/0006-tiered-storage-and-retention.md) | Proposed |
| End-to-end integrity verification | [0007](adr/0007-end-to-end-integrity-verification.md) | Proposed |
| Append-only: never delete on the hub because a file was deleted on the phone | [0008](adr/0008-append-only-no-deletion-propagation.md) | Proposed |
| Event-driven triggers | [0009](adr/0009-event-driven-triggers.md) | Proposed |
| Phone: wired and wireless | [0010](adr/0010-phone-connectivity-wired-and-wireless.md) | Accepted |
| Every automatic job goes through the load check / consent gate | [0011](adr/0011-request-consent-and-load-aware-gating.md) | Proposed |
| Monitoring and reminders; the website only receives status | [0012](adr/0012-monitoring-and-reminders.md) | Proposed |
| DB backups via native dumps | [0013](adr/0013-database-backups-via-native-dumps.md) | Proposed |
| Encryption and secrets | [0014](adr/0014-encryption-and-secrets.md) | Proposed |
| Cloud overflow vs. offsite | [0015](adr/0015-cloud-overflow-vs-offsite.md) | Deferred |
| Tailscale roaming, WoL, node-online trigger | [0016](adr/0016-deferred-remote-wake-and-roaming.md) | Deferred |
| Licensing: AGPL-3.0 + commercial dual licence | [0017](adr/0017-licensing-agpl-with-commercial-dual-licence.md) | Proposed |

## 6. Open questions
1. **Databases:** which engines (Postgres, MySQL, SQLite, Mongo, SQL Server)? Native installs or Docker?
2. **Website:** what is it built with and hosted on? Public badges only, or a page behind a login?
3. **Offsite / cloud:** now, later or never? Which provider (B2, S3, Google Drive)?
4. **Reminder periods:** 14 days for the SSD and 30 days for the HDD?
5. **Phone client:** Termux + MacroDroid, a Syncthing-fork, or a small native Android app? Settled by spike #ING-5.
6. **Default load thresholds:** proposed CPU > 70%, memory > 85%, disk active > 80%, each sustained for 30 s.

---

## 7. Backlog (GitHub-issue ready)

Suggested labels: `epic:*`, `type:feature|spike|chore`, `prio:P0|P1|P2|P3`, `platform:windows|android|linux`.
Format: **ID: title** (priority) · depends on · acceptance criteria.

### Epic M0: Foundations
- **FND-1: Repo scaffold, Go module, Cobra CLI skeleton** (P0)
  AC: `photovault --help` lists `init, agent, status, sync, snooze`; CI builds on Windows and Linux.
- **FND-2: Config file (`config.yaml`) with validation** (P0) · FND-1
  AC: drives (by serial), sources, DB list, thresholds, retention and Healthchecks URLs; clear errors for invalid values.
- **FND-3: State DB (SQLite): transfers, hashes, jobs, last-success times** (P0) · FND-1
- **FND-4: Drive identification by serial (Windows `MSFT_PhysicalDisk`/`MSFT_Disk`; Linux `lsblk`)** (P0)
  AC: finds the SSD/HDD by serial and returns the mount path; refuses an unknown drive or a pendrive.
- **FND-5: Structured logging and per-run JSON status files** (P1)

### Epic M1: Ingest
- **ING-1: Inbox → staging pipeline: SHA-256, SQLite dedupe, `YYYY/MM` by EXIF date, manifest CSV** (P0) · FND-3
- **ING-2: Wired phone via adb (`adb pull`, detect device on USB connect)** (P0) · ING-1
- **ING-3: Wired phone via MTP / WPD fallback (no USB debugging needed)** (P2) · ING-1
- **ING-4: Wireless: hub HTTP API (`/health`, `/request`, `/upload` chunked + SHA-256)** (P1) · ING-1
- **ING-5: SPIKE: phone client options on S24 Ultra (Termux+MacroDroid vs. Syncthing-fork vs. native app)** (P1)
  AC: a written comparison covering battery-kill behaviour, storage access, setup effort and support for the consent flow (§3.2).
- **ING-6: Phone client: scan DCIM, track sent files, send backup request, handle the reply** (P1) · ING-4, ING-5
- **ING-7: Append-only rule: phone-side deletions never remove hub data** (P1) · ING-1

### Epic M2: Protect
- **PRO-1: restic SSD repository: init, password in Windows Credential Manager, backup staging** (P0) · FND-4
- **PRO-2: Verify after backup: restore the new snapshot and compare with the manifest; only then clean up staging** (P0) · PRO-1, ING-1
- **PRO-3: Replicate SSD → HDD (`restic copy`) and nightly subset check** (P0) · PRO-1
- **PRO-4: Retention policy (`restic forget --prune`) for each repository** (P1)
- **PRO-5: Weekly full `restic check --read-data` and monthly random restore test** (P1)
- **PRO-6: Flush and safe-eject drives, then "safe to unplug" notification** (P0) · PRO-3
- **PRO-7: Resume after shutdown or unplugging partway through a backup** (P1)

### Epic M3: Agent & UX
- **AGT-1: Tray app started at logon (Task Scheduler), status icon colours** (P0)
- **AGT-2: Calculate pending work on startup, with hover summary and startup toast** (P0) · FND-3
- **AGT-3: React to drive arrival (volume-change events) → run the steps those drives allow** (P0) · FND-4
- **AGT-4: Load monitor: CPU, memory, per-disk activity, battery, fullscreen app, top processes** (P0)
- **AGT-5: Consent gate for every automatic job (now / slowly / when idle / later / skip)** (P0) · AGT-4
- **AGT-6: SPIKE: Windows toast action buttons from Go (AUMID, activation)** (P0)
- **AGT-7: "Back up slowly" mode: low CPU/IO priority, bandwidth limit** (P1)
- **AGT-8: Local dashboard (opened from the tray): history, drive usage, SMART, verification results** (P2)

### Epic M4: Databases
- **DB-1: Dump adapters (pg_dump / mysqldump / sqlite `.backup` / mongodump), Docker-aware** (P1) · open question 1
- **DB-2: Restore test for DB dumps (load into a temporary instance, or check the dump file is valid)** (P2)

### Epic M5: Monitoring & reminders
- **MON-1: Healthchecks.io pings (start/success/fail) per job; reminder periods from config** (P0)
- **MON-2: SMART health via smartctl, warn on worsening values** (P1)
- **MON-3: Push sanitized `status.json` to the personal website (outbound only)** (P2) · open question 2
- **MON-4: Website status page / Healthchecks badges** (P3) · MON-3

### Epic M6: Phone-side smarts (from Idea A)
- **PHN-1: Hub-offline detection, with "Turn on PC to sync N photos" notification** (P1) · ING-6
- **PHN-2: `snooze <duration>` and active hours** (P2) · PHN-1
- **PHN-3: Show the hub's consent reply on the phone (started / deferred / skipped)** (P1) · AGT-5, ING-6

### Epic M7: Distribution
- **DST-1: GoReleaser + GitHub Actions on tag (windows/amd64, linux/amd64, android/arm64)** (P2)
- **DST-2: `install.ps1` / `install.sh` one-line installers** (P3)

### Later / ideas
- **LTR-1:** Offsite copy to cloud (B2/S3) with `restic copy`: the disaster copy (P2, open question 3)
- **LTR-2:** Cloud *overflow* with a hard GB quota while the hub is offline (P3)
- **LTR-3:** Laptops as clients (P3)
- **LTR-4:** Tailscale for roaming clients (P3)
- **LTR-5:** Wake-on-LAN (the phone can send the packet on the LAN; no always-on node needed) (P3)
- **LTR-6:** Tailscale node-online webhook as a backup trigger (P3)

### Suggested build order
`FND-1..4 → ING-1 → PRO-1, PRO-2 → AGT-1..3 → PRO-3, PRO-6 → MON-1 → AGT-6, AGT-4, AGT-5 → ING-2 → ING-5 → ING-4, ING-6 → the rest`

The first usable version is "plug in phone + SSD + HDD → everything backed up and verified → safe to unplug". That's FND, ING-1, ING-2, PRO-1..3, PRO-6, AGT-1..3 and MON-1.
