# Setup and Configuration

## 1. Product requirement

A useful first run must be simple.

The common path should not require users to understand PulseFold internals.

## 2. Desired V0 experience

Conceptual:

```bash
docker run ... pulsefold shadow
```

or:

```bash
pulsefold init
pulsefold doctor
pulsefold shadow
```

Then point an OTLP exporter at PulseFold and configure the existing destination.

## 3. Configuration principles

- one documented configuration file format
- environment variables for secrets
- no hidden defaults with destructive behaviour
- shadow mode by default
- configuration validation before start
- `doctor` command for diagnostics
- explicit provider scopes

## 4. Example conceptual config

```yaml
mode: shadow

ingest:
  otlp:
    grpc: 0.0.0.0:4317
    http: 0.0.0.0:4318

export:
  otlp:
    endpoint: http://existing-collector:4317

compaction:
  exact_duplicates: true
  template_analysis: true
  active: false

incidents:
  enabled: true

notifications:
  webhook:
    enabled: false

ai:
  enabled: false
```

This is illustrative, not a final schema.

## 5. Doctor checks

Potential checks:
- ports available
- exporter reachable
- credentials readable
- write access for local state
- object archive connectivity
- notification connectivity
- model endpoint connectivity
- clock sanity
- configuration conflicts

## 6. Deployment targets

Priority:
1. local Docker
2. generic container runtime
3. Kubernetes after core behaviour stabilises

Cloud marketplace packaging is not an early goal.
