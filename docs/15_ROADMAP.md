# Roadmap

## Milestone 0: Architecture ready

Exit:
- research complete
- ADRs accepted
- V0 scope frozen
- benchmarks defined
- threat model challenged

## Milestone 1: Shadow

User outcome:

> PulseFold shows me how much telemetry could be reduced without changing what I currently send downstream.

Exit:
- OTLP pass-through
- reduction analysis
- report
- performance baseline
- no semantic telemetry changes

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
- model abstraction
- external agent contract
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
