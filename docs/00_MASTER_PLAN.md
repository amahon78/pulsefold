# PulseFold Master Plan

## 1. Purpose

PulseFold will be an open source telemetry intelligence layer that:

1. reduces unnecessary observability volume without discarding diagnostic evidence
2. detects operational signals proactively
3. correlates signals into incidents
4. prepares structured evidence for humans and AI agents
5. optionally performs evidence constrained AI investigation
6. routes incident updates to the systems and agents organisations already use

The project is intentionally broader than log analysis but narrower than a complete Datadog replacement.

## 2. Problem statement

Engineering teams commonly pay to ingest, index and retain very large telemetry volumes even when much of that volume repeats the same operational information.

When incidents occur, engineers still perform expensive manual correlation:

- search logs
- examine metrics
- inspect traces
- identify recent deployments
- inspect code changes
- compare prior incidents
- decide whether the evidence supports a cause
- communicate status

Current AI incident tools often start after telemetry has already been collected and paid for. Current telemetry optimisation products often stop before incident reasoning.

PulseFold combines the two layers.

## 3. Product boundaries

PulseFold is:

- a telemetry optimisation layer
- a signal detection layer
- an incident evidence layer
- a notification/routing layer
- an optional agent investigation layer

PulseFold is not initially:

- a primary long term observability database
- a complete APM platform
- a dashboard replacement
- a RUM platform
- a session replay platform
- a generic autonomous remediation system
- a proprietary AI model
- a cloud specific monitoring product

## 4. User stories

### Engineering lead

> I want to know how much of my telemetry spend is repetitive before changing anything.

### Platform engineer

> I want to place PulseFold in front of my existing observability backend without changing every application.

### On-call engineer

> I want one incident with evidence and updates rather than twenty disconnected alerts.

### AI enabled engineering team

> I want my existing agent to be proactively notified when PulseFold detects something worth investigation.

### Security conscious team

> I want useful telemetry optimisation even if all external AI features are disabled.

### Small team

> I want a ten minute path from installation to a useful shadow mode report.

## 5. Product phases

### Phase 0: prove the design deserves to exist

No production implementation.

Outputs:

- competitor map
- compaction research
- OpenTelemetry integration decision
- benchmark corpus
- threat model
- evaluation strategy
- architecture review
- frozen V0 requirements

### Phase 1: telemetry laboratory

Goal:

> Analyse logs-only OTLP replay or a best-effort, non-authoritative replica and accurately estimate safe reduction potential without modifying production forwarding.

Deliverables:

- offline OTLP replay and replica-input boundary
- redacted truth-labelled fixture generator
- bounded exact-duplicate opportunity analysis
- report-only template analysis
- coverage and economics accounting
- versioned shadow reduction report
- fault and performance benchmark

Not Phase 1 deliverables:

- an inline production forwarding gateway
- active compaction or payload extraction
- incident state, routing or notifications
- a built-in investigator or model call

### Phase 2: safe compaction

Goal:

> Compact selected telemetry classes with deterministic, auditable rules and preserved evidence references.

No AI deletion decisions.

Phase 2 begins only after Phase 1 demonstrates safety, coverage and resource bounds for the chosen candidates.

### Phase 3: proactive incident intelligence

Goal:

> Detect and correlate operational signals into stateful incidents and route useful context proactively.

This remains a core product phase. It is not part of V0.

### Phase 4: agent investigation

Goal:

> Allow a chosen model or external agent to investigate incidents using safe tools and evidence constrained claims.

External-agent contracts precede any built-in investigator.

### Phase 5: operational hardening

Goal:

> Make installation, upgrades, failure recovery, resource limits and real world running boring and dependable.

## 6. Quality bar

PulseFold should be rejected as not ready if any of the following are true:

- setup requires deep knowledge of internals for the common path
- compaction cannot quantify what information was transformed
- unique events can be silently dropped
- the agent can make material claims with no evidence reference
- enabling AI is required for basic telemetry optimisation
- switching model providers requires core changes
- switching cloud providers requires core changes
- PulseFold failure can silently black-hole telemetry
- a shadow report hides incomplete coverage, state overflow or analysis failure
- benchmarks are absent
- incident evaluations have no known ground truth
- threat modelling is treated as a later task

## 7. North star metrics

Telemetry:

- raw canonical bytes
- encoded and compressed wire bytes where measured
- counterfactual reduction opportunity
- analysis coverage and exclusion rate
- unique event preservation
- compaction precision
- reconstruction/evidence availability
- processing throughput
- added latency
- CPU and memory overhead

Incident intelligence:

- signal precision
- incident recall
- alert reduction
- incident merge/split accuracy
- time to first useful notification

Agent:

- root cause accuracy
- evidence precision
- unsupported claim rate
- confidence calibration
- investigation latency
- model cost per incident
- sensitive data leakage rate
- prompt injection success rate

Product:

- time to useful shadow report
- successful first install rate
- configuration steps
- upgrade failure rate

## 8. V0 scope freeze

V0 is a redacted, coverage-aware, logs-only shadow reduction analyser. It accepts offline OTLP replay or an explicitly best-effort replicated, non-authoritative stream. It produces a reproducible report and does not modify production forwarding.

V0 does not perform active compaction, run an incident engine, route notifications, invoke a built-in investigator or make generic dollar-saving claims.

## 9. Public project narrative

PulseFold should be explainable in one sentence:

> PulseFold reduces observability noise, detects emerging incidents and gives humans or AI agents the evidence they need to investigate production systems.

The project should be demonstrable without reference to any employer, private system or private telemetry.
