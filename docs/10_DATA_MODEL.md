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
Immutable provenance manifest used by future signals and investigations. Evidence content is separately lifecycle-managed.

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

Evidence manifests should use policy-scoped content commitments and lifecycle-managed content references. A raw content hash must not be assumed safe to retain or disclose.

An evidence reference should state whether the source is:
- inline
- local database
- object archive
- external observability backend
- cloud provider

An erasure or expiry event records a tombstone with its reason and time. The manifest remains resolvable as unavailable by policy.

## 5. Sensitivity

Proposed sensitivity classes:
- public
- internal
- confidential
- restricted

Model/provider policy can restrict which classes may cross an external boundary.

## 6. V0 report provenance

V0 requires only report provenance: source partition, observed time range, redaction-policy version, analyser version, candidate-policy version, coverage counters and lineage references. V0 does not require an evidence-content store.

## 7. Implemented V0 report fields

The initial JSON report has schema version `pulsefold/v0alpha1`. PF-1005 adds fields compatibly without changing the semantics of existing fields. It records source identity and delivery class, analysis interval, input coverage, redaction-policy version, analyser version and C1/C2 candidate-policy versions. Input coverage now marks whether captured bytes and observed-record denominators are known. C1 fields include a `counterfactual` classification, records presented and analysed, unique identities, duplicate groups and records, counterfactual record and canonical-byte counts, exclusions and state-cap status. C2 fields include an `inferred` classification, C1 representatives, inferred and source-template candidate group counts, structural counterfactual records, partition count, configured caps and explicit exclusions.

The additive `accounting` object records input, C1 and C2 coverage ratios as exact numerator and denominator pairs, never floating-point percentages. It records `unknown` wire bytes unless an adapter supplies a concrete measurement. Its optional billing object records a `user_supplied` generic meter model, baseline units and any C1/C2 gross observability cost opportunity in integer currency minor units. An unavailable billing object provides a stable reason rather than a zero value. A C2 coverage exclusion is counted in C1 representatives, while the C2 analysis object retains weighted record exclusions.

The report contains no telemetry body, attribute value, canonical identity, fingerprint, template string, extracted parameter or raw error text. Its C1/C2 statement explicitly describes a counterfactual reduction opportunity, not observed financial saving or a transformation decision.
