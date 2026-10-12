# Observability of PulseFold

PulseFold must expose its own health clearly.

## Required self metrics

Telemetry plane:
- events received
- bytes received
- events exported
- bytes exported
- events compacted
- events passed through
- processing latency
- queue depth
- dropped events
- fail-open events
- exporter failures

Intelligence plane:
- signals emitted
- incidents open
- investigation requests
- investigation failures
- model calls
- model latency
- model cost estimate
- notification failures

## Logs

PulseFold logs should be structured and must avoid logging telemetry payloads by default.

## Health

Separate:
- liveness
- readiness
- exporter health
- intelligence provider health

An unavailable AI provider should not necessarily make the telemetry gateway unready.
