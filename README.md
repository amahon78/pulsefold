# PulseFold

**Keep the evidence. Fold the noise.**

PulseFold is an open source, vendor neutral telemetry intelligence layer designed to reduce observability waste, detect emerging incidents proactively and give humans or AI agents the evidence they need to investigate production problems.

PulseFold is not intended to be another dashboard, another log database or another model specific SRE bot. It uses OpenTelemetry as the canonical telemetry interface and is designed to complement existing observability destinations.

## Product thesis

Modern observability wastes money and engineering time in two places:

1. **Telemetry repetition:** large volumes of logs contain repeated structure, repeated payloads and repeated events that are expensive to index and retain.
2. **Incident investigation:** engineers manually reconstruct incidents by moving between logs, metrics, traces, deployments, source control and runbooks.

PulseFold addresses both.

```text
V0: OTLP replay or independent best-effort replica
            |
            v
  redact -> bounded analysis -> coverage-aware report

Future phases: deterministic optimisation -> signals -> evidence manifests
            -> optional external agents and incident destinations
```

## Core principles

- **Real operational value first.**
- **Cloud agnostic.**
- **Model agnostic.**
- **OpenTelemetry first.**
- **Read only agents by default.**
- **No LLM in the destructive data path.**
- **Evidence before conclusions.**
- **Shadow mode before compaction.**
- **Simple setup is a product requirement.**
- **Open source core.**
- **Synthetic and public data for development.**
- **Benchmarks and evaluations are part of the product.**

## Initial scope

PulseFold V0 is frozen as a redacted, coverage-aware, logs-only shadow reduction analyser. It accepts offline OTLP replay or an explicitly best-effort replicated, non-authoritative stream. It does not modify production forwarding, compact telemetry, run incidents or invoke AI.

V0 will answer:

> How much telemetry are you wasting, and what can be compacted safely?

PulseFold V1 will answer:

> Something changed in production. What happened, what evidence supports the diagnosis and who or what should be notified?

## Current status

**V0 scope is frozen. Phase 1 implementation is limited to the V0 analyser and its benchmark harness.**

The current implementation accepts one bounded OTLP/JSON `ExportLogsServiceRequest` for offline replay. It applies a versioned redaction policy before running bounded, collision-safe C1 exact-duplicate opportunity analysis, then emits a deterministic JSON report. It is an isolated analysis input only and has no network listener or production-forwarding role.

See:

- `PROJECT_STATE.md`
- `TASK_BOARD.md`
- `AGENTS.md`
- `docs/00_MASTER_PLAN.md`
- `docs/01_PRODUCT_REQUIREMENTS.md`
- `docs/02_ARCHITECTURE.md`
- `docs/09_EVALUATION_STRATEGY.md`
