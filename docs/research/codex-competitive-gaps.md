# PF-0006: Competitive integration and economics review

**Reviewer:** Codex
**Date:** 2026-10-12
**Scope:** HolmesGPT, Keep, Alertmanager and telemetry-reduction economics. Product capability statements are based on the linked primary project documentation, not independent performance tests.

## Findings

| ID | Finding and primary source | Classification | Risk | Recommendation |
|---|---|---|---|---|
| CG-01 | HolmesGPT Operator is an alpha Kubernetes operator. It runs one-time, scheduled and deployment-triggered health checks as CRDs, sends failures to destinations and asks a Holmes API service to perform an LLM-backed check. It can query connected data sources and its documentation proposes GitHub integration that can open pull requests. | **Verified.** [Operator overview](https://github.com/HolmesGPT/holmesgpt/blob/d64377138c604a1b60af5766f4b01e0bde7ef879/docs/operator/index.md), [deployment verification](https://github.com/HolmesGPT/holmesgpt/blob/d64377138c604a1b60af5766f4b01e0bde7ef879/docs/operator/deployment-verification.md). | Proactive investigation is already supplied for Kubernetes users. It is alpha, needs Kubernetes plus a Holmes API and an LLM call per check, so it is neither a replacement for low-cost deterministic ingest detection nor an availability-free dependency. | **Accept integration, reject duplication in V0.** PulseFold should emit an evidence-bearing event or provide a read-only evidence tool that Holmes can call. Do not add a built-in investigator in V0. |
| CG-02 | Operator Mode's proactive capability is schedule and rollout-trigger based, not a high-rate streaming log compactor or a deterministic anomaly engine. The published model uses a controller and stateless API servers, and stores check history in Kubernetes resource status. | **Verified.** [Operator architecture](https://github.com/HolmesGPT/holmesgpt/blob/d64377138c604a1b60af5766f4b01e0bde7ef879/docs/operator/index.md). | Treating Holmes as the signal engine would impose LLM cost and diagnosis latency on every check and exclude non-Kubernetes V0 users. | **Accept decision C with a narrower definition:** proactive deterministic signals remain a core objective, while proactive LLM health checks are an optional downstream consumer. |
| CG-03 | Keep provides alert deduplication, enrichment, filtering, correlation, workflows and a broad set of integrations. Its workflows can trigger on alerts, incidents or schedules and include arbitrary provider actions. | **Verified.** [Keep README](https://github.com/keephq/keep/blob/465e99152e7586fcbe4c2732213d31c5c28def4b/README.md). | Building notification routing, a collaboration UI and general incident automation before demonstrating unique telemetry value would duplicate a large, fast-moving surface. Some Keep actions are mutating, which conflicts with the read-only default unless separately governed. | **Accept integration, reject Keep as a required dependency.** Emit a signed generic webhook event compatible with Keep workflows. PulseFold should retain ownership only of telemetry provenance and deterministic findings. |
| CG-04 | Alertmanager groups alert notifications by configured labels and routes them to receivers. Its `group_wait` deliberately trades notification completeness against time to first notification, and Alertmanager documents grouping as notification behaviour. | **Verified.** [Alertmanager configuration](https://github.com/prometheus/alertmanager/blob/55b27cfb6525641ad435e9673e655f38b329e4f4/docs/configuration.md). | Alertmanager is not a telemetry store, template miner or evidence graph. Replacing PulseFold's incident model with an Alertmanager group would collapse evidence lineage and state semantics. | **Accept as a notification destination, not as the incident engine.** V0 should support Alertmanager-compatible webhook payloads only if a concrete integration need emerges. |
| CG-05 | OTel Arrow reported 30–70% less network bandwidth than comparably configured OTLP/gRPC with large batches and zstd in its 2024 production case study. The report also exposes separate uncompressed and wire-byte metrics. | **Verified, vendor/project-reported result rather than a PulseFold benchmark.** [OTel Arrow production report](https://opentelemetry.io/blog/2024/otel-arrow-production/). | A claim of savings based only on serialised OTLP size could be mistaken for network or ingestion savings, and can exaggerate C3 benefit. | **Accept as the transport baseline.** Every PulseFold report must display raw canonical bytes, encoded bytes, compressed wire bytes and the measurement method. |
| CG-06 | Datadog exposes estimated `logs.ingested_bytes`; Splunk Log Observer documents ingest and indexing usage separately in GB. These facts demonstrate that vendor accounting can differ from network transfer and retention, but neither source establishes a universal billing rule. | **Verified.** [Datadog usage metrics](https://docs.datadoghq.com/account_management/billing/usage_metrics/), [Splunk usage documentation](https://help.splunk.com/en/splunk-observability-cloud/administer/monitor-subscription-usage-and-billing/logs-subscription-usage). | Any generic dollar-saving estimate is speculative unless it uses the customer's contract, destination accounting definition, retention tier, egress price and added PulseFold cost. | **Reject** generic customer-savings claims in V0. Present a parameterised estimate with an `unknown` state when the billing model is unavailable. |

## HolmesGPT Operator Mode: what PulseFold adds

HolmesGPT already provides proactive LLM work in four forms: one-time checks, cron-like scheduled checks, automatic checks after matching Deployment rollouts and deployment-verification checks that can gate CI/CD. Its documentation explicitly describes Slack destinations and optional GitHub PR creation. The latter makes operator configuration a privileged boundary, even though PulseFold itself remains read-only.

PulseFold should not compete with that orchestration. Its credible addition is upstream and deterministic:

- estimate safe reduction opportunity without changing the backend payload
- detect a newly observed, redacted template or deterministic rate change at ingest time
- attach immutable provenance to the observation, including policy and analyser version
- expose bounded, redacted evidence references to a Holmes investigation
- report coverage and uncertainty, including whether the analysis stream dropped records

This gives Holmes a more trustworthy trigger and input. It does not turn the LLM's conclusion into evidence. The interface must be optional, vendor-neutral and disabled by default.

## Economics: four quantities that must not be conflated

| Quantity | What it measures | Can C1/C2 reduce it in shadow mode? | Can it prove customer saving? |
|---|---|---|---|
| Raw canonical bytes | Deterministically measured size of the incoming OTLP representation selected by PulseFold. | No. Shadow mode transmits unchanged data. It can estimate a counterfactual. | No. It is an engineering volume, not a bill. |
| Encoded and compressed wire bytes | Bytes on a particular link after OTLP protobuf/JSON choice, batching, TLS framing and compression. | No for unchanged forwarding. Potentially, but only after measuring the exact path. | Only for a customer actually paying egress on that path. OTel Arrow can materially change this baseline. |
| Billable ingest/index/retention units | The vendor's contract-specific accounting quantity, often GB or events, sometimes with separate index and retention dimensions. | Usually only after active transformation before the vendor's accounting boundary. | Only after mapping the vendor's documented meter and the customer's price schedule. |
| Net customer saving | Avoided contractual charge minus added collector/analyser compute, storage, archive, egress, operations and any feature-tier side effects. | Not in shadow mode. | Yes, but only with customer inputs and a time-window comparison. |

### Required V0 report fields

1. Input records and canonical raw bytes.
2. Counterfactual output records and bytes for each proposed transformation.
3. Wire-byte result for the measured baseline, labelled with protocol, compression, batch size and link.
4. A vendor-meter mapping supplied by the operator, or `billing model unknown`.
5. PulseFold resource cost and analysis coverage, including input dropped before analysis.
6. A confidence interval or explicit exclusion for high-cardinality, secret-redacted, novel and error records.

The report may say `estimated reduction opportunity`. It must not say `customer saving` without the fourth and fifth inputs.

## Integration decision

**Accept decision D, external agents before a built-in investigator.** Extend it to an explicit V0 rule: HolmesGPT, Keep and Alertmanager are optional consumers or destinations, never required control-plane dependencies. The PulseFold core differentiates through reproducible telemetry analysis and provenance, not by recreating investigation, workflow or paging products.
