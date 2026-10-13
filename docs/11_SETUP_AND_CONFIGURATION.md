# Setup and Configuration

## 1. Product requirement

A useful first run must be simple.

The common path should not require users to understand PulseFold internals.

## 2. Implemented PF-1001 offline input

PF-1001 accepts a single UTF-8 OTLP/JSON protobuf mapping of an OTLP `ExportLogsServiceRequest`. The accepted source format identifier is `otlp-json-v1`. A whitespace-only file is an explicit empty replay. A malformed request fails deterministically without including raw telemetry in the error.

The current replay boundary has a default 64 MiB input-byte cap before decoding. If the cap is exceeded, the input is rejected and coverage records that the byte limit was reached. Per-record malformed or unsupported values are rejected individually and never counted as accepted. The analysis interval is the minimum and maximum non-zero OTLP log timestamps, not local processing time.

Offline replay identifies its source and optional partition. The same source interface reserves `best_effort_replica` for a future transport implementation, but it must explicitly declare itself non-authoritative. PF-1001 implements no network listener and no production forwarding.

PF-1002 applies redaction policy `v1-fixture-safety` before any canonicalisation or C1 analysis. It replaces values of attributes whose names indicate credentials and replaces strings that match a small, documented fake-fixture policy, including bearer-token and key/value secret shapes. The policy is deterministic and intended to validate the architecture, not to claim universal DLP coverage. Its version appears in successful reports.

PF-1003 currently exposes the local workflow:

```bash
go run ./cmd/pulsefold analyse ./telemetry.otlp.json > pulsefold-report.json
```

The command emits JSON for both a complete analysis and a safe, classified input failure. A failure exits non-zero. It never forwards the replayed data or prints raw telemetry in its diagnostics.

PF-1005 optionally accepts a local, strict JSON billing model:

```bash
go run ./cmd/pulsefold analyse ./telemetry.otlp.json --billing ./billing.json > pulsefold-report.json
```

The model uses `pulsefold/billing-model/v1`, one generic metric (`captured_input_bytes`, `accepted_records`, `c1_canonical_bytes`, `wire_bytes` or `retained_byte_days`), an indexed fraction and either a fixed rate or marginal tiers. Prices are integer minor units per integer unit block. `wire_bytes` and `retained_byte_days` are recognised but unavailable in PF-1005 because it does not measure them. The model is validated before analysis and is not a credential, vendor integration or pricing lookup. Omitting it is normal and produces an explicit unavailable billing result.

## 3. Desired V0 experience

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

## 4. Configuration principles

- one documented configuration file format
- environment variables for secrets
- no hidden defaults with destructive behaviour
- shadow mode by default
- configuration validation before start
- `doctor` command for diagnostics
- explicit provider scopes

## 5. Example conceptual config

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

## 6. Doctor checks

Potential checks:
- ports available
- write access for local state
- replay input readable and parseable
- redaction policy valid
- report destination writable
- configured state caps valid
- clock sanity
- configuration does not enable active compaction, incidents, routing or AI in V0

## 7. Deployment targets

Priority:
1. local Docker
2. generic container runtime
3. best-effort replica integration after replay behaviour is measured
4. Kubernetes after core behaviour stabilises

Cloud marketplace packaging is not an early goal.
