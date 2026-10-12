# ADR 0007: Isolated Shadow Analysis and Coverage

**Status:** Accepted

## Context

Collector fan-out within a shared process is not a failure boundary. An analysis branch can apply backpressure, return an error, consume memory or terminate the forwarding process. V0 must not modify production forwarding or claim that analysis is fail-open.

## Decision

V0 analysis runs in a separate failure domain from production forwarding.

V0 accepts either:

- offline OTLP replay, or
- an explicitly best-effort replicated, non-authoritative OTLP stream.

V0 does not own or modify the production forwarding path. Every shadow report records input coverage, rejected or malformed records, capture drops, state overflows and analysis failures.

## Consequences

Positive:

- analysis failure cannot be represented as a transparent forwarding guarantee
- reports disclose incomplete input rather than silently estimating from it
- offline replay provides a safe, reproducible starting point

Negative:

- a replicated stream can be incomplete or unavailable
- production forwarding remains an operator-owned concern in V0
- an integration topology must be tested before any availability claim is made
