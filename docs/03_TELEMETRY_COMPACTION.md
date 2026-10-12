# Telemetry Compaction Specification

## 1. Objective

Reduce repeated telemetry while preserving diagnostic information and maintaining operator trust.

The target is not maximum compression. The target is maximum **safe informational efficiency**.

## 2. Core rule

> Compress repetition, not uniqueness.

A novel error event should generally be preserved more aggressively than the ten-thousandth instance of an already understood event.

## 3. Event identity

PulseFold should maintain more than one identity:

### Exact identity
Hash over normalised stable event representation.

### Structural/template identity
Represents the event shape with variables removed.

Example:

```text
Payment timeout user=839281 request=abc123
Payment timeout user=239102 request=def993
```

Potential template:

```text
Payment timeout user={user_id} request={request_id}
```

### Context identity
Resource/service/version/environment dimensions may distinguish events that otherwise share a template.

## 4. Aggregated representation

A compacted event group should be able to represent:

```yaml
group_id: pfgrp_...
template_id: pftpl_...
count: 14832
first_seen: ...
last_seen: ...
dimensions:
  service:
    payments-api: 14832
  version:
    v193: 14417
    v192: 415
parameters:
  user_id:
    approximate_cardinality: 3812
  request_id:
    approximate_cardinality: 14832
samples:
  - evidence_id: ...
  - evidence_id: ...
raw_archive_ref: optional
policy:
  name: template-window
  version: 1
```

## 5. Representative samples

Sampling is for evidence quality, not only debugging convenience.

Candidate strategy should preserve:
- first event
- last event
- one or more reservoir samples
- examples from materially different dimensions
- examples around state changes

## 6. Repeated verbose payloads

Large repeated fields can dominate telemetry cost.

Examples:
- repeated stack trace
- repeated HTTP response body
- repeated exception chain
- repeated metadata blob

PulseFold should evaluate content addressed extraction:

```text
payload_hash -> one stored payload
events -> payload reference
```

Active transformation must account for whether the downstream destination requires original standalone events.

## 7. Time windows

Compaction needs bounded windows.

Questions to benchmark:
- fixed versus sliding windows
- per-template windows
- burst aware window sizing
- memory impact
- incident sensitivity

## 8. Safety invariants

1. No unique event may be classified as a duplicate solely because of a lossy hash.
2. Hash collisions must be handled safely.
3. Template mining errors must not merge unrelated severity classes by default.
4. Environment boundaries must not be collapsed silently.
5. Security/audit categories may be configured as never compact.
6. Compaction must be versioned and auditable.
7. Operators must be able to exclude fields from template variable detection.
8. Redaction and compaction ordering must be specified.

## 9. Shadow mode report

Minimum output:

- events observed
- bytes observed
- exact duplicate opportunities
- template aggregation opportunities
- repeated payload opportunities
- unique event volume
- top reducible templates
- estimated downstream reduction
- CPU/memory used by analysis
- confidence/limitations

## 10. Unknowns requiring Phase 0 research

- preferred template mining algorithm
- cardinality sketch implementation
- sample strategy
- template drift handling
- payload extraction semantics
- compressed representation when forwarding to arbitrary vendors
- relationship to OpenTelemetry collector processors
