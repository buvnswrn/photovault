# ADR-0012: Monitoring and reminders

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0003, ADR-0009

## Context
The PC is rarely on and there's no always-on server, so the PC can't remind the user while it's off. The user wants to see status from anywhere and has a personal website. An earlier idea, a Tailscale-reachable dashboard on the PC, is useless while the PC is off.

## Decision
Three layers:
1. **Healthchecks.io "dead man's switch"** (hosted; self-hostable later). One check per tier (e.g. `ssd-backup` every 14 days, `hdd-backup` every 30 days). The agent pings `/start`, success or `/fail` with a short, **sanitized** reason. If no success ping arrives within the period plus grace, Healthchecks emails or pushes a reminder: *"No SSD backup for 14 days. Start the PC and plug in the drives."*
2. **Local status** on the PC: a tray icon in four colours (all backed up / pending / running / failed), a hover summary of pending work, and a local dashboard for history, drive usage, SMART health and verification results.
3. **Personal website (optional):** at the end of a session the agent **pushes** a sanitized `status.json` (last success per tier, sizes, drive health) to the site. The site can show Healthchecks badges publicly and details only behind a login.

**Security rule:** data only flows *out* of the PC. The website and Healthchecks never connect to the PC. No paths, filenames, database names or serials are sent outside the PC.

## Alternatives considered
- **Tailscale + a dashboard on the PC as the remote view:** unavailable while the PC is off. Kept only for optional roaming (ADR-0016).
- **Self-hosted Uptime Kuma / Grafana:** needs an always-on server, which doesn't exist.
- **Phone-only reminders (blueprint v0's offline notice):** kept (PHN-1), but they depend on the phone app staying alive; Healthchecks is the independent fallback.

## Consequences
- Depends on a third-party service for reminders; it's free for this scale and can be self-hosted later.
- The website integration depends on its stack (open question).
