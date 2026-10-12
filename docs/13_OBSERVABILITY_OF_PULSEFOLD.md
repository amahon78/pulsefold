# Observability of PulseFold

PulseFold must expose its own health clearly.

## Required self metrics

V0 analyser:
- records accepted from replay or replica input
- raw canonical bytes analysed
- malformed or rejected records
- replica capture drops where observable
- state entries and bytes
- state overflow and analysis failures
- report generation latency

Future telemetry plane:

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

Future intelligence plane:
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
