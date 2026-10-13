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
Canonical, redacted stable event representation. A hash may be used only for bucketing; an active compaction policy must verify full equality before classifying records as duplicates.

PF-1003 implements this as a report-only C1 identity. It includes redacted resource and scope attributes, schema URLs, scope name and version, severity, body, log attributes, event name, flags, trace ID and span ID. Attribute and structured-body key order are canonicalised, while value types remain distinct. Event and observed timestamps are intentionally excluded because recurring events are compared within one finite replay analysis window. A SHA-256 bucket is followed by byte-for-byte canonical equality verification.

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

PF-1004 implements a deliberately conservative, report-only C2 candidate. It processes only C1 representatives from redacted input. It partitions candidates by complete redacted resource and scope context plus severity, preserves ordinary attributes as part of the candidate shape and masks only strongly recognisable identifier, UUID, IP, timestamp and source-location forms. Error and status codes remain literal. It does not extract, retain or report raw parameter values, template strings or persistent learner state. Structured bodies, bodies over the configured limit, token overflows and state overflows are exclusions rather than candidate groups.

### Context identity
Resource/service/version/environment dimensions may distinguish events that otherwise share a template.

## 4. Aggregated representation

This representation is post-V0. V0 reports counterfactual groups and does not emit compacted downstream events.

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

Selection must not be predictable from record position alone. Future sampling should include randomised reservoir selection to reduce sample-poisoning risk.

## 6. Repeated verbose payloads

Large repeated fields can dominate telemetry cost.

Examples:
- repeated stack trace
- repeated HTTP response body
- repeated exception chain
- repeated metadata blob

PulseFold may evaluate content addressed extraction after V0:

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
- hard entry and byte caps
- overflow behaviour and report coverage

## 8. Safety invariants

1. No unique event may be classified as a duplicate solely because of a lossy hash.
2. Hash collisions must be handled safely.
3. Template mining errors must not merge unrelated severity classes by default.
4. Environment boundaries must not be collapsed silently.
5. Security/audit categories may be configured as never compact.
6. Compaction must be versioned and auditable.
7. Operators must be able to exclude fields from template variable detection.
8. Redaction precedes canonicalisation, fingerprinting, template training, snapshots, reports and external exposure.
9. Stateful analysis has hard time, entry and byte caps. Overflow must be observable and must not silently transform input.

## 9. Shadow mode report

Minimum output:

- events observed
- bytes observed
- input coverage, capture drops and analysis failures
- exact duplicate opportunities
- template aggregation opportunities
- repeated payload opportunities
- unique event volume
- top reducible templates
- counterfactual output and raw-byte reduction opportunity
- separately measured wire-byte estimate where configuration permits
- billing estimate only when an operator supplies the billing model and cost inputs
- CPU/memory used by analysis
- confidence/limitations

## 10. Unknowns requiring Phase 1 benchmark and design validation

- preferred template mining algorithm
- cardinality sketch implementation
- sample strategy
- template drift handling
- payload extraction semantics
- compressed representation when forwarding to arbitrary vendors
- relationship to OpenTelemetry collector processors

## 11. V0 boundary

V0 is logs-only and accepts replayed or explicitly best-effort replicated input. It performs redacted, bounded C1 and C2 opportunity analysis only. C3 payload extraction and all active transformation are deferred to Phase 2 after safety gates pass.

The current implementation supports C1, bounded report-only C2 and shadow accounting. The report labels C1 canonical-byte figures as `counterfactual` and C2 structural-record figures as `inferred`. Capture coverage is `measured` only when the bounded replay fully reads its input and enumerates records. It does not measure compressed wire bytes, vendor billable units or customer financial savings. C2 output is not an active-compaction policy, a global template identity or evidence content.

An optional operator-supplied billing model may price one explicitly mapped generic quantity using exact integer arithmetic: captured input bytes, accepted records or C1 canonical bytes. It also recognises wire bytes and retained byte-days, but reports them unavailable until those measures exist. Indexed subsets and fixed or marginal tiered prices are supported. Billing output identifies its captured-input or analysed-subset scope and reports gross observability cost opportunity only. Missing models, incomplete denominators, unsupported meters and overflow are unavailable states, not zero estimates. Canonical bytes, wire bytes, billable units and net saving remain distinct quantities.
