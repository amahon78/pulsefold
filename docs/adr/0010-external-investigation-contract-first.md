# ADR 0010: External Investigation Contract First

**Status:** Accepted

## Context

Existing agent and incident platforms already provide investigation and workflow capabilities. PulseFold must remain model agnostic, useful without AI and focused on trustworthy telemetry evidence.

## Decision

PulseFold will define a model-neutral, read-only external investigation event and evidence contract before considering a built-in investigator.

V0 does not invoke agents or provide a built-in investigator. Future external integrations use bounded, redacted, evidence-referenced data and do not allow telemetry or model output to select callback URLs, tools or access scope.

## Consequences

Positive:

- avoids duplicating model orchestration before telemetry value is proven
- preserves model and agent neutrality
- makes provenance and scope part of the integration boundary

Negative:

- a built-in investigator is deferred
- external integrations need their own authentication, capability scope and evaluation design
- the event schema remains a future implementation task
