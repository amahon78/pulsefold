# PulseFold

**Keep the evidence. Fold the noise.**

PulseFold is an open source, vendor neutral telemetry intelligence layer designed to reduce observability waste, detect emerging incidents proactively and give humans or AI agents the evidence they need to investigate production problems.

PulseFold is not intended to be another dashboard, another log database or another model specific SRE bot. It sits between telemetry producers and existing observability destinations, using OpenTelemetry as the canonical telemetry interface.

## Product thesis

Modern observability wastes money and engineering time in two places:

1. **Telemetry repetition:** large volumes of logs contain repeated structure, repeated payloads and repeated events that are expensive to index and retain.
2. **Incident investigation:** engineers manually reconstruct incidents by moving between logs, metrics, traces, deployments, source control and runbooks.

PulseFold addresses both.

```text
Applications / Infrastructure
            |
            v
      OpenTelemetry
            |
            v
        PulseFold
    +------------------+
    | normalise        |
    | fingerprint      |
    | compact          |
    | detect           |
    | correlate        |
    | preserve evidence|
    +------------------+
       |            |
       v            v
Existing backend   Incident intelligence
                   |
              +----+----+
              |         |
              v         v
           Humans     Agents
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

PulseFold V0 will answer:

> How much telemetry are you wasting, and what can be compacted safely?

PulseFold V1 will answer:

> Something changed in production. What happened, what evidence supports the diagnosis and who or what should be notified?

## Current status

**Planning and specification phase. No production code should be written until the Phase 0 exit criteria are met.**

See:

- `PROJECT_STATE.md`
- `TASK_BOARD.md`
- `AGENTS.md`
- `docs/00_MASTER_PLAN.md`
- `docs/01_PRODUCT_REQUIREMENTS.md`
- `docs/02_ARCHITECTURE.md`
- `docs/09_EVALUATION_STRATEGY.md`
