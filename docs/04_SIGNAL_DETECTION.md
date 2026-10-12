# Signal Detection

**Phase boundary:** Deterministic proactive detection remains a core objective for Phase 3. V0 does not emit operational signals, open incidents or notify destinations.

## 1. Principle

PulseFold does not wait to be asked what is wrong.

It continuously evaluates structured telemetry and emits a signal when the evidence warrants attention.

## 2. Separation from notification

A signal is not a Slack message.

Signals are machine level observations.

Examples:

```text
new_error_template
error_rate_spike
latency_rate_change
slo_burn
deployment_temporal_correlation
known_failure_recurrence
```

The incident engine decides whether the signal belongs to an existing incident or warrants a new incident.

## 3. Detector classes

### Threshold detector
Absolute value exceeds a configured threshold.

### Ratio detector
Error/total or failure/success ratio changes.

### Rate-of-change detector
Fast change from baseline.

### Novelty detector
New fingerprint/template appears.

### Burst detector
Known event increases rapidly.

### Absence detector
Expected activity disappears.

### Correlation detector
Multiple signals co-occur around a deployment, dependency event or resource change.

## 4. Agentic detection

Post-V0 and only after deterministic detector evaluation. External-agent contracts precede any built-in investigator.

An LLM should not continuously inspect all telemetry.

Instead:

1. deterministic/statistical detectors emit structured candidates
2. PulseFold assembles a bounded context package
3. optional agent evaluates whether combined weak signals warrant investigation
4. evidence remains attached

This controls cost and reduces model induced noise.

## 5. Signal schema

Proposed:

```yaml
signal_id:
detector_id:
type:
observed_at:
window:
severity_hint:
resource:
baseline:
observed:
evidence_ids: []
dimensions: {}
correlations: []
```

## 6. Suppression

Required:
- cooldowns
- deduplication
- maintenance windows
- known noise patterns
- incident-aware suppression
- flapping controls

## 7. Evaluation

Synthetic streams should measure:
- precision
- recall
- detection delay
- duplicate signal rate
- false opening rate
