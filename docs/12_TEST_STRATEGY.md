# Test Strategy

## 1. Pyramid

### Unit
Pure algorithms, parsers, policies, state transitions.

### Property
Compaction invariants, hash behaviour, round-trip/reference behaviour.

### Integration
OTLP receive/export, storage, webhooks, provider adapters.

### Scenario
Synthetic end-to-end telemetry and incident flows.

### Adversarial
Malformed telemetry, injection attempts, secrets, high cardinality.

### Performance
Sustained throughput and burst behaviour.

## 2. Required properties

Examples:
- shadow mode output equals input semantically
- unique protected events are preserved
- incident state transitions are valid
- evidence IDs resolve or fail explicitly
- model provider failure cannot block telemetry plane
- notification failure cannot block telemetry plane

## 3. Golden fixtures

Use versioned fixture sets for:
- structured logs
- unstructured logs
- stack traces
- high cardinality fields
- repeated large payloads
- interleaved services
- multi environment events

## 4. Regression policy

Every discovered data loss bug should gain a permanent regression fixture.

Every incorrect incident RCA used in a public benchmark should be retained as an evaluation case where licensing permits.
