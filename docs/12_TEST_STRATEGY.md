# Test Strategy

## 1. Pyramid

### Unit
Pure algorithms, parsers, policies, state transitions.

### Property
Compaction invariants, hash behaviour, round-trip/reference behaviour.

### Integration
OTLP replay, report persistence and later storage, webhooks and provider adapters.

### Scenario
Synthetic end-to-end replay and report flows. Incident flows are post-V0.

### Adversarial
Malformed telemetry, injection attempts, secrets, high cardinality.

### Performance
Sustained throughput and burst behaviour.

## 2. Required properties

Examples:
- V0 does not modify production forwarding
- replay input is redacted before derived output
- every capture drop, malformed record, refusal and state overflow is reported
- exact-duplicate opportunity analysis does not classify records as equal solely on a lossy hash
- unique protected events are excluded from transformation recommendations
- reports distinguish raw canonical bytes, measured wire bytes and billing estimates
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
- coverage loss and state-cap overflow
- secret-shaped bodies and attributes
- telemetry prompt-injection strings

## 4. Regression policy

Every discovered data loss bug should gain a permanent regression fixture.

Every incorrect incident RCA used in a public benchmark should be retained as an evaluation case where licensing permits.

## 5. V0 fault tests

V0 must test blocked analysis output, replay interruption, restart, malformed records, high cardinality, giant payloads and redaction failures. Active compaction, incident state, notifications and models are outside the V0 test surface.
