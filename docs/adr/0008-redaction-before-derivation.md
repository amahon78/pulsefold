# ADR 0008: Redaction Before Derivation

**Status:** Accepted

## Context

Fingerprints, templates, content commitments, state snapshots, reports and model inputs are derived from telemetry. If they are created before redaction, secrets and personal data can persist even when later output is redacted.

## Decision

Apply the approved redaction policy before any PulseFold derivation from telemetry, including canonicalisation, fingerprinting, template training, content commitments, state persistence, reports and external model or agent exposure.

Derived outputs record the redaction-policy version. Raw wildcard and parameter extraction are disabled by default unless an approved policy explicitly permits them.

## Consequences

Positive:

- derived artefacts do not become an unintended secret store
- reports and evidence can identify the policy under which they were created
- redaction is testable at one explicit boundary

Negative:

- redaction can change grouping and cardinality results
- redaction quality remains a security control that needs tests and review
- an approved forwarding policy may differ from the analysis policy, but each must be explicit
