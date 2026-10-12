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
pulsefold replay --input telemetry.otlp
pulsefold report
```

The common V0 path uses offline OTLP replay. A replicated live input, where configured, is explicitly best effort and non-authoritative. V0 does not replace or modify the existing production forwarding configuration.

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

input:
  mode: replay # replay or best_effort_replica
  otlp_file: ./telemetry.otlp

analysis:
  exact_duplicates: true
  template_analysis: true
  max_entries: 100000
  max_bytes: 256MiB
  active_compaction: false

redaction:
  policy_version: v1

report:
  output: ./pulsefold-report.json
```

This is illustrative, not a final schema.

## 5. Doctor checks

Potential checks:
- ports available
- write access for local state
- replay input readable and parseable
- redaction policy valid
- report destination writable
- configured state caps valid
- clock sanity
- configuration does not enable active compaction, incidents, routing or AI in V0

## 6. Deployment targets

Priority:
1. local Docker
2. generic container runtime
3. best-effort replica integration after replay behaviour is measured
4. Kubernetes after core behaviour stabilises

Cloud marketplace packaging is not an early goal.
