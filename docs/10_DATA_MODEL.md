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

PF-1007 adds a deterministic top-level `limitations` array. It is derived from status, knownness, exclusions, wire measurement and billing availability. It describes the limits of the individual report, including any incomplete coverage, rather than inventing a combined confidence or reduction number. It is covered by the integrity commitment.

## 8. Implemented V0 evidence manifest (PF-1006)

Every report, including `input_failed` and `analysis_failed` reports, carries an additive `manifest` object with schema `pulsefold/evidence-manifest/v0alpha1`. It records the manifest, report and analyser versions; opaque source identity and partition references; delivery class (`kind`, `non_authoritative`); the versioned input format; the source-observed analysis interval; completion state and static failure code; coverage knownness; content lifecycle state; fixed lineage; an integrity commitment; and a volatile `created_at`.

Lineage is the fixed V0 pipeline `capture -> redaction -> c1 -> c2 -> accounting -> report`, not a generic graph. Stage entries record the redaction-policy version, C1 canonicalisation and candidate-policy versions, C2 algorithm, candidate-policy and partition-policy versions and bounds, the accounting policy version and the billing-model schema version and commitment when an operator supplies a model. Stage origins are `system_generated`, `operator_supplied`, `telemetry_derived` or `unknown`. Stage classifications reuse the PF-1005 classes (`measured`, `counterfactual`, `inferred`, `user_supplied`, `unknown`); there is no separate classification model. A stage that did not run carries no versions or classification.

The manifest contains no telemetry body, attribute value, canonical identity, template text, secret, content locator or local filesystem path.

### Identifiers

| Field | Stable across replay | Unique per run | Globally unique | Opaque | Content-derived |
| --- | --- | --- | --- | --- | --- |
| `source.identity_reference`, `source.partition_reference` | yes | no | no | yes | yes: SHA-256 of a namespaced operator label |
| `manifest_id` (`pfm_v0_<hex>`) | yes | no | no | yes | yes: derived from the integrity commitment |
| `lineage.*.reference` | yes | no | no, scoped to one manifest | no | no: fixed stage names |
| `billing_model_commitment` | yes | no | no | yes | yes: SHA-256 of the validated model |

Source references are opaque but are not confidential: a low-entropy label can be confirmed by guessing. They are derived only from already-validated operator labels, never from telemetry. Two identical replays produce the same `manifest_id`; it identifies report content, not a run. V0 has no per-run identifier.

### Integrity

`integrity` uses SHA-256 with canonicalisation `pulsefold/report-content-json-v1`: the compact `encoding/json` encoding of this schema version's typed structs (declaration field order, no maps, UTC RFC 3339 times) of the report with its manifest replaced by a view that clears `integrity`, `manifest_id` and `created_at`. The commitment is therefore not self-referential and excludes wall-clock time. It covers aggregate report content only, which is already post-redaction; it is not a commitment to raw telemetry and two inputs with identical aggregate results share a commitment. `authenticated` is always `false`: anyone able to change a report can recompute the commitment. V0 has no signing, PKI, key management or attestation.

### Deterministic and volatile fields

Every report and manifest field is deterministic for identical input, configuration and analyser version except `manifest.created_at`. PulseFold does not claim byte-for-byte reproducibility across analyser, Go or dependency versions.

The C2 `additional_counterfactual_records` value counts analysed C1 representatives that collapse into an existing inferred group. It never re-counts C1 exact duplicates (candidate policy `c2-report-only-no-parameters-v2`; v1 reports over-stated it by the C1 reduction). The priced C2 potential is measured over the C1-weighted candidate subset and overlaps the C1 opportunity; a committed limitation says so whenever both are priced. See `docs/research/grok-v0-adversarial-review.md`.

### Content lifecycle

V0 stores no evidence content and always emits `content.state: not_stored`. The lifecycle record type supports `not_stored`, `available`, `expired` and `erased` with a time, a short machine reason code for `expired` and `erased`, and an opaque reference only for `available`. Lifecycle records are separate from the immutable manifest and do not change its commitment. Only `available` with a reference is reported as retrievable, and V0 has no backing store to resolve one.

### Source labels

Source identity and partition labels are limited to 128 ASCII letters, digits, `.`, `_` and `-`. Paths, URLs, query strings, whitespace, control characters, traversal-shaped labels and common credential-shaped labels are refused without being echoed. The CLI uses the input file base name when it is a safe label and `offline-replay-file` otherwise.
