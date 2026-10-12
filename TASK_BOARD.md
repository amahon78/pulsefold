# PulseFold Task Board

## Phase 0: Research and architecture

### PF-0001: Competitive landscape
**Owner:** Grok  
Map direct and adjacent projects across observability cost reduction, log compaction, AI SRE, incident intelligence and agent integrations.

**Output:** `docs/research/competitive-landscape.md`

**Acceptance criteria:**
- distinguishes direct competitors from components and adjacent tooling
- identifies capabilities we should reuse rather than rebuild
- identifies at least five credible differentiation risks
- includes licence and activity information where available

### PF-0002: Log compaction algorithm review
**Owner:** Grok  
Research exact deduplication, template mining, dictionary/content addressed payload storage, cardinality estimation, adaptive sampling and anomaly preserving retention.

**Output:** `docs/research/log-compaction-research.md`

### PF-0003: OpenTelemetry extension strategy
**Owner:** Grok, reviewed by Codex  
Determine whether PulseFold should ship as:
- a custom OpenTelemetry Collector distribution
- one or more Collector processors/connectors
- a standalone OTLP proxy
- a hybrid

**Output:** `docs/research/otel-extension-options.md`

### PF-0004: Public benchmark corpus
**Owner:** Grok  
Find public datasets and define synthetic generators for repetitive logs and known failure scenarios.

**Output:** `docs/research/benchmark-corpus.md`

### PF-0005: Threat model challenge
**Owner:** Grok  
Challenge `docs/08_SECURITY_THREAT_MODEL.md`, focusing on telemetry injection, prompt injection, secrets, PII, SSRF, tool abuse, credential scope and multi tenant risks.

### PF-0006: Architecture review
**Owner:** Codex  
Review all accepted Phase 0 research and challenge `docs/02_ARCHITECTURE.md`.

**Output:** architecture review notes plus ADR proposals where required.

### PF-0007: V0 scope freeze
**Owner:** Alex/Chad  
Freeze the first executable scope after research.

## Phase 1: Telemetry laboratory

### PF-1001: OTLP receive/pass-through skeleton
### PF-1002: telemetry fixture framework
### PF-1003: exact duplicate baseline
### PF-1004: template mining benchmark
### PF-1005: repeated payload extraction benchmark
### PF-1006: evidence reference model
### PF-1007: shadow mode reduction report
### PF-1008: performance benchmark harness

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
