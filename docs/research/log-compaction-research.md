# PF-0002: Log Compaction Algorithm Review

**Status:** Draft (Grok, 2026-10-12). Recommends **benchmark candidates**, not a winner (per brief).

## 0. Summary

- FACT: exact dedup (C1) and Drain template annotation (C2) exist as alpha OTel Collector contrib processors [1][2].
- FACT: on the large Loghub-2.0 benchmark, Drain had the best *grouping* accuracy among classic parsers, but its F1 grouping accuracy drops from ~0.75 (Loghub-2k) to ~0.55 (Loghub-2.0) [3]. LLM parsers (LILAC, LogBatcher) improve *template* accuracy on some datasets [4][5].
- INFERENCE: grouping errors directly translate into **unsafe merges** under compaction. A parser accurate enough for *pattern insight* is not necessarily accurate enough for *destructive aggregation*. PulseFold's C2 safety story must assume ~10–45% grouping error on heterogeneous logs unless benchmarks prove otherwise.
- RECOMMENDATION: compaction decisions should key on **(template_id + severity + resource identity + policy-declared stable fields)** and *never* on the template alone; templates are a grouping hint, exactness is the safety gate.

## 1. Exact deduplication (C1)

| Approach | Facts | Notes for PulseFold |
|---|---|---|
| OTel `logdedup` | Identity = body + resource attrs + severity + log attrs; interval default 10s; `exclude_fields` removes fields from identity *and from output*; emitted Timestamp/ObservedTimestamp become emit time; `metadata_keys` partitioning; `metadata_cardinality_limit` default unlimited (warns) [1] | (a) Rewriting timestamps is a **semantic change** many backends/alerts will notice. (b) Not hash-collision-prone in the PulseFold sense (in-memory map keys) — verify implementation. (c) Per-instance state: duplicates split across load-balanced collectors are not merged. |
| Datadog OP dedupe | ≤3 fields, LRU 5k events, per-worker only [6] | Documented same per-instance limitation. |
| Cribl Suppress | N per key per period, `suppressCount` on next allowed event [7] | Count rides on the *next* event — lossy if no next event. |
| Vector `dedupe` | LRU cache of field values (MPL-2.0) | Drop-based, no counts by default (verify). |

Safety invariant 1 ("no unique event classified duplicate solely by lossy hash") — RECOMMENDATION: use a 128-bit+ hash for bucketing *and* full-byte comparison against the stored representative (or store canonical bytes); count collisions as a metric. Cost: memory per bucket. Confidence High (standard practice).

Canonicalisation is the hard part: attribute map ordering, numeric vs string types, body as AnyValue map vs string. OTel pdata maps are ordered as received; two semantically identical records may serialise differently. Benchmark must report "canonicalisation-induced misses".

## 2. Template mining (C2)

### Candidates to benchmark

