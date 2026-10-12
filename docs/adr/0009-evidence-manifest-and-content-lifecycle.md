# ADR 0009: Evidence Manifest and Content Lifecycle

**Status:** Accepted

## Context

Evidence provenance must remain auditable, while telemetry content can require retention expiry, erasure or access restriction. Treating raw content as immutable conflicts with those obligations.

## Decision

Evidence uses an immutable, small provenance manifest and separately managed content.

The manifest records an opaque evidence ID, observed time range, source partition, redaction-policy version, analyser version, transformation-policy version, policy-scoped content commitment, sensitivity class and lineage references. Content is optional, encrypted where stored and governed by an explicit retention policy.

Erasure or expiry creates an append-only tombstone containing the reason and time. The content is removed or made irrecoverable, while the manifest resolves as unavailable by policy.

## Consequences

Positive:

- provenance survives content lifecycle changes
- conclusions can cite stable evidence IDs without embedding raw telemetry
- retention and erasure are explicit and auditable

Negative:

- this is not a legal-compliance guarantee
- content storage, key management and retention remain open implementation decisions
- content commitments must not leak sensitive values or enable unintended correlation
