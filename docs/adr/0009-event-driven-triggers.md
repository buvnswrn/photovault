# ADR-0009: Event-driven triggers instead of fixed schedules

- **Status:** Accepted
- **Date:** 2026-11-20
- **Related:** ADR-0003, ADR-0005, ADR-0011, ADR-0012

## Context
The PC is rarely on, and the drives are plugged in by hand. A fixed schedule ("every night at 01:00") would almost never fire, and when it did, the drives probably wouldn't be there. Blueprint v0 also proposed cron-based sweeps (`robfig/cron`).

## Decision
The agent reacts to **what's available right now**:
| Event | Action |
|---|---|
| Agent starts (user logs in) | Work out pending work; show a summary toast; start local-only steps (DB dumps, process the inbox) |
| Known SSD arrives (serial matched) | Staging → SSD, verify, clean up staging |
| Known HDD arrives | SSD → HDD (needs both drives), check |
| Phone connected by USB / backup request over Wi-Fi | Ingest (via consent gate, ADR-0011) |
| All steps done | Flush, eject, "safe to unplug" |

- Each step is a **state machine with stored progress**, so it can be interrupted (shutdown, unplug) and resumed.
- Periodic tasks (weekly full check, monthly restore test) are tracked as "**due**" items that run at the next opportunity, not at fixed times.
- Every automatic action goes through the load/consent gate (ADR-0011).

## Alternatives considered
- **Task Scheduler / cron schedules:** don't fit a PC that is rarely on.
- **Fully manual "Run backup" button:** too easy to forget; kept as a tray menu item, not the main trigger.

## Consequences
- Needs Windows device-arrival events (volume change / WMI) and USB device detection for the phone.
- The PC can't remind the user while it's off, so reminders come from outside (ADR-0012).
