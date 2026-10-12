# PF-0006: Independent architecture review

**Reviewer:** Codex
**Date:** 2026-10-12
**Status:** Recommendation only. Accepted ADRs and product requirements are not modified.

## Executive assessment

The locked principles are coherent, but the proposed gateway diagram overstates what an in-process Collector pipeline can safely isolate. The strongest Phase 0 outcome is not a new Collector feature set. It is a measurable, redacted, evidence-aware shadow analysis methodology that can run independently of production forwarding and can feed existing incident and agent tools.

The recommended architecture is a hybrid in the narrow sense: use Collector components and distributions for telemetry transport and integration, while a separate analysis service owns bounded aggregation and reports. Defer incident state, routing UI, archive design, active compaction and built-in investigation until the shadow report proves value and safety.

## Ranked findings

| ID | Finding | Classification | Severity | Decision |
|---|---|---|---|---|
| AR-01 | Same-process Collector fan-out is synchronous and returns combined downstream errors. It is not availability isolation. See [OTV-01](codex-otel-verification.md#findings). | Verified | Critical | **Reject** as the V0 fail-open design. |
| AR-02 | `log_dedup` is aggregation with changed timestamp semantics, bounded only by its interval and memory, not a lossless evidence primitive. Its identity relies on 64-bit hashes. See [OTV-05](codex-otel-verification.md#findings) and [OTV-06](codex-otel-verification.md#findings). | Verified | Critical for active compaction | **Use as benchmark baseline only.** |
| AR-03 | The proposed immutable evidence item conflicts with retention, deletion and redaction needs if immutability includes raw content. | Architecture inference, supported by the project's own security requirements | High | **Accept decision F only after redefining it** as immutable identity and provenance with lifecycle-managed content. |
| AR-04 | A template learner, dedup map and cardinality sketch are stateful ingestion components. Unbounded state, high-cardinality input and large values are a direct availability attack. | Verified for upstream component defaults, see [OTV-07](codex-otel-verification.md#findings) and [OTV-08](codex-otel-verification.md#findings) | High | **Accept only with hard entry, byte and time caps plus pass-through or analysis-drop overflow.** |
| AR-05 | Redaction after any fingerprint, template, snapshot, report or model boundary allows secret persistence and weakens erasure. Upstream Drain can emit raw wildcard values. | Verified, see [OTV-10](codex-otel-verification.md#findings) | High | **Accept decision E.** Make ordering a security invariant. |
| AR-06 | A template is an inference, not a stable universal ID. Independent Drain instances can start with different state, while shared snapshot persistence has no merge or write coordination. | Verified/inference boundary stated in [OTV-09](codex-otel-verification.md#findings) | High | **Investigate** before multi-instance C2 aggregation. |
| AR-07 | Proactive detection is valuable only when deterministic signals preserve source observations, rate limits and uncertainty. HolmesGPT Operator already performs LLM-driven scheduled and rollout-triggered proactive health checks. | Verified, see [CG-01](codex-competitive-gaps.md#findings) and [CG-02](codex-competitive-gaps.md#findings) | Medium | **Accept** proactive deterministic detection. **Defer** proactive LLM investigation. |
| AR-08 | An incident engine, workflow router and built-in investigator duplicate mature adjacent products before PulseFold has proven the telemetry value proposition. | Market/architecture inference, grounded in [CG-03](codex-competitive-gaps.md#findings) and [CG-04](codex-competitive-gaps.md#findings) | Medium | **Reject from V0.** |
| AR-09 | Telemetry prompt injection includes plausible, attacker-controlled root-cause narratives, not merely obvious instruction strings. Read-only tools reduce mutation risk but do not prevent harmful recommendations or provider data exfiltration. | Verified threat class. [USENIX AIOpsDoom paper](https://www.usenix.org/conference/usenixsecurity26/presentation/pasquini) | High for future agent work | **Accept** a taint and evidence-rendering design before any model integration. |

## Challenge to the six proposed decisions

| Decision | Assessment | Recommendation |
|---|---|---|
| A. V0 starts with shadow analysis only. | Necessary and consistent with ADR 0004. Shadow is not automatically safe: the analysis capture path can still affect forwarding and reports can be manipulated by incomplete or adversarial input. | **Accept**, conditioned on explicit analysis coverage and no claim of fail-open until tested. |
| B. Reuse existing OpenTelemetry primitives. | Correct for transport, configuration and comparative baselines. Incorrect if it assumes alpha components provide PulseFold's evidence, collision or isolation guarantees. | **Accept with boundary.** Reuse first, benchmark first, but own only the missing safety contract. |
| C. Proactive detection remains a core objective. | Deterministic novelty and rate signals create a credible link from noise reduction to operator value. Proactive LLM health checks are already an adjacent product capability. | **Accept only deterministic, evidence-linked detector requirements beyond V0.** |
| D. External agents precede built-in investigation. | Avoids duplicated model orchestration and respects model neutrality. External agents still need precise, signed, redacted and scope-bound evidence references. | **Accept.** V0 need only publish the contract, not operate agents. |
| E. Redaction precedes fingerprinting. | Necessary because fingerprints, templates and state are derived data that can preserve a secret. It changes group cardinality, so reports must state the redaction policy version. | **Accept.** Propose a new ADR to lock it. |
| F. Evidence identity/provenance is immutable, while content supports retention and erasure. | Viable only with content separated from its immutable manifest. A content hash can itself be personal data or facilitate correlation if it uses unsalted values. | **Accept with design constraints.** Use a non-reversible, policy-scoped content commitment, immutable manifest and append-only tombstone/erasure event. |

## Proposed boundaries and state ownership

```text
Production forwarding process
  Receive OTLP -> redaction policy if forwarding policy requires it -> batch/retry/export
  Owns: transport queues and delivery telemetry only
  Must not wait for: analysis, evidence persistence, incident state, webhooks or LLMs

Independent analysis process
  Receive an explicitly replicated, non-authoritative stream or offline replay
  -> redact -> canonicalise -> bounded C1/C2 analysis -> versioned report
  Owns: bounded per-window analysis state and coverage accounting
  Must not: alter production forwarding or become a required exporter

Future control service
  Consume signed report/events -> evidence manifests -> deterministic signals
  -> optional generic webhook/MCP to external systems
  Owns: incident state only after a concrete detector and consumer exist
```

The replication boundary is a product decision, not an implementation detail. A shared Collector receiver is insufficient for isolation. V0 should support offline replay and a deliberately best-effort replicated stream. If the replicated stream is incomplete or drops, the report must state it rather than estimate as though it were complete.

## Evidence lifecycle recommendation

An evidence manifest should be immutable and small: `evidence_id`, observed time range, source partition, redaction policy version, analyser version, transformation-policy version, content commitment, sensitivity class and lineage references. Raw or redacted content belongs in a separately encrypted store with an explicit retention policy. Erasure creates a tombstone with the erasure reason and time, removes the content or decryption key and leaves the manifest resolvable as `content unavailable by policy`.

This avoids silently rewriting history while respecting retention and erasure. It does not solve legal compliance by itself and requires a threat-model review before a content store exists.

## Security requirements missing from the current architecture

1. Taint fields by origin. User-controlled request paths, headers, bodies and exception strings must remain data and be visibly marked when rendered to an agent.
2. Never allow telemetry text to select a callback URL, tool, package, command or evidence query scope.
3. Require tool credentials scoped to an incident and a time limit. Separate forwarding credentials from control-plane credentials.
4. Hard-cap every stateful map, cluster tree, sample reservoir, queue and report cardinality. Overflow behaviour must be recorded and must not silently transform or hide input.
5. Use randomised representative selection rather than first or last only, otherwise an attacker can poison the sample retained for a high-volume group.
6. Bind reports to input coverage, configuration and analyser version. A shadow report without these is not auditable evidence for enabling compaction.
7. Treat outbound model providers and notification endpoints as data-exfiltration boundaries, even for read-only investigations.

## Proposed ADRs, without changing accepted ADRs

| Proposed ADR | Decision to record | Why it is needed |
|---|---|---|
| ADR 0007: Isolated shadow analysis and coverage | Production forwarding and non-essential analysis run in separate failure domains. Shadow reports carry coverage and drop status. | Clarifies ADR 0004 and prevents an unsupported fail-open claim. |
| ADR 0008: Redaction before derivation | Apply the approved redaction policy before fingerprints, templates, content commitments, state snapshots, reports or model exposure. | Locks decision E and closes a secret persistence gap. |
| ADR 0009: Evidence manifest and content lifecycle | Immutable provenance manifest, lifecycle-managed encrypted content, append-only tombstones and policy-scoped commitments. | Makes decision F compatible with retention and erasure. |
| ADR 0010: External investigation contract first | V0 publishes a model-neutral, read-only external-agent event/evidence contract; built-in investigation is deferred. | Narrows V0 and records decision D without weakening ADR 0003. |

## Initial benchmark gates to approve before Phase 1

These are proposed gates, not measured results:

- No unsafe C1 merge in a collision-injection test that uses deliberately equal bucket keys. Full equality verification must be demonstrable if active C1 is proposed.
- No silently unreported analysis loss. Every overflow, refusal, capture drop and unsupported field must appear in the report.
- The forwarding path retains its declared delivery behaviour when analysis is blocked, crashed or memory-limited.
- C2 is report-only until its unsafe merge rate, template agreement and memory envelope are measured on synthetic truth-labelled streams and a licensed external corpus.
- Reports separate raw bytes, wire bytes and a contract-specific billing estimate as described in [CG-05](codex-competitive-gaps.md#findings) and [CG-06](codex-competitive-gaps.md#findings).

## Architecture decision

**Recommend a small hybrid: Collector-based forwarding and optional component packaging, plus a separate analysis service.** Do not create a standalone OTLP proxy that reimplements Collector exporters. Do not package an incident platform, storage system or investigator into V0. This leaves a credible route to the longer-term architecture only after measured evidence supports it.
