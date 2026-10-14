# ADR 0013: Advisory Inference and Telemetry Taint

**Status:** Accepted

## Context

Telemetry can assert `log.record.template` and influence analysis counts and resource use. C2 regex grouping can merge semantically different records.

## Decision

C2 is permanently advisory. It may identify patterns and suggest investigation or policy work, but never directly authorises destructive transformation or emits an executable destructive policy. A structural policy, regardless of whether C2 suggested it, independently satisfies every active-policy gate.

Evidence and advisory inputs carry one of: `telemetry_asserted`, `pulsefold_inferred`, or `operator_defined`. `log.record.template` is `telemetry_asserted`. Telemetry-controlled metadata cannot determine destructive eligibility, authorise transformation or silently gain operator authority.

Telemetry-asserted advisory state must be separately budgeted or isolated from PulseFold inference so asserted templates cannot starve later inferred analysis. Its influence and exclusions are observable.

## Consequences

Source-template counts remain useful but are persuasion-prone, untrusted advisory data. Transport authentication alone does not prove an individual asserted attribute is trustworthy.
