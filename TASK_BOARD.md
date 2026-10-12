# PulseFold Task Board

## Phase 0: Research and architecture

### PF-0001: Competitive landscape
**Status:** Complete (2026-10-12). Research accepted during PF-0007.
**Owner:** Grok  
Map direct and adjacent projects across observability cost reduction, log compaction, AI SRE, incident intelligence and agent integrations.

**Output:** `docs/research/competitive-landscape.md`

**Acceptance criteria:**
- distinguishes direct competitors from components and adjacent tooling
- identifies capabilities we should reuse rather than rebuild
- identifies at least five credible differentiation risks
- includes licence and activity information where available

### PF-0002: Log compaction algorithm review
**Status:** Complete (2026-10-12). Research accepted during PF-0007.
**Owner:** Grok  
Research exact deduplication, template mining, dictionary/content addressed payload storage, cardinality estimation, adaptive sampling and anomaly preserving retention.

**Output:** `docs/research/log-compaction-research.md`

### PF-0003: OpenTelemetry extension strategy
**Status:** Complete (2026-10-12). Reconciled by ADR 0007 and PF-0006 verification.
**Owner:** Grok, reviewed by Codex  
Determine whether PulseFold should ship as:
- a custom OpenTelemetry Collector distribution
- one or more Collector processors/connectors
- a standalone OTLP proxy
- a hybrid

**Output:** `docs/research/otel-extension-options.md`

### PF-0004: Public benchmark corpus
**Status:** Complete (2026-10-12). Corpus use remains a Phase 1 execution gate.
**Owner:** Grok  
Find public datasets and define synthetic generators for repetitive logs and known failure scenarios.

**Output:** `docs/research/benchmark-corpus.md`

### PF-0005: Threat model challenge
**Status:** Complete (2026-10-12). Findings reconciled by ADRs 0008 and 0009.
**Owner:** Grok  
Challenge `docs/08_SECURITY_THREAT_MODEL.md`, focusing on telemetry injection, prompt injection, secrets, PII, SSRF, tool abuse, credential scope and multi tenant risks.

### PF-0006: Architecture review
**Status:** Complete (2026-10-12). Output: `docs/research/codex-architecture-review.md`.
**Owner:** Codex  
Review all accepted Phase 0 research and challenge `docs/02_ARCHITECTURE.md`.

**Output:** architecture review notes plus ADR proposals where required.

### PF-0007: V0 scope freeze
**Status:** Complete (2026-10-12). V0 is frozen by ADRs 0007 through 0010 and the reconciled documentation.
**Owner:** Alex/Chad  
Freeze the first executable scope after research.

## Phase 1: Telemetry laboratory

### PF-1001: Offline OTLP replay and replica-input boundary
**Status:** Complete (2026-10-12).
Accept logs-only OTLP replay. Document any replicated input as best effort and non-authoritative. Do not modify production forwarding.

**Delivered:** bounded single-request OTLP/JSON offline replay, deterministic coverage accounting and a source contract that reserves replicated input for an explicit best-effort, non-authoritative source.

### PF-1002: Redacted, truth-labelled telemetry fixture framework
**Status:** Complete (2026-10-13).
Create deterministic synthetic fixtures with coverage, secret-shaped values, prompt-injection strings, high cardinality and expected grouping truth.

**Delivered:** seeded synthetic multi-service OTLP logs with independently represented C1 truth, adversarial near-duplicates, high-cardinality fields, large values and fake secret-shaped values. The versioned redaction boundary precedes all current derivation.

### PF-1003: Canonicalisation and bounded exact-duplicate opportunity analysis
**Status:** Complete (2026-10-13).
Measure report-only C1 opportunities with collision-safe identity, hard state caps and visible overflow accounting.

**Delivered:** typed, order-independent canonical identities with full equality verification after SHA-256 bucketing, explicit entry, state-byte and identity-byte limits, deterministic JSON reporting and C1 benchmarks. C1 is report-only and does not modify replay input.

### PF-1004: Report-only template analysis benchmark
Benchmark source-provided templates and inferred templates. Measure unsafe merges, novelty preservation and cross-instance agreement. No template-driven transformation.

### PF-1005: Shadow coverage and economics accounting
Record accepted input, capture drops, malformed records, state overflow, raw canonical bytes and separately measured wire or contract-specific billing estimates.

### PF-1006: Evidence manifest and report provenance model
Implement only the manifest and report provenance needed by V0. Defer evidence content storage and archival.

### PF-1007: Coverage-aware shadow reduction report
Produce a reproducible versioned report with C1/C2 opportunities, safety exclusions and no generic dollar-saving claim.

### PF-1008: Fault and performance benchmark harness
Test replay, restart, blocked analysis output, malformed data, giant payloads, high cardinality and bounded-resource behaviour.

## Phase 2: Safe compaction

### PF-2001: compaction policy engine
### PF-2002: representative sample retention
### PF-2003: cold evidence archive adapter
### PF-2004: reconstruction/drill-down support
### PF-2005: active compaction safety gates

## Phase 3: Incident intelligence

### PF-3001: signal event model
### PF-3002: deterministic detectors
### PF-3003: incident state machine
### PF-3004: incident evidence graph
### PF-3005: generic webhook notification
### PF-3006: Slack notification adapter
### PF-3007: external agent event contract

## Phase 4: Agent investigation

### PF-4001: model provider interface
### PF-4002: agent connector interface
### PF-4003: investigation tool contract
### PF-4004: evidence constrained investigator
### PF-4005: model comparison harness
### PF-4006: prompt injection evaluation suite

## Phase 5: Pilot quality

### PF-5001: install/doctor workflow
### PF-5002: upgrade path
### PF-5003: resource limits
### PF-5004: failure recovery
### PF-5005: example reference environment
### PF-5006: public benchmark report
