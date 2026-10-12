# Decision Log

This is a human readable index. Formal architecture decisions live in `docs/adr`.

## Confirmed

### D-001
PulseFold is model agnostic.

### D-002
PulseFold is cloud agnostic.

### D-003
OpenTelemetry is the canonical telemetry interface.

### D-004
PulseFold must provide value without AI.

### D-005
The product is proactive, not query only.

### D-006
PulseFold routes incident intelligence to humans and external agents.

### D-007
The initial agent posture is read only.

### D-008
Evidence is required for factual incident conclusions.

### D-009
Shadow mode is the initial adoption path.

### D-010
AI does not make destructive retention decisions initially.

### D-011
V0 is a redacted, coverage-aware, logs-only shadow reduction analyser using offline OTLP replay or an explicitly best-effort replicated, non-authoritative stream.

### D-012
V0 does not modify production forwarding, perform active compaction, run incident state or routing, invoke a built-in investigator or claim generic dollar savings.

### D-013
Shadow analysis and production forwarding run in separate failure domains. Shadow reports disclose coverage and analysis loss.

### D-014
Redaction precedes canonicalisation and all telemetry derivation.

### D-015
Evidence manifests are immutable while content follows a retention and erasure lifecycle.

### D-016
External investigation contracts precede a built-in investigator.

## Pending

### P-001
Go as core implementation language.

### P-002
Collector distribution versus standalone gateway.

### P-003
Apache 2.0 licence.

### P-004
Initial local state store.

### P-005
First supported notification adapter after webhook.

### P-006
First model/provider adapters.

### P-007
The implementation of evidence content storage, key management and retention.
