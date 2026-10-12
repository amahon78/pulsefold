# Evaluation Strategy

## 1. Principle

PulseFold should be measurable.

"Looks good in a demo" is not an acceptance criterion.

## 2. Telemetry evaluation

### Reduction
- raw bytes
- output bytes
- raw event count
- output event count
- reduction by technique

### Information preservation
- unique event preservation
- error event preservation
- template purity
- parameter recovery
- representative sample coverage
- evidence reference validity

### Performance
- events/sec
- MB/sec
- p50/p95/p99 added latency
- memory
- CPU
- queue depth
- behaviour under burst

### Failure
- exporter unavailable
- storage unavailable
- malformed telemetry
- extremely high cardinality
- giant payload
- clock skew
- process restart

## 3. Signal evaluation

Known synthetic streams should measure:
- detector precision
- detector recall
- detection delay
- duplicate signal rate
- suppression correctness

## 4. Incident evaluation

Ground truth scenarios:
- bad deployment
- database connection exhaustion
- dependency outage
- DNS failure
- expired credential
- rate limiting
- queue backlog
- memory leak
- CPU saturation
- schema mismatch/migration
- permission regression
- noisy neighbour
- latency degradation
- retry storm

Measure:
- correct incident grouping
- incorrect merges
- incorrect splits
- time to open
- time to resolve

## 5. Agent evaluation

For each model/provider:

- root cause accuracy
- evidence precision
- evidence recall where measurable
- unsupported claim rate
- useful recommendation rate
- confidence calibration
- tool calls
- investigation latency
- input/output tokens
- monetary cost
- redaction failures
- prompt injection success

## 6. Comparative evaluation

PulseFold should support reproducible comparison:

```text
Model A vs Model B
Detector X vs Detector Y
Compactor X vs Compactor Y
Raw context vs compacted context
```

## 7. Release gates

No agent feature should graduate from experimental without:
- benchmark corpus
- known ground truth
- published scoring method
- failure examples
- repeatable run command
