# Architecture

## 1. Architectural style

PulseFold is designed as a pipeline with explicit boundaries between:

1. telemetry transport
2. telemetry interpretation
3. telemetry compaction
4. signal detection
5. incident state
6. evidence storage/reference
7. investigation
8. routing

No layer should require AI merely because a later layer can use AI.

## 2. V0 architecture

```text
Offline OTLP replay or explicit best-effort replica
                         |
                         v
  redaction -> canonicalisation -> bounded C1/C2 analysis -> report manifest
                         |                                      |
                         +---------- coverage accounting --------+

Production forwarding remains outside the V0 process and is not modified.
```

V0 is logs-only. C1 and C2 are report-only candidates, not downstream transformations. V0 does not contain an incident engine, notification router or built-in investigator.

The current Phase 1 slice implements C1, C2 and report-only accounting. It uses offline OTLP/JSON replay, a mandatory redaction boundary, bounded exact-duplicate analysis and bounded structural-candidate analysis in one local process. C2 receives only C1 representatives from the redacted batch, partitions by full redacted resource and scope context plus severity, and emits no templates or extracted parameters. This process is an analyser, not a forwarding component.

Accounting distinguishes measured replay capture from analysis. A denominator is available only after the replay has completed its bounded read and record enumeration. C1 coverage is measured against accepted records. C2 coverage is measured against C1 representatives, while its existing exclusion counters remain weighted record counts. C1 canonical-byte reduction is counterfactual, C2 structural reduction is inferred, and neither changes telemetry. Wire and compressed transport bytes are unknown in the current implementation.

An optional billing model is an offline, vendor-neutral contract mapping. It supports complete captured-input bytes, accepted records and C1 canonical bytes, with an optional indexed fraction and fixed or marginal tiered integer pricing. It also recognises wire bytes and retained byte-days so it can explicitly report that PF-1005 has not measured them. The model never maps canonical bytes to wire bytes. It can report gross observability cost opportunity for a stated scope, but cannot calculate retention, egress, PulseFold operating cost or net customer saving.

PF-1007 presents these existing facts through one terminal or JSON shadow report. It keeps measured input, C1 counterfactual exact repetition, C2 inferred structural repetition, economics and unavailable values separate. A deterministic limitations list is committed with the report, so a renderer cannot silently omit the constraints under which an opportunity was calculated.

## 3. Longer-term logical architecture

The longer-term design retains explicit boundaries between telemetry transport, telemetry interpretation, deterministic compaction, signal detection, incident state, evidence manifests, investigation and routing. These capabilities are delivered in later phases only after the V0 safety and benchmark gates pass.

## 4. Telemetry plane versus control plane

### Telemetry plane

High throughput path.

Future responsibilities:
- receive and parse
- redact before derivation
- normalise and fingerprint
- apply deterministic policy
- forward and emit aggregate metadata

Requirements:
- bounded latency
- bounded memory
- graceful backpressure
- explicitly tested forwarding behaviour where configured
- no required external LLM call

### Control/intelligence plane

Lower throughput path.

Responsibilities:
- analyse groups
- detect signals
- maintain incidents
- retrieve evidence
- call external APIs
- invoke models/agents
- notify users

A slow model provider must not block telemetry flow.

V0 analysis is a separate process and has no dependency on this future control plane.

## 5. Canonical telemetry

OpenTelemetry semantic structures are the preferred canonical format.

Provider adapters translate cloud or vendor specific representations at the edge.

Core components should consume:
- logs
- metrics
- traces
- resource attributes
- scope attributes

## 6. Compaction pipeline

Proposed stages:

### Stage C0: pass through
No transformation.

V0 uses C0 only for its analysis input. It does not forward telemetry.

### Stage C1: exact duplicate aggregation
Group identical events within a bounded time window.

V0 reports opportunities only. Any future active C1 policy must use collision-safe identity, full equality verification and bounded state.

### Stage C2: template aggregation
Identify stable message structure and variable parameters.

V0 reports source-provided and inferred template candidate counts only. The PF-1004 inferred candidate is a conservative Drain-style token normaliser, evaluated against `github.com/jaeyo/go-drain3` v0.1.2, the implementation wrapped by the upstream Collector Drain processor, but intentionally not a Collector processor. The baseline test demonstrates that its default similarity threshold can generalise distinct error codes, so C2 keeps error and status codes literal. C2 has explicit global cluster, partition, byte, body and token caps, does not extract raw parameters and does not persist learner state. Structured and oversized bodies are visible exclusions. A template is not globally stable evidence identity and cannot authorise deletion.

### Stage C3: repeated payload extraction
Detect repeated large payload sections and replace internal copies with a content reference for PulseFold analysis.

Whether downstream output may reference externalised payloads is a separate policy decision.

### Stage C4: adaptive policy
Potential future stage based on event rarity, value and incident context.

C4 is not part of V0 and must not be activated without evidence preservation guarantees.

## 7. Evidence layer

An evidence manifest is immutable once emitted. Evidence content is separately lifecycle-managed and may expire or be erased under policy.

Suggested fields:

```text
evidence_id
evidence_type
source
observed_at
resource_identity
content_commitment
summary
content_reference
transformation_reference
sensitivity_class
```

V0 emits report provenance only: a payload-free manifest with fixed lineage, `not_stored` content state and an unauthenticated SHA-256 integrity commitment (see `docs/10_DATA_MODEL.md` section 8). Content storage is not a V0 component.

## 8. Signal engine

Two categories:

### Deterministic signals
- threshold
- ratio
- rate of change
- new template
- recurrence burst
- saturation
- SLO burn
- missing expected heartbeat
- known fingerprint

### Higher order signals
Correlate multiple weak signals into a candidate.

Higher order detection may use statistical or model based methods later but must output the underlying observations.

## 9. Incident engine

Not a V0 component. It remains a Phase 3 design and implementation area.

An incident is a stateful object, not a notification.

Signals may:
- open an incident
- update an incident
- increase/decrease severity
- attach evidence
- attach affected entities
- trigger investigation
- resolve an incident

## 10. Investigation gateway

Two modes:

### Built in investigator

Deferred until after external-agent integration and evaluation. A future built-in investigator may supply:
- investigation orchestration
- read only tools
- evidence constrained output
- model provider adapter

### External agent

Future PulseFold emits:
- incident
- context
- evidence references
- investigation request
- callbacks/webhooks for updates

This supports existing organisational agents such as internal Grok, Claude or custom bots.

## 11. Notification router

Not a V0 component. It remains a later incident-intelligence capability.

Notifications consume incident events, not raw telemetry events.

Examples:
- incident opened
- severity changed
- investigation complete
- recommendation available
- incident recovered
- incident resolved

## 12. Failure philosophy

PulseFold should never become a hidden single point of telemetry loss. V0 avoids owning the forwarding path rather than claiming transparent fail-open forwarding.

Required design work:
- queues with bounded behaviour
- backpressure
- pass-through/fail-open mode
- health endpoints
- dead letter strategy for intelligence plane work
- explicit degraded mode
- operator visible dropped-event counters
