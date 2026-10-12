# Product Requirements

## PR-001: Vendor neutrality

PulseFold must accept and emit open telemetry formats where practical.

The common path must not require a specific:
- cloud
- observability vendor
- model provider
- agent framework
- notification provider

## PR-002: Shadow mode

PulseFold must support a mode where:
- all inbound telemetry is forwarded unchanged
- optimisation opportunities are analysed
- no event is dropped or transformed for downstream output
- a reduction opportunity report is produced

Shadow mode is the default for new installations.

## PR-003: Deterministic compaction

Initial active compaction decisions must be made by deterministic or reproducible algorithms.

LLM judgement must not decide irreversible telemetry deletion.

## PR-004: Evidence preservation

Every compacted group must retain enough metadata to understand:
- the template or fingerprint used
- event count
- first and last observed times
- representative samples
- relevant dimensions
- transformation policy
- raw evidence location where configured

## PR-005: Rare event protection

PulseFold must favour preservation of unique and rare telemetry.

Any policy that can discard rare events requires explicit configuration and a warning.

## PR-006: Fail open telemetry behaviour

Where possible, a PulseFold processing failure should result in unmodified forwarding rather than silent telemetry loss.

If fail open cannot be guaranteed for a mode, the limitation must be explicit.

## PR-007: Proactive signals

PulseFold must be capable of generating a signal without a user asking a question.

Examples:
- a new error template
- rapid error rate increase
- unusual rate of change
- SLO burn
- burst of a known failure pattern
- correlated change after deployment

## PR-008: Stateful incidents

Signals should be correlatable into an incident with:
- stable incident ID
- current state
- severity
- opening time
- last update
- evidence references
- affected entities
- related changes
- investigation state
- resolution state

## PR-009: Multi destination routing

Incident events must be routable to:
- generic webhook
- human notification destinations
- external agents

Slack is a desired early adapter but cannot be a core dependency.

## PR-010: External agent support

PulseFold must support organisations that already operate an agent.

PulseFold should be able to emit a structured investigation request instead of requiring use of its built in investigator.

## PR-011: Built in model abstraction

If PulseFold ships a built in investigator, provider selection must occur through an interface.

No provider specific assumptions may leak into the evidence, incident or telemetry core.

## PR-012: Read only initial agent permissions

Initial investigation tools must be read only.

Production mutation requires a later explicit capability and permission model.

## PR-013: Evidence constrained conclusions

Material investigation claims must cite machine resolvable evidence IDs.

The system should distinguish:
- observed fact
- derived correlation
- hypothesis
- recommendation

## PR-014: Easy first run

Target common path:

1. start PulseFold
2. point OTLP traffic at PulseFold
3. configure existing destination
4. run shadow mode
5. receive a report

The target for a technically competent user is under ten minutes excluding external vendor credential creation.

## PR-015: AI optional

When no model provider is configured:
- telemetry optimisation works
- deterministic signal detection works
- incident state works
- notifications work
- evidence is available

## PR-016: Auditability

Configuration changes, compaction policy changes and incident investigation actions should be auditable.

## PR-017: Secret safety

The system must support telemetry redaction before model exposure.

## PR-018: Resource control

Operators must be able to set:
- memory limits
- queue limits
- processing concurrency
- model call budgets
- investigation rate limits
