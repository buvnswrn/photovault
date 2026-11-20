# ADR-0010: Phone connectivity: wired and wireless

- **Status:** Accepted (the phone-client technology is still open; see spike ING-5)
- **Date:** 2026-11-20
- **Related:** ADR-0004, ADR-0011, ADR-0016

## Context
The phone is a Samsung S24 Ultra (Android). The user wants **both** wired (USB) and wireless transfers. Over USB, Android uses MTP, so the phone **gets no drive letter** and ordinary file copying doesn't work. Blueprint v0 proposed Termux + a Go binary + MacroDroid for wireless push; One UI's battery optimization and Android's background-process killer make that fragile.

## Decision
Support both modes, feeding the same inbox → staging pipeline:
- **Wired:**
  - Primary: **adb** (`adb pull`), when USB debugging is enabled. Detect the phone by its USB/adb serial when it's plugged in.
  - Fallback: **MTP** via the Windows Portable Devices (WPD) API, needing no phone setup.
- **Wireless:**
  - The phone client talks to the hub's HTTP API over the **LAN** (chunked upload with SHA-256), using the request → consent protocol (ADR-0011).
  - The hub is found by a configured address or mDNS. Tailscale is optional (ADR-0016).
- The phone-client technology (Termux + MacroDroid vs. a Syncthing fork in send-only mode vs. a small native app) is decided by **spike ING-5**, judged on: surviving One UI background limits, storage access, setup effort, and support for the consent flow.

## Alternatives considered
- **Wireless only (blueprint v0):** the user explicitly wants wired too; USB is also faster for big initial imports.
- **MTP only:** no setup, but slow and awkward to automate; kept as a fallback.
- **Cloud relay (Google Photos / Drive) as transport:** sends private data to a third party, and download quotas apply; rejected as the main path.

## Consequences
- Several transport types to maintain; mitigated by a common `Source` interface (list, hash, fetch).
- adb requires USB debugging, which has security implications; the user enables it knowingly. MTP remains as the no-setup path.
