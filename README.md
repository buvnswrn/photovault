# photovault

Personal backup system: phone (wired or wireless), laptops and local databases go to a staging area on the desktop, then to an encrypted, verified, versioned backup on an SSD, then to an HDD, and optionally to the cloud.

It is built for a PC that is **rarely on**, with backup drives that are **plugged in by hand**, and **no always-on server**. You switch on the PC and plug in the drives; photovault does the rest and tells you when it's safe to unplug.

> **Status:** design phase. Nothing is implemented yet.

## Docs
- [Roadmap and backlog](docs/ROADMAP.md): comparison of the ideas, merged architecture, backlog
- [Architecture Decision Records](docs/adr/README.md): what was decided, why, and which alternatives were rejected
- [Safety invariants](docs/SAFETY.md): what photovault must never do, and the fault-injection tests that prove it
- [Recovery runbook](docs/RECOVERY.md): restore your data with plain restic, without photovault
- [Original blueprint (v0)](docs/ideas/photovault-blueprint-v0.md)

## Licensing

Copyright (C) 2026 Bhuvaneshwaran I

photovault is **free and open-source software** under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).

- **Personal and non-commercial use:** free. Use it, modify it, share it.
- **Using it in a product or service:** allowed under the AGPL, but you must publish the complete source of your modified version, including when you only offer it over a network.
- **Commercial licence:** if you'd rather not publish your source, a separate commercial licence is available. Open an issue titled "Commercial licence enquiry" or contact the maintainer via the GitHub profile.
- **Contributing:** contributions require signing the [Contributor License Agreement](CLA.md). The CLA Assistant bot will ask on your first pull request.

Background: [ADR-0017](docs/adr/0017-licensing-agpl-with-commercial-dual-licence.md).

## Issues
The backlog lives in GitHub issues, generated from [`scripts/issues.json`](scripts/issues.json) by [`scripts/create-issues.ps1`](scripts/create-issues.ps1). The script is safe to re-run; existing issues are skipped. New issues can also be opened with the templates (Feature, Bug, Spike, Decision).
