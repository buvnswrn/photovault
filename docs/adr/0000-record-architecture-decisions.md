# ADR-0000: Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-11-20

## Context
photovault started from two separate idea sets (the original blueprint and a later "Backup Agent" design), discussed across chats. Without a written record, ideas and the reasons ideas were rejected get lost.

## Decision
We will record significant decisions as ADRs in `docs/adr/`, using the template in this folder. Rejected and deferred ideas also get recorded, under "Alternatives considered" or as their own ADRs with a deferred status.

## Alternatives considered
- **Roadmap only:** the roadmap holds *what* to build, not *why*; the reasons get lost when it is edited.
- **GitHub issues only:** issues get closed and buried; they aren't a design history.

## Consequences
- Every non-trivial design change needs a new or updated ADR.
- `docs/ROADMAP.md` links to the ADRs instead of repeating their reasoning.
