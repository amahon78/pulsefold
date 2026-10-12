# ADR 0011: Go as the Core Implementation Language

**Status:** Accepted

## Context

PulseFold V0 requires a small, bounded analyser that consumes OpenTelemetry log data without entering the production forwarding path. The project needs a supported implementation language and module identity before it can create a reproducible implementation, test suite or benchmark harness.

## Decision

Use Go as PulseFold's core implementation language. The canonical Go module path is `github.com/amahon78/pulsefold`.

Use the supported stable Go toolchain for development and CI. Use official OpenTelemetry Go libraries and generated protobuf types at the telemetry boundary where they meet the required contract.

## Rationale

- Go aligns strongly with the OpenTelemetry Collector ecosystem.
- It provides first-class OpenTelemetry Go libraries, pdata support and generated protobuf types.
- Its performance and concurrency model suit bounded telemetry processing.
- A compiled binary provides a simple deployment path.
- It permits PulseFold to reuse, benchmark and understand upstream Collector implementations without adding another runtime.
- It remains a suitable language if PulseFold later ships Collector components.

## Consequences

Positive:

- Phase 1 can use normal Go modules, tests, race detection and benchmark tooling.
- PulseFold can use the official OTLP protobuf API rather than defining a custom telemetry schema.
- Future Collector integration can share a language and telemetry ecosystem with the core.

Costs:

- Go is now a project-level technical commitment.
- OpenTelemetry Go API and version evolution become implementation dependencies.
- Contributors need Go familiarity.

This ADR does not decide whether PulseFold ships a Collector distribution, uses SQLite, selects a licence or changes V0's isolated shadow-analysis architecture.
