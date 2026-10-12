# ADR 0006: No LLM Destructive Retention Decisions

**Status:** Accepted

## Decision

LLMs do not decide what telemetry is permanently discarded in the initial architecture.

## Rationale

Retention must be reproducible, testable and auditable. Models may advise or classify in experimental paths but cannot silently destroy evidence.
