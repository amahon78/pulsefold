# PulseFold Project State

**Status:** V0 scope frozen, Phase 1 telemetry laboratory in progress
**Working name:** PulseFold  
**Tagline:** Keep the evidence. Fold the noise.  
**Repository owner:** Personal/open source project  
**Repository checkout:** `<repository-root>`

## Locked decisions

The following are treated as locked until superseded by an ADR:

- PulseFold is cloud agnostic.
- PulseFold is model agnostic.
- OpenTelemetry is the canonical telemetry interface.
- PulseFold is useful without AI.
- The initial agent model is read only.
- Evidence must support incident conclusions.
- Shadow mode precedes active telemetry compaction.
- V0 is a redacted, coverage-aware, logs-only shadow reduction analyser.
- V0 accepts offline OTLP replay or an explicitly best-effort replicated, non-authoritative stream.
- V0 does not modify production forwarding, perform active compaction, run incidents or routing, invoke a built-in investigator or claim generic dollar savings.
- Shadow analysis and production forwarding have separate failure domains.
- Redaction precedes telemetry derivation.
- Evidence uses immutable provenance manifests with lifecycle-managed content.
- External investigation contracts precede a built-in investigator.
- Go is the core implementation language. The canonical module path is `github.com/amahon78/pulsefold`.
- LLMs do not control irreversible deletion in the initial architecture.
- Existing observability backends are supported rather than replaced initially.
- The first technical problem is safe telemetry optimisation.
- Proactive incident detection and routing are first class features.
- Humans and external agents are both valid incident consumers.
- Development uses personal infrastructure and synthetic or public data.

## Proposed but not yet locked

- Build on or beside the OpenTelemetry Collector.
- SQLite for lightweight local state.
- Apache 2.0 licence.
- Slack and generic webhook as first notification sinks.
- Generic webhook/MCP as first external agent integration.
- Docker as the primary first run deployment experience.

## Open decisions

- Collector distribution versus standalone gateway after V0.
- Exact compaction algorithm stack.
- Content storage, key management and retention implementation for evidence manifests.
- Initial metric and trace support depth.
- Incident correlation algorithm.
- Built in investigator scope after external-agent evaluation.
- Configuration schema.
- Licence.
- Public benchmark corpus.
- Minimum supported deployment footprint.

## Phase 0 progress notes

- 2026-10-12: PF-0001..PF-0005 research and PF-0006 independent architecture verification accepted. PF-0007 froze V0 and accepted ADRs 0007 through 0010.
- 2026-10-12: ADR 0011 accepted Go and `github.com/amahon78/pulsefold` as the core implementation language and module path. PF-1001 accepted a bounded offline OTLP/JSON replay boundary. It is not a production forwarding path.
- 2026-10-13: PF-1002 added deterministic, truth-labelled synthetic OTLP fixtures and the versioned `v1-fixture-safety` redaction boundary. PF-1003 added bounded, collision-safe, report-only C1 analysis and the initial machine-readable report command. PF-1004 added bounded C2 structural candidate analysis after C1, with resource/scope and severity partitions, no raw parameter extraction, visible exclusions and synthetic correctness, partition and performance tests. PF-1005 added explicit captured-byte and record-denominator knownness, measured C1/C2 coverage, counterfactual and inferred classifications, unknown wire accounting and optional operator-supplied generic billing estimates using exact minor-unit arithmetic. It does not measure wire bytes or claim customer or net savings. PF-1006 added a payload-free evidence manifest to every report with fixed lineage, opaque source references, validated source labels, `not_stored` content state and an unauthenticated SHA-256 integrity commitment. It stores no evidence content and does not sign or authenticate reports.
- Phase 1 benchmarks remain execution gates. PF-1004 benchmark results do not imply active-compaction safety, global cross-instance template stability or production readiness.

## Phase 0 exit condition

Phase 1 implementation begins with the frozen V0 scope. Active compaction and later phases remain gated on:

1. compaction approaches being benchmarked on representative datasets before active transformation
2. evidence content storage and retention design being accepted before content archival
3. threat-model and evaluation gates passing before the applicable release
4. deterministic signal and incident evaluation being accepted before Phase 3
5. external-agent and model evaluation being accepted before Phase 4
