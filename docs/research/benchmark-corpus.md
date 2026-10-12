# PF-0004: Public Benchmark Corpus

**Status:** Draft (Grok, 2026-10-12).

## 1. Public datasets

| Dataset | Content | Licence (as found) | Suitability | Limitations |
|---|---|---|---|---|
| **Loghub** (logpai) — 16+ system logs (HDFS, BGL, Thunderbird, OpenStack, Apache, Linux, Android, HPC, Zookeeper, Spark, etc.) [1] | Raw logs; 2k-line labelled samples | GitHub SPDX NOASSERTION; README: "freely available for research or academic work", cite paper [1] | Template mining, dedup ratios, novelty | **Licence not OSI**: do *not* vendor into an Apache-2.0 repo; download at benchmark time. Mostly old, unstructured, non-OTel, non-microservice. |
| **Loghub-2.0** [2] | Full-size labelled templates for 14 datasets (~50M lines) | NOASSERTION (verify) | Grouping accuracy at scale — the key C2 safety benchmark | Same licence caveat; labels have known errors (paper discusses corrections). |
| **RCAEval** [3] | 735 failure cases, Online Boutique / Sock Shop / Train Ticket, 11 fault types, metrics+logs+traces, 15 baselines | Code MIT; data licence: verify (hosted on Zenodo) | Incident/RCA ground truth, signal detection | Lab microservices; injected faults; short windows. |
| **OpenRCA** (Microsoft, ICLR'25) [4] | 335 failures, 68+ GB telemetry, natural-language RCA tasks with scoring | Verify repo licence | Agent investigation eval (Phase 4) | Not OTel-native; scoring targets component/time/reason. |
| **ITBench** (IBM) | SRE/FinOps/CISO agent scenarios on K8s | Verify (Apache-2.0 expected; Low confidence) | Agent eval | Requires running environments. |
| LogPub / LogPM etc. | Further parsing benchmarks | Unverified | Optional | — |

## 2. Synthetic generation (recommended primary source for repo fixtures)

Rationale: AGENTS.md rule 9 + licence issues mean committed fixtures should be **generated**, deterministic and Apache-2.0.

**Environment:** OpenTelemetry Demo (Apache-2.0, has built-in feature-flag fault injection via flagd) [5] and/or Google Online Boutique (Apache-2.0) [6]. Emits native OTLP logs/metrics/traces → PulseFold-realistic.

**Generator spec (design, not code):**
- Seeded RNG; output OTLP/JSON or protobuf files plus a **ground-truth sidecar** (per record: true template id, true duplicate group, is_novel, scenario_id, injected_fault, timestamps).
- Knobs: events/s, template count, Zipf exponent over templates, parameter cardinality per variable, payload size distribution (stack traces 2–50KB), severity mix, number of services/versions, clock skew, attribute-order shuffle (canonicalisation test), multi-instance split (cross-collector test).
- Adversarial records: near-duplicates differing only in severity, or only in `deployment.environment`, or in one digit of an error code (unsafe-merge tests); prompt-injection strings in bodies/attributes (PF-4006); secrets (fake tokens matching gitleaks/trufflehog patterns, never real); 1MB bodies; invalid UTF-8.
- Scenarios mapped to 09_EVALUATION_STRATEGY §4 list: bad deploy (new template after `service.version` change), DB pool exhaustion, dependency outage, DNS failure, expired credential, rate limit 429 storm, queue backlog, memory leak, CPU saturation, schema mismatch, permission regression, noisy neighbour, latency degradation, retry storm. Each scenario defines expected signals, expected incident grouping and onset time.

## 3. Recommendations

1. Commit only synthetic fixtures; fetch Loghub/Loghub-2.0/RCAEval at benchmark time with checksums and a licence notice. (High confidence)
2. Use Loghub-2.0 labelled sets as the C2 grouping/unsafe-merge benchmark. (High)
3. Use OTel Demo fault flags to produce realistic OTLP shadow-mode runs. (Medium — verify flag list in current demo release)
4. Use RCAEval/OpenRCA only for Phase 3/4 evaluation, not V0. (Medium)

## Sources
1. https://github.com/logpai/loghub
2. https://github.com/logpai/loghub-2.0 ; https://arxiv.org/abs/2308.10828
3. https://github.com/phamquiluan/RCAEval ; https://arxiv.org/abs/2412.17015
4. https://github.com/microsoft/OpenRCA
5. https://github.com/open-telemetry/opentelemetry-demo ; https://opentelemetry.io/docs/demo/feature-flags/
6. https://github.com/GoogleCloudPlatform/microservices-demo
