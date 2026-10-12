# PulseFold Project State

**Status:** V0 scope frozen, Phase 1 telemetry laboratory ready
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
- LLMs do not control irreversible deletion in the initial architecture.
- Existing observability backends are supported rather than replaced initially.
- The first technical problem is safe telemetry optimisation.
- Proactive incident detection and routing are first class features.
- Humans and external agents are both valid incident consumers.
- Development uses personal infrastructure and synthetic or public data.

## Proposed but not yet locked

- Core implementation language: Go.
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
- Phase 1 benchmarks remain execution gates. No product implementation or benchmark result is implied by this status.

## Phase 0 exit condition

Phase 1 implementation begins with the frozen V0 scope. Active compaction and later phases remain gated on:

1. compaction approaches being benchmarked on representative datasets before active transformation
2. evidence content storage and retention design being accepted before content archival
3. threat-model and evaluation gates passing before the applicable release
4. deterministic signal and incident evaluation being accepted before Phase 3
5. external-agent and model evaluation being accepted before Phase 4
