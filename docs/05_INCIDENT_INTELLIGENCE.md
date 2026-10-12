# Incident Intelligence

**Phase boundary:** This is a post-V0, Phase 3 capability. V0 has no incident engine, state machine or notification routing.

## 1. Incident model

An incident represents an evolving operational hypothesis supported by evidence.

It is not merely a container for alerts.

## 2. Proposed state machine

```text
candidate
   |
   v
open ---> suppressed
 |
 v
investigating
 |
 v
identified
 |
 v
monitoring
 |
 v
resolved
 |
 v
closed
```

Transitions must be explicit and recorded.

## 3. Incident fields

```yaml
incident_id:
title:
state:
severity:
opened_at:
updated_at:
resolved_at:
resources: []
affected_entities: []
signal_ids: []
evidence_ids: []
related_changes: []
investigation:
  status:
  hypotheses: []
  conclusion:
  confidence:
  evidence_ids: []
recommendations: []
routing:
  destinations: []
```

## 4. Incident correlation

Initial correlation should use deterministic dimensions where possible:

- service/resource identity
- environment
- time proximity
- template/fingerprint relationships
- trace relationships
- deployment relationships

Model assisted correlation may be added later but must be benchmarked against known ground truth.

## 5. Evidence language

Investigation outputs should separate:

### Observation
Directly supported by evidence.

### Correlation
Two observations are related in time or structure.

### Hypothesis
Possible explanation requiring further evidence.

### Conclusion
Best supported explanation after investigation.

### Recommendation
Proposed action, not a factual observation.

## 6. Proactive agent flow

This is a future flow. Deterministic signals and evidence manifests must be evaluated before it is implemented.

```text
Signal threshold reached
        |
        v
Incident opened
        |
        v
Context package assembled
        |
        +------> Slack "investigating"
        |
        v
Agent requested
        |
        v
Evidence queries
        |
        v
Investigation result
        |
        +------> Incident updated
        +------> Slack update
        +------> External agent event
```

## 7. Incident history

Resolved incidents become useful operational knowledge.

Future matching may use:
- exact fingerprints
- template overlap
- service/resource overlap
- evidence pattern similarity
- deployment pattern

The project should not jump directly to vector search without proving it is needed.
