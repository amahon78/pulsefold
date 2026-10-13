# Roadmap

## Milestone 0: Architecture ready

**Status:** Complete, 2026-10-12.

Delivered:
- accepted Phase 0 research and independent architecture review
- ADRs 0001 through 0010
- frozen V0 scope
- benchmark definitions and threat-model reconciliation

## Milestone 1: Shadow

**Laboratory implementation status:** Complete, 2026-10-13. This does not approve production deployment, active compaction or C2-based destructive transformation.

User outcome:

> PulseFold shows me how much telemetry could be reduced without changing what I currently send downstream.

Exit:
- logs-only OTLP replay or explicitly best-effort replicated input
- redaction before derivation
- bounded report-only C1 and C2 analysis
- coverage-aware, versioned reduction report
- payload-free provenance manifest and integrity verification for every report
- explicit capture and analysis coverage plus optional scoped billing opportunity accounting
- fault and performance benchmark baseline
- no production-forwarding modification, active compaction, incidents, routing or AI

## Milestone 2: Compact

User outcome:

> PulseFold safely reduces selected repetitive telemetry with auditable policies and recoverable evidence.

Exit:
- deterministic compaction
- policy engine
- representative samples
- evidence references
- active mode safety tests

## Milestone 3: Detect

User outcome:

> PulseFold proactively tells me when something operationally significant is emerging.

Exit:
- signal detectors
- incident state
- webhook
- Slack
- suppression

## Milestone 4: Investigate

User outcome:

> PulseFold gives my chosen agent, or its built in investigator, the evidence needed to investigate the incident.

Exit:
- read only tools
- external agent contract first
- model abstraction after external-agent evaluation
- evidence constrained outputs
- evaluations

## Milestone 5: Operate

User outcome:

> PulseFold is boring enough to leave running.

Exit:
- upgrades
- failure recovery
- resource controls
- security review
- polished install
- public benchmark report
