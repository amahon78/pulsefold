# Test Strategy

## 1. Pyramid

### Unit
Pure algorithms, parsers, policies, state transitions.

### Property
Compaction invariants, hash behaviour, round-trip/reference behaviour.

### Integration
OTLP replay, report persistence and later storage, webhooks and provider adapters.

PF-1001 covers valid OTLP/JSON replay across resources and services, structured bodies, empty and malformed input, unsupported values and source contracts, deterministic coverage, bounded input and cancellation. It also verifies that malformed-input errors do not contain the supplied telemetry text.

PF-1002 adds seeded synthetic fixtures with independent C1 grouping truth, fake secret-shaped values, high-cardinality fields, large values and adversarial near-duplicates. Redaction tests verify that fixture secrets do not cross into cloned redacted records, canonical C1 identities, JSON reports or diagnostics.

PF-1003 tests C1 ordinary and absent duplicates, resource and severity boundaries, reordered structured values, string versus numeric distinction, forced hash-bucket collisions, entry and byte caps, oversized identities, high-cardinality inputs, cancellation and input immutability. Bounded fuzz targets exercise the OTLP/JSON boundary and C1 canonicalisation. Benchmarks cover repeated, unique, high-cardinality and large-body inputs.

PF-1004 adds independently labelled structural-template cases for variable user and request identifiers, UUIDs, timestamps, IPs, URLs, stack locations, JSON-like bodies, fake secrets, drift, service and severity boundaries, error codes and structured bodies. C2 tests verify redaction-only input, explicit structured and oversized exclusions, global cluster, partition and byte caps, cancellation, deterministic order behaviour, two- and four-partition conservative accounting, source-template accounting and bounded fuzzing. C2 benchmarks cover combined C1/C2 ordinary and high-cardinality paths.

PF-1005 adds coverage and economics tests for complete, partial, empty, malformed and bounded inputs, C1 and C2 exclusion coverage, deterministic reports, no raw telemetry leakage through accounting or configuration diagnostics, unknown wire measurement and absent billing models. Versioned synthetic billing fixtures cover byte, record, indexed-subset, tiered and retained-volume-unmeasured meters. Tests also cover zero input, fully repeated input, incompatible meters, invalid fractions and exact-arithmetic overflow. A bounded fuzz target exercises strict billing-model parsing, and a benchmark measures incremental accounting allocation cost.

PF-1006 tests complete, partial, input-failed and analysis-failed manifests with and without billing, C1 and C2 classifications, all four content lifecycle states, lineage versions, deterministic replay, identifier semantics, integrity detection of report and manifest changes with `created_at` excluded, strict JSON round trip, malformed manifests and lifecycle records, report ownership and aliasing, and source-label privacy for Unix, Windows and relative paths, credential URLs, query strings, tokens, control characters, oversized labels and traversal. Bounded fuzz targets exercise manifest parsing and source-label validation, and benchmarks measure manifest construction and integrity verification.

PF-1007 adds end-to-end text and JSON report tests for complete structural fixtures, malformed input, source-label privacy, billing availability, C1/C2 separation, integrity status and static limitations. It verifies that terminal output remains payload-free and never presents integrity as authentication. A complete report-path benchmark covers replay, redaction, C1, C2, report construction and text rendering.

PF-1008 adds the deterministic `internal/harness` fault matrix across the complete V0 pipeline. It covers input failures, source-label attacks, replay-limit boundaries, secret and prompt-shaped content, C1 collision and caps, C2 caps and adversarial grouping, accounting availability, report integrity, renderer failure, replay isolation and payload-free output. Its eight-workload performance matrix and three-size scaling series report local allocations and throughput without machine-specific pass/fail thresholds.

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
