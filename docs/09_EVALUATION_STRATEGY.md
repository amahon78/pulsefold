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

V0 evaluates reduction opportunity, not generic dollar savings. PF-1005 tests complete and partial capture denominators, malformed and rejected input, C1 and C2 exclusions, zero input, fully repeated input, deterministic coverage, unknown wire measurement, absent and incompatible billing models, indexed subsets, fixed and tiered integer pricing and arithmetic overflow. It must not extrapolate a partial analysis to the full input.

The implementation can calculate a gross observability cost opportunity only for an operator-supplied generic meter mapping. It does not have retention assumptions, PulseFold operating cost, egress measurements or vendor contract integration, so it cannot calculate net customer saving.

### V0 shadow report

PF-1007 verifies the complete replay-to-report path for complete, zero-input, malformed and bounded-input outcomes. The report must retain C1 and C2 as separate claims, show the relevant coverage denominator, preserve unavailable economics and verify its manifest commitment after JSON round trip. Terminal output is inspected for the same safety distinctions and for absence of fixture secrets or input paths. These are report-correctness checks, not evidence that active compaction is safe or that an estimate will match a production bill.

### PF-1008 laboratory harness

Run the deterministic fault matrix with `go test ./internal/harness -run TestFaultMatrix -v`. Its verbose output contains a small machine-readable result with schema `pulsefold/v0-fault-matrix/v1`. The matrix drives the full offline pipeline through empty, malformed, truncated, bounded, unsafe-source and adversarial payload cases, C1 and C2 state exclusions, compatible and unavailable billing mappings, privacy checks, provenance verification and text-renderer failure.

Run the performance matrix with `go test ./internal/harness -run '^$' -bench BenchmarkV0Pipeline -benchmem`. It includes small and medium unique and repetitive workloads, high-cardinality records, large bodies, a C2-heavy fixture and mixed synthetic telemetry. `BenchmarkV0PipelineScaling` uses 1,024, 2,048 and 4,096 deterministic unique-record workloads. Existing component benchmarks isolate C1, C2, accounting, manifest and integrity work.

On 2026-10-13, an Apple M1 Pro running macOS 26.7.1 and Go 1.27.2 measured the complete matrix at roughly 0.47 ms to 108 ms per operation, depending on workload. The scaling series measured 26.2 ms, 51.8 ms and 100.4 ms for 1,024, 2,048 and 4,096 unique records, with throughput near 8 MB/s. These are machine-specific synthetic observations, not regression thresholds, complexity proofs or deployment sizing guidance. A 4,096-high-cardinality resource observation allocated about 21 MB during the run while C1 and C2 state stayed below their configured 1 MiB caps. This is not a whole-process memory ceiling.

The harness does not test a live replica, forwarding backpressure, process restart, persistent-state recovery, production retention or destructive compaction. It verifies only the stateless, offline V0 analyser behaviour.

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
