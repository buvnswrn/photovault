# ADR-0016: Deferred: Tailscale roaming, Wake-on-LAN, node-online triggers

- **Status:** Proposed (deferred)
- **Date:** 2026-11-20
- **Related:** ADR-0010, ADR-0012, [blueprint v0 §2, §4C](../ideas/photovault-blueprint-v0.md)

## Context
Blueprint v0 included:
- **Tailscale** (WireGuard mesh, MagicDNS) so clients reach the hub from anywhere without opening router ports;
- **Wake-on-LAN** through an always-on home node;
- **Tailscale node-online webhooks** to start a backup the moment the hub powers on.

Under the current constraints (the PC is rarely on, the drives are plugged in by hand, no always-on server), these give little value: even a woken PC can't back up to drives that aren't connected, and the agent already acts at logon (ADR-0009).

## Decision
Defer all three; record them so they aren't lost:
- **Tailscale (LTR-4):** optional transport for roaming clients (e.g. a laptop away from home) while the hub is on. The design must not assume LAN-only addresses (use a configurable hub URL).
- **Wake-on-LAN (LTR-5):** if wanted, the **phone can send the magic packet** on the home LAN, so no always-on node is needed. Only useful if the drives are left connected.
- **Node-online trigger (LTR-6):** not needed; the logon-time agent already covers "PC just came on".

## Revisit when
- The drives become permanently attached (e.g. an internal drive or a dock that is always connected), or
- an always-on device (Raspberry Pi / NAS) becomes available, or
- laptops become active clients that are often away from home.

## Consequences
- Clients use a configurable hub URL from the start, so Tailscale can be added later without protocol changes.
