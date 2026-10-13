# Evaluation Strategy

## 1. Principle

PulseFold should be measurable.

"Looks good in a demo" is not an acceptance criterion.

## 2. Telemetry evaluation

### Reduction
- raw canonical bytes
- counterfactual output bytes
- encoded and compressed wire bytes where measured
- contract-specific billable units where supplied by the operator
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
- report coverage, exclusion and overflow visibility

PF-1004 measures C2 against independently labelled synthetic groups with weighted pairwise precision, recall and F1, unsafe merge and split rates, novel and error preservation, source-template candidate coverage, deterministic shuffled-order agreement and two- and four-partition comparisons. It reports the limits of this synthetic measurement: C2 has no claim of permanent cross-instance convergence and public corpora remain subject to an explicit licence review.

### Performance
- events/sec
- MB/sec
- p50/p95/p99 added latency
- memory
- CPU
- queue depth
- behaviour under burst

### Failure
- blocked analysis output
- replay or replica-input interruption
- malformed telemetry
- extremely high cardinality
- giant payload
- clock skew
- process restart

### V0 economics

V0 evaluates reduction opportunity, not generic dollar savings. A monetary estimate requires an explicit customer billing model, retention assumptions and PulseFold operating costs.

## 3. Signal evaluation

Post-V0. V0 does not generate operational signals.

Known synthetic streams should measure:
- detector precision
- detector recall
- detection delay
- duplicate signal rate
- suppression correctness

## 4. Incident evaluation

Post-V0. V0 does not open or correlate incidents.

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

Post-V0. V0 invokes no models or agents.

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
