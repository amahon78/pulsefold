# ADR 0012: Destination Representation Preservation

**Status:** Accepted

## Context

Analytical redaction, canonicalisation and C2 inference are useful for shadow analysis but do not describe the telemetry a destination would have received. Lossy redaction can make distinct source records appear equal.

## Decision

For an active policy, the **destination representation** is the complete OTLP log representation that the configured downstream destination would have received without PulseFold's destructive transformation, after only the declared forwarding transformations. It includes resource and scope data and schema URLs, log timestamps and observed timestamps, severity, body, attributes, event name, flags, trace and span IDs, and all other serialised log fields.

A policy may replace N records with fewer outputs only when, for every destination-representation field, it is either preserved unchanged, retained losslessly in verified recoverable evidence for the declared recovery period, or named with its exact transformation in an operator-approved, versioned aggregation contract. Redacted equality, canonical-projection equality, template similarity and C2 inference are never sufficient destructive equality.

Initially, any record changed by lossy analytical redaction, or whose analytical and forwarding representations differ, is ineligible and bypasses transformation unchanged. Keyed equality tokens are deferred.

## Consequences

The first active policy is deliberately conservative. Timestamps and ordering are not silently replaced by a count. Recovery is required but permanent raw archival is not: the initial experiment uses a secondary path with the existing destination authoritative.
