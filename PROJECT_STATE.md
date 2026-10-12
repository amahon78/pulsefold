# PulseFold Project State

**Status:** Phase 0, product and architecture definition  
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

- Collector distribution versus standalone gateway.
- Exact compaction algorithm stack.
- Raw evidence retention/reference design.
- Initial metric and trace support depth.
- Incident correlation algorithm.
- Built in investigator scope versus external agent first.
- Configuration schema.
- Licence.
- Public benchmark corpus.
- Minimum supported deployment footprint.

## Phase 0 exit condition

No implementation phase begins until:

1. Competitive and ecosystem research is complete.
2. Compaction approaches are benchmarked on representative datasets.
3. The canonical data model is accepted.
4. Threat model is reviewed.
5. Evaluation methodology is accepted.
6. V0 scope is frozen.
7. The architecture has been challenged by at least two independent agent reviews.
