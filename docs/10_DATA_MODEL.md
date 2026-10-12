# Canonical Data Model

## 1. Goals

Data structures must allow:
- provider independence
- evidence traceability
- deterministic replay where feasible
- explicit sensitivity
- incident correlation

## 2. Core entities

### TelemetryEnvelope
References one log, metric point, span or provider event in canonical form.

### Fingerprint
Stable identity for an exact or structural pattern.

### Template
Learned or configured event structure.

### CompactionGroup
Aggregation of repeated telemetry under a defined policy.

### Evidence
Immutable reference used by signals and investigations.

### Signal
Machine observation that may warrant attention.

### Incident
Stateful operational event composed of signals and evidence.

### Investigation
Recorded reasoning activity associated with an incident.

### Recommendation
Proposed action, explicitly separate from observation.

## 3. Identity rules

IDs should be:
- opaque enough to avoid embedding secrets
- globally unique within an installation
- stable where stability is useful
- sortable by time only if there is clear operational value

## 4. Evidence references

Evidence should prefer immutable content references.

An evidence reference should state whether the source is:
- inline
- local database
- object archive
- external observability backend
- cloud provider

## 5. Sensitivity

Proposed sensitivity classes:
- public
- internal
- confidential
- restricted

Model/provider policy can restrict which classes may cross an external boundary.
