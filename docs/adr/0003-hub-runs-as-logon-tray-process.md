# ADR-0003: Hub runs as a logon-time tray process, not a Windows Service

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0009, ADR-0011

## Context
Blueprint v0 registers the hub as a Windows Service / systemd unit (`kardianos/service`). But the hub must:
- show a tray icon with status;
- show notifications **with action buttons** (consent prompts, ADR-0011);
- show "safe to unplug" messages.

Windows Services run in **session 0**, which is isolated from the user's desktop, so they can't show a tray icon or interactive notifications. The PC is also a single-user desktop that is only on while the user is at it.

## Decision
On Windows, the hub (`photovault agent`) runs as a **user process started at logon** (a Task Scheduler "At log on" task, hidden window) with a tray icon. On Linux, it runs as a **systemd user unit** in the graphical session.

## Alternatives considered
- **Windows Service + a separate tray helper talking over IPC:** works, but adds a second process, an IPC protocol and elevated install for no benefit on a single-user PC. We can revisit if headless use is ever needed.
- **Windows Service only (blueprint v0):** can't prompt the user. Rejected.

## Consequences
- Backups only run while someone is logged in. That's acceptable: the PC is only on when the user is using it.
- Installation needs no administrator rights.
- The Windows toast action buttons need an App User Model ID (AUMID) and activation handling; spike AGT-6 covers this.