| Algorithm | Type | Implementation | Licence | Streaming? | Notes |
|---|---|---|---|---|---|
| Drain (He et al. ICWS'17) | fixed-depth parse tree, similarity threshold | OTel `drainprocessor` (Go, via `jaeyo/go-drain3`); Drain3 (Py, MIT) | Apache-2.0 / MIT | Yes | De facto standard (Loki, Edge Delta, OTel all use it) [2][8][9] |
| Brain (Yu et al. 2023) | bidirectional tree, longest common pattern | logpai/logparser (research code) | logparser licence NOASSERTION per GitHub API — verify | Partly | Reported strong on Loghub-2.0 [3] |
| Spell | LCS-based streaming | logparser | as above | Yes | Older; slower on long lines |
| LogMine | hierarchical clustering | logparser | as above | Batch | Not suitable for hot path |
| IPLoM, AEL, LenMa, etc. | heuristics | logparser | as above | mixed | Historical baselines only |
| LILAC (ICSE'24) | LLM + adaptive parsing cache | research code | — | cache-based | **Violates ADR 0002 on hot path**; usable offline to propose seed templates |
| LogBatcher (2024) | LLM, demonstration-free, batching | research code | — | batch | Same as above [5] |
| Source-derived templates | capture logging-call template at SDK (`log.record.template`) | OTel Java instrumentation PRs (#15423, #19154) referenced from semconv #1283 [10] | Apache-2.0 | n/a | **Zero inference error** where available. Strongest option; PulseFold should prefer it when present. |

### Key facts / conflicts

- Loghub-2.0 paper (ISSTA'24): many parsers' accuracy on the 2k-line samples overstated real performance; on full datasets (~50M lines total) grouping accuracy dropped notably [3]. Conflict: LILAC/LogBatcher papers report near-perfect GA on some sets (HDFS, OpenStack) [4][5] — those are the most regular datasets; heterogeneous sets are much worse.
- **Determinism ≠ order independence.** The `drainprocessor` README claims instances "converge on identical templates because the Drain algorithm is deterministic" with an early-training caveat [2]. INFERENCE (Medium-High): Drain's output depends on arrival order and on `max_clusters` LRU eviction, so horizontally scaled collectors can produce *different template strings for the same line* indefinitely, not only during warmup. PulseFold must test cross-instance template agreement explicitly. Mitigation offered upstream: `seed_templates`, `warmup_min_clusters`, snapshot persistence.
- Drain merges on token count + prefix tokens: lines with different severity words or a leading variable token are classic failure modes; ensure severity is part of identity (Safety invariant 3).

### Template drift

Facts: Drain3 and drainprocessor support persisted state; no standard drift metric exists in OTel. RECOMMENDATION: track (a) new-template rate, (b) template-split/merge events (cluster template changed after N matches), (c) version-correlated template changes (`service.version`). A template whose string changes must keep a stable `template_id` lineage or downstream group counts become non-comparable. Confidence Medium.

## 3. Repeated payload extraction / content addressing (C3)

- Techniques: content-defined chunking (Rabin/FastCDC), whole-field hashing for stack traces/bodies, dictionary compression (zstd trained dictionaries), columnar dictionary encoding (OTel Arrow [11]), CLP's "log type + variable dictionary + encoded vars" [12].
- FACT: OTel Arrow already achieves dictionary-style dedup on the wire (30–70% vs OTLP+zstd, OTel-reported) [11]; zstd itself removes much repetition within a batch.
- INFERENCE (High): for **network egress**, C3 gains largely duplicate what OTAP+zstd already give losslessly. C3's real value is only if the **destination bills on uncompressed ingested bytes** (Datadog/Splunk/most SaaS) — then replacing a 20KB stack trace with a reference reduces billed bytes. But that **breaks standalone events in the destination** (the doc already flags this, §6 of 03). Operators may reject events whose stack trace lives elsewhere.
- RECOMMENDATION: in V0 shadow report, estimate C3 savings **separately for "billed ingest bytes" vs "wire bytes after zstd/OTAP"**, or the savings estimate will be inflated.

## 4. Cardinality estimation

- HyperLogLog++ (Heule et al. 2013) — standard; `cardinalityguardian` uses `axiomhq/hyperloglog` (~2KB/tracker, 1–2% error) [13].
- Alternatives: Theta sketches (Apache DataSketches, supports set ops), KMV, Count-Min (frequency, not cardinality), Space-Saving / Misra-Gries for heavy hitters (top templates/dimensions — what the shadow report's "top reducible templates" actually needs).
- RECOMMENDATION: benchmark HLL++ (cardinality) + Space-Saving (top-k) together; memory bounded per template × parameter, with a hard cap and an "untracked" counter.

## 5. Representative sampling

- Reservoir sampling (Vitter Algorithm R / L) for uniform samples per group; weighted reservoir (Efraimidis-Spirakis A-Res) to bias toward errors/rare dimension values.
- Stratified sampling across `service.version`, `k8s.pod.name`, etc. (doc §5 "materially different dimensions").
- Honeycomb dynsampler-go (already embedded in `adaptive_tail_sampling`) gives per-key dynamic rates with known sample rate encoded for re-weighting [14]. OTel consistent probability sampling (tracestate `ot=th`) is the standard way to make sampled counts re-weightable [14].
- RECOMMENDATION: every kept sample carries its sampling weight so counts reconstructed downstream are unbiased; adopt OTel's threshold encoding rather than a custom field where signals are traces.

## 6. Rarity / anomaly-preserving retention

- Novelty: new template (needs stable template IDs — see §2 drift problem); rare dimension combo (Count-Min below threshold).
- Statistical: EWMA / Holt-Winters baselines, robust z (MAD), Isolation Forest (contrib processor exists, can drop data [15]).
- Grafana Adaptive Logs uses **query usage** rather than rarity as value signal [16] — conflicting design philosophy worth noting: rare ≠ valuable; frequent ≠ worthless (e.g. per-request audit logs).
- RECOMMENDATION: V0 should report *what would be compacted* stratified by severity and novelty; C4 adaptive retention stays out of scope (consistent with 02_ARCHITECTURE).

## 7. Windowing

Facts: logdedup uses fixed interval; Cribl Suppress uses per-key period; Datadog uses LRU without time. Trade-off: longer windows → more reduction, more memory, **higher detection latency for the downstream backend** (a burst becomes visible only at flush). RECOMMENDATION: benchmark 1s/10s/60s fixed and per-key-first-pass-through ("emit first occurrence immediately, aggregate the rest") — the latter preserves time-to-first-signal for incidents. Confidence Medium.

## 8. Benchmark plan (input to PF-1003..1005)

Metrics per candidate: event/byte reduction (raw, after zstd, after OTAP), grouping accuracy & F1-GA on Loghub-2.0 labelled sets, **unsafe merge rate** (different ground-truth template or severity merged), novel-event preservation, cross-instance template agreement (2–4 shuffled instances), memory/CPU at 10k/100k events/s, p99 added latency.

Candidates: C1 = `logdedup` vs PulseFold-style exact-with-first-passthrough; C2 = drainprocessor (default + tuned), Brain, Drain3 (reference), source-provided templates (`log.record.template`) where synthetic generator emits them; C3 = zstd-dict, OTAP, whole-field content hash.

## Uncertainty register

| Item | Confidence | Why |
|---|---|---|
| Drain cross-instance divergence persists past warmup | Medium | Reasoned from algorithm; not yet measured |
| C3 gains mostly duplicated by OTAP/zstd for wire bytes | Medium-High | OTel blog numbers; depends on vendor billing |
| logparser licences | Low | GitHub reports NOASSERTION; read LICENSE files before use |

## Sources

1. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor
2. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/drainprocessor
3. Jiang et al., "A Large-Scale Evaluation for Log Parsing Techniques: How Far Are We?" ISSTA 2024, https://arxiv.org/abs/2308.10828 ; https://github.com/logpai/loghub-2.0
4. Jiang et al., LILAC, ICSE 2024, https://arxiv.org/abs/2310.01796
5. Xiao et al., LogBatcher, https://arxiv.org/abs/2406.06156
6. https://docs.datadoghq.com/observability_pipelines/processors/dedupe/
7. https://docs.cribl.io/stream/suppress-function/
8. https://grafana.com/docs/loki/latest/get-started/components/
9. https://docs.edgedelta.com/log-to-pattern-node/
10. https://github.com/open-telemetry/semantic-conventions/issues/1283 ; https://github.com/open-telemetry/semantic-conventions/issues/2064
11. https://opentelemetry.io/blog/2024/otel-arrow-production/
12. Rodrigues et al., CLP, OSDI 2021, https://www.usenix.org/conference/osdi21/presentation/rodrigues
13. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/cardinalityguardianprocessor ; Heule et al., HyperLogLog in Practice, EDBT 2013
14. https://github.com/honeycombio/dynsampler-go ; https://opentelemetry.io/docs/specs/otel/trace/tracestate-probability-sampling/
15. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/isolationforestprocessor
16. https://grafana.com/docs/grafana-cloud/observe-and-act/adaptive-telemetry/adaptive-logs/introduction/
- He et al., Drain, ICWS 2017; Drain3 https://github.com/logpai/Drain3
