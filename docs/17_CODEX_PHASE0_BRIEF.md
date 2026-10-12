# Codex Phase 0 Architecture Review Brief

## Timing

Do not start this work until Grok research outputs exist.

## Mission

Act as a sceptical senior infrastructure engineer.

Review PulseFold's proposed architecture against:
- OpenTelemetry implementation realities
- data plane performance
- failure safety
- Go ecosystem constraints
- deployment simplicity
- testability
- maintainability

## Questions to answer

1. Should PulseFold be a Collector distribution, processor set or standalone gateway?
2. Which architectural assumptions would cause excessive maintenance?
3. Can fail open behaviour actually be guaranteed?
4. Where are memory explosions likely?
5. How should bounded grouping windows work under high cardinality?
6. What state belongs in the telemetry plane?
7. What state belongs in the intelligence plane?
8. How should evidence references survive restart?
9. How can we prevent agent functionality from contaminating the data plane?
10. Which V0 requirements should be cut?

## Required output

- architectural risks ranked by severity
- proposed changes
- ADR proposals
- suggested package/module boundaries
- proposed initial benchmark thresholds
- no broad implementation unless explicitly authorised

## Rule

Do not optimise the architecture for the fastest possible demo.
Optimise for a small, credible foundation that can survive continued development.
