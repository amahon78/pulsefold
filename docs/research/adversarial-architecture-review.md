# Adversarial Architecture Review (Brief work package E)

**Status:** Draft (Grok, 2026-10-12). Deliberately argues *against* the thesis. Input to Codex PF-0006. No architecture doc changed.

## 1. Attempted disproof of the thesis

Thesis (README): telemetry repetition and manual incident reconstruction are both wasteful; one product placed between producers and backends should address both.

**Argument against, part 1 — the telemetry half is already commodity.** Stock `otelcol-contrib` provides exact dedup (`logdedup`), Drain templates (`drain`), filtering, OTTL transforms, redaction, sampling, cardinality guarding and logs→metrics connectors (see competitive-landscape.md). Sawmills and Edge Delta productise this with UIs and recommendations. Confidence High.

**Part 2 — the incident half is crowded and owned by workflow incumbents.** incident.io, Rootly, PagerDuty, Datadog Bits, Resolve, Traversal, plus OSS HolmesGPT (CNCF), k8sgpt, Keep, Coroot. Confidence High.

**Part 3 — the combination is weaker than it looks.** The two halves have different buyers (FinOps/platform vs on-call), different latency needs, and different trust models. Combining them doubles the trust surface (08) and the scope. A pre-backend layer *lacks* the context investigators need (it doesn't hold history; doc says PulseFold isn't a primary DB) and must query back into the backend anyway — so its investigation position is no better than a tool sitting beside the backend. Confidence Medium.

**What survives the attack (INFERENCE, Medium):**
- An open, measurable **"safe reduction" methodology and shadow report** — nobody publishes reproducible numbers with preserved-event guarantees.
- Being the **evidence-reference layer**: stable evidence IDs + compacted-group provenance, so any agent can cite and resolve evidence. Pre-backend placement genuinely helps here because PulseFold sees every event before sampling/compaction.
- "Proactive signals" from template novelty at ingest time (new error template after a deploy) — cheap, deterministic, useful, but also already in Edge Delta.

## 2. Answers to the brief's questions

| Question | Answer |
|---|---|
| Why won't operators trust compaction? | It changes timestamps (logdedup rewrites them), breaks alert queries and dashboards that count lines, removes the exact line someone greps for, and template errors are invisible. Trust requires per-template opt-in, reversibility (archive) and a test that their alerts still fire. |
| Why would a vendor already solve this? | Backend vendors profit from ingest volume (disincentive), but pipeline vendors (Cribl, Sawmills, Edge Delta, Datadog OP) and Grafana Adaptive Logs already do. Grafana chose query-usage-based dropping — evidence that vendors view this as solved enough. |
| What is much harder than it appears? | (1) Canonicalising OTel records for exact identity; (2) cross-instance consistency of templates & dedup under load balancing; (3) fail-open in a shared process; (4) honest savings estimates (wire vs billed bytes, after-zstd); (5) evaluation ground truth for RCA. |
| What could make it unsafe? | Telemetry reward-hacking of investigator (AIOpsDoom); evidence laundering via compaction; secrets baked into templates; unbounded state → OOM → black hole. |
| What would cause false savings estimates? | Counting uncompressed bytes when the backend bills compressed or per event; ignoring OTAP/zstd baseline; ignoring that compacted aggregate events carry extra attributes; single-instance analysis of a multi-instance fleet (overstates dedup); analysing a quiet week; ignoring logs needed by existing alerts/dashboards (Grafana's query-usage signal). |
| What makes an AI investigator worse than existing tools? | No history/topology/runbook context; less integration depth; model cost; hallucination; incumbents already embedded in Slack/on-call. |
| Which components are commodity? | OTLP receive/export, dedup, Drain, filtering, sampling, redaction, sketches, Slack/webhook notify, model-provider abstraction (LiteLLM-style), MCP server scaffolding, alert grouping (Alertmanager/Keep). |

## 3. Unsafe / naive / overcomplex items in the current docs

1. **Immutable evidence store vs secrets/PII erasure** (02 §6) — legal conflict; needs tombstoning design. (Unsafe)
2. **Redaction/compaction ordering left open** (03 §8.8) — must be decided: redact before fingerprint. (Unsafe)
3. **8-stage pipeline + incident engine + investigator + router in one product for V0–V1** — overcomplex for a personal OSS project. (Overcomplex)
4. **Built-in investigator + model provider abstraction** duplicates HolmesGPT and LiteLLM-class libraries. (Commodity)
5. **C3 payload extraction** savings largely duplicated by OTAP/zstd on the wire and breaks standalone events at destination. (Technically naive if justified on bytes alone)
6. **Fail-open claim** (PR-006) is hard inside a Collector process; the doc treats it as configuration. (Naive)
7. **Incident state machine with 8 states** before any detector exists. (Premature)
8. **Similar-incident matching** without a history store contradicts "not a primary DB". (Inconsistent)

## 4. Recommended scope narrowing (RECOMMENDATION, for Alex)

V0 = **shadow-mode analyser only**, built as Collector component(s) in a separate pipeline/process, using upstream `logdedup`/`drain` logic as baselines, producing a reproducible report with honest byte accounting (raw / zstd / OTAP / billed-model) and safety metrics. Defer built-in investigator indefinitely in favour of an evidence/event contract consumed by HolmesGPT/external agents.

## Sources
See competitive-landscape.md, log-compaction-research.md, otel-extension-options.md, threat-model-challenge.md.
