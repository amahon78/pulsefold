# ADR 0002: AI Is Not Required for the Telemetry Plane

**Status:** Accepted

## Context

PulseFold must remain useful in security sensitive environments and avoid model latency/cost in the high throughput path.

## Decision

Core telemetry receive, normalise, analyse, compact and export functions must operate with AI disabled.

## Consequences

AI is an optional intelligence layer rather than a dependency for data movement.
