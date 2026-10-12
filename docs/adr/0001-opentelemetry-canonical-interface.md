# ADR 0001: OpenTelemetry as the Canonical Telemetry Interface

**Status:** Accepted

## Context

PulseFold must remain cloud and observability vendor agnostic.

## Decision

Use OpenTelemetry representations and OTLP as the preferred canonical interface between telemetry sources, PulseFold and compatible destinations.

Cloud/vendor specific integrations belong at adapters or edges.

## Consequences

Positive:
- broad interoperability
- existing receiver/exporter ecosystem
- easier vendor neutrality
- less custom protocol design

Negative:
- provider specific concepts may not map perfectly
- OpenTelemetry evolution becomes a dependency
- some integrations will still require adapters
