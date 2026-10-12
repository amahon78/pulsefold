# Phase 0 Research Index (Grok, PF-0001..PF-0005)

**Author:** Grok Bot (research agent), 2026-10-12 (AEST)
**Status:** Accepted as Phase 0 decision input on 2026-10-12. PF-0006 verification and PF-0007 reconciliation produced ADRs 0007 through 0010. The reports retain their original author attribution and confidence labels.

| Task | File |
|---|---|
| PF-0001 Competitive landscape | `competitive-landscape.md` |
| PF-0002 Log compaction algorithm review | `log-compaction-research.md` |
| PF-0003 OpenTelemetry extension strategy | `otel-extension-options.md` |
| PF-0004 Public benchmark corpus | `benchmark-corpus.md` |
| PF-0005 Threat model challenge | `threat-model-challenge.md` |
| Brief work package E (adversarial review) | `adversarial-architecture-review.md` |

## Conventions used in every file

- **FACT** = verifiable from the cited primary source on the date accessed (2026-10-12).
- **INFERENCE** = my reasoning from facts; may be wrong.
- **RECOMMENDATION** = an opinion for Alex/Codex to accept or reject.
- **Confidence:** High / Medium / Low.
- GitHub stars/licence/last-push figures were pulled from the GitHub REST API on 2026-10-12 and will drift.
- Vendor capability claims come from vendor docs/marketing; they are *vendor-reported*, not independently verified.

## Historical headline (read with PF-0006 verification)

The OpenTelemetry Collector contrib repository already contains, as of 2026, a **`logdedup` processor** (exact duplicate aggregation with counts and first/last timestamps), a **`drain` processor** (Drain template mining that annotates `log.record.template`, added April 2026), a **`cardinalityguardian` processor** (HLL++ based cardinality explosion detection), an **`isolationforest` processor** (inline anomaly scoring), **`count`/`sum`/`signal_to_metrics` connectors** (logs → metrics), **tail sampling / adaptive tail sampling**, **redaction**, and a work-in-progress **`telemetry_policy` processor**. Stages C1 and C2 of `02_ARCHITECTURE.md` are therefore substantially *existing OTel functionality*. PulseFold's defensible surface is narrower than the docs assume: the **shadow-mode safety/savings report**, **evidence-preserving (lossless-by-reference) compaction semantics**, and the **incident evidence contract for external agents**. See each file for detail.
