# ADR-0017: Licensing: AGPL-3.0 with commercial dual licensing

- **Status:** Proposed (licence file applied on 2026-11-20 at the owner's request; formal acceptance via Q-6)
- **Date:** 2026-11-20
- **Related:** [CLA.md](../../CLA.md), Q-7, FND-6

## Context
The owner wants two things: **maximum adoption** and **the ability to earn money** when others make money with photovault. The repo is public and contains no code yet, so there's a clean starting point: nothing has been released under another licence.

Facts that constrain the choice:
- Open-source licences (as defined by the OSI) **can't require payment** or royalties.
- The audience (self-hosters, privacy-minded individuals) prefers real open source. **F-Droid**, the likely store for the Android client, accepts only free/open-source software.
- restic is BSD-2-Clause and is called as a separate program, so it's compatible with any licence; bundled binaries must include its licence text.

## Decision
- photovault is licensed under the **GNU Affero General Public License v3.0** (`AGPL-3.0`, see `LICENSE`).
- **Dual licensing:** companies that don't want to meet the AGPL's obligations (publishing the source of modified versions, including when offered as a network service) can buy a **commercial licence** from the maintainer.
- **Every external contribution needs a CLA** ([CLA.md](../../CLA.md)), enforced with the CLA Assistant bot, so that all code can be dual-licensed.
- Optional extra income, without restricting the licence: GitHub Sponsors, a voluntary "supporter licence", a paid convenience build (e.g. on the Play Store), and possibly a paid offsite-storage option later (LTR-1).

## Alternatives considered
- **MIT / Apache-2.0:** highest adoption, but no leverage; a company could sell a product built on photovault and owe nothing.
- **GPL-3.0:** doesn't cover use as an online service; a hosted backup service could use a modified version without sharing it.
- **PolyForm Noncommercial / Small Business:** directly requires payment for commercial use, but isn't open source: blocked from F-Droid and Linux distributions, and less trusted by the target audience.
- **BSL 1.1 / FSL:** built to stop hosted competitors in company products; little benefit for a personal tool, and also not open source at release time.
- **No licence:** blocks all legal reuse, so no adoption at all.

## Consequences
- Individuals can use photovault freely. Companies that would embed or host it either share their changes or pay.
- External contributions without a signed CLA can't be merged.
- The AGPL applies going forward: versions already released stay AGPL even if the licensing changes later.
- **Before the first release (Q-7):** confirm that the owner holds the rights. Part of the design work was done on an employer-provided laptop, and German law (§69b UrhG) assigns the economic rights to software created by employees *in the course of their duties* to the employer. A private project outside one's duties is normally the employee's own, but this should be confirmed against the employment contract.
- Commercial licence contracts and the CLA text need a lawyer's review before the first sale.
