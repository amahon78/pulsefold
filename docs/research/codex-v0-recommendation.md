# PF-0006: Smallest credible V0 recommendation

**Reviewer:** Codex
**Date:** 2026-10-12
**Recommendation status:** Requires Alex's approval before PF-0007 freezes scope.

## Recommendation

V0 should be a **redacted, coverage-aware shadow reduction analyser**. It accepts an OTLP replay or an explicitly replicated, non-authoritative OTLP stream, produces a reproducible report and never transforms the telemetry sent to the existing observability backend.

This is deliberately narrower than the current logical architecture. It validates the first technical problem, safe telemetry optimisation, without making PulseFold a production-critical forwarding, incident-management or AI system.

## V0 scope

| Include | Exclude or defer |
|---|---|
| OTLP logs input for offline replay and a documented best-effort replicated stream | Active compaction, filtering, sampling or payload externalisation |
| Explicit redaction policy before all derivation | Metrics and traces beyond transport accounting |
| Bounded exact-duplicate opportunity analysis, with collision-safe identity design | An in-process promise that analysis failure cannot affect production forwarding |
| Drain or source-provided templates as report-only candidates, with limits and uncertainty | Template-driven deletion or aggregation |
| Versioned JSON report containing raw bytes, estimated output bytes, wire-byte baseline when measurable, coverage and safety exclusions | Generic savings percentages or dollar-saving claims without a customer billing model |
| Synthetic, truth-labelled fixtures and fault tests | Raw evidence archive, content-addressed payload store and reconstruction UI |
| A written external event/evidence schema draft | Incident state machine, correlation engine, Slack adapter, routing UI or built-in investigator |
| Optional benchmark of upstream `log_dedup`, `drain` and OTel Arrow against the same corpus | A custom Collector distribution unless it becomes necessary for the analyser integration |

## Deployment shape

```text
V0 forwarding remains stock, operator-owned Collector configuration
  application -> Collector -> existing backend

V0 analyser receives either:
  a) offline OTLP replay, or
  b) an explicitly replicated best-effort stream in a separate process

  replicated/replayed data -> redact -> bounded analyser -> JSON report
```

This design honours shadow semantics: the backend receives the unmodified payload from its existing forwarding configuration. It does not claim that a production traffic mirror is free, complete or isolated. Those properties must be measured and reported as `coverage`, `capture drops` and `analysis failures`.

## Report contract

Each report should record:

1. Input identity: source partition, UTC interval, input record count and raw canonical byte count.
2. Configuration identity: analyser version, redaction-policy version, canonicalisation version and candidate policy settings.
3. Coverage: accepted records, rejected or malformed records, replication drops, state overflows and any unsupported values.
4. C1 opportunity: candidate group count, records represented, counterfactual output records, raw-byte delta and groups excluded for safety.
5. C2 opportunity: source-provided templates separately from inferred templates, template confidence and unsafe-merge test status. No conversion into a drop recommendation by default.
6. Economics: compressed-wire measurement only when the exact protocol, batching and compression are known; billing estimate only when the operator supplies a billing mapping and cost inputs.
7. Safety stratification: error severity, novel templates, high-cardinality dimensions, tainted fields and redacted input. These must be shown, not hidden within a global percentage.

## Definition of credible completion

V0 is ready to freeze only if its benchmark harness demonstrates the following on committed synthetic data and a separately fetched, licence-reviewed corpus:

- output is reproducible for a fixed input, configuration and seed
- every state cap produces a visible report exclusion or degradation marker
- no record on the forwarding path is changed by V0
- tests cover blocked analysis output, restart, malformed input, high cardinality, giant payloads, secret-shaped values and telemetry prompt-injection strings
- measured C1 and C2 results include false/unsafe merge rate, novel/error preservation, CPU, heap, throughput and p99 processing latency
- the report distinguishes an opportunity estimate from an actual observed financial saving

## Decision request for Alex

Approve or amend the following before PF-0007:

1. Freeze V0 to logs-only shadow analysis and an offline or best-effort replicated input, not an inline production gateway.
2. Adopt the four proposed ADR topics in [the architecture review](codex-architecture-review.md#proposed-adrs-without-changing-accepted-adrs) before implementation.
3. Treat external agents, including HolmesGPT Operator, as optional future consumers. Do not build an investigator, incident engine or notification router in V0.
4. Make redaction-before-derivation and bounded-state reporting non-negotiable V0 safety gates.
5. Require a customer-specific billing mapping before presenting any dollar-saving calculation.

## Why this still supports the long-term vision

The V0 artefact establishes the interfaces that later phases need: canonical telemetry input, redaction, deterministic grouping candidates, quantified safety exclusions, provenance and evidence-oriented reports. It does not prematurely choose a database, agent provider or incident workflow. Once this foundation has measured value, Phase 2 can consider active deterministic policies and an evidence lifecycle, and Phase 3 can add deterministic signals for external systems such as Keep, Alertmanager or HolmesGPT.
