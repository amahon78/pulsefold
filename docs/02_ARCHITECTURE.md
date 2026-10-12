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

## 2. Logical architecture

```text
                 +-----------------------+
                 | Applications / Cloud  |
                 +-----------+-----------+
                             |
                          OTLP / adapters
                             |
                             v
+-----------------------------------------------------------+
|                    PulseFold Gateway                      |
|                                                           |
|  +------------+   +------------+   +------------------+   |
|  | Normaliser |-->| Fingerprint|-->| Compaction       |   |
|  +------------+   +------------+   | Analysis/Policy  |   |
|                                     +--------+---------+   |
|                                              |             |
|                             +----------------+---------+   |
|                             |                          |   |
|                             v                          v   |
|                       Signal Engine              OTLP Export|
|                             |                          |   |
|                             v                          v   |
|                       Incident Engine          Existing O11y|
|                             |                              |
|                             v                              |
|                       Evidence Layer                       |
|                             |                              |
|                +------------+------------+                 |
|                |                         |                 |
|                v                         v                 |
|        Notification Router       Investigation Gateway     |
+----------------+-------------------------+-----------------+
                 |                         |
       Slack/Teams/Webhook          Built-in / external agent
```

## 3. Telemetry plane versus control plane

### Telemetry plane

High throughput path.

Responsibilities:
- receive
- parse
- normalise
- fingerprint
- apply deterministic policy
- forward
- emit aggregate metadata

Requirements:
- bounded latency
- bounded memory
- graceful backpressure
- fail open where configured
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

## 4. Canonical telemetry

OpenTelemetry semantic structures are the preferred canonical format.

Provider adapters translate cloud or vendor specific representations at the edge.

Core components should consume:
- logs
- metrics
- traces
- resource attributes
- scope attributes

## 5. Compaction pipeline

Proposed stages:

### Stage C0: pass through
No transformation.

### Stage C1: exact duplicate aggregation
Group identical events within a bounded time window.

### Stage C2: template aggregation
Identify stable message structure and variable parameters.

### Stage C3: repeated payload extraction
Detect repeated large payload sections and replace internal copies with a content reference for PulseFold analysis.

Whether downstream output may reference externalised payloads is a separate policy decision.

### Stage C4: adaptive policy
Potential future stage based on event rarity, value and incident context.

C4 is not part of V0 and must not be activated without evidence preservation guarantees.

## 6. Evidence layer

An evidence item is immutable once emitted.

Suggested fields:

```text
evidence_id
evidence_type
source
observed_at
resource_identity
content_hash
summary
raw_reference
transformation_reference
sensitivity_class
```

Investigations cite evidence IDs rather than copying unbounded telemetry into conclusions.

## 7. Signal engine

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

## 8. Incident engine

An incident is a stateful object, not a notification.

Signals may:
- open an incident
- update an incident
- increase/decrease severity
- attach evidence
- attach affected entities
- trigger investigation
- resolve an incident

## 9. Investigation gateway

Two modes:

### Built in investigator

PulseFold supplies:
- investigation orchestration
- read only tools
- evidence constrained output
- model provider adapter

### External agent

PulseFold emits:
- incident
- context
- evidence references
- investigation request
- callbacks/webhooks for updates

This supports existing organisational agents such as internal Grok, Claude or custom bots.

## 10. Notification router

Notifications consume incident events, not raw telemetry events.

Examples:
- incident opened
- severity changed
- investigation complete
- recommendation available
- incident recovered
- incident resolved

## 11. Failure philosophy

PulseFold should never become a hidden single point of telemetry loss.

Required design work:
- queues with bounded behaviour
- backpressure
- pass-through/fail-open mode
- health endpoints
- dead letter strategy for intelligence plane work
- explicit degraded mode
- operator visible dropped-event counters
