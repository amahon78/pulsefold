# PulseFold Agent Operating Contract

This file applies to any AI agent working in this repository, including Grok, Codex, Claude, Gemini and future agents.

## 1. Mission

Build PulseFold as a serious open source infrastructure product, not a demo and not a prompt wrapper.

The product must solve real operational problems, be easy to deploy, remain cloud agnostic and remain model agnostic.

## 2. Non negotiable architecture constraints

1. OpenTelemetry is the canonical telemetry representation unless an ADR explicitly changes this.
2. The core must not depend on a single cloud provider.
3. The core must not depend on a single model provider.
4. PulseFold must provide useful telemetry optimisation without AI enabled.
5. LLMs must not make irreversible retention or deletion decisions in the initial product.
6. Read only investigation is the default.
7. Incident claims must be linked to evidence.
8. Shadow mode must exist before active compaction.
9. Real customer data, employer data, credentials and proprietary source code must never enter public fixtures or tests.
10. Every architectural change that contradicts a locked decision requires an ADR before implementation.

## 3. Engineering behaviour

Do not:

- silently change scope
- silently change data semantics
- introduce a new runtime or database without justification
- add provider specific logic to core packages
- optimise for a demo at the expense of production correctness
- claim a feature works without tests
- create broad abstractions without a concrete use case
- add an LLM call where deterministic code is sufficient
- merge unrelated work into one change
- treat model output as ground truth

Do:

- read the relevant specification before editing
- state assumptions
- identify conflicts between requirements
- prefer small reversible changes
- add tests with implementation changes
- add benchmarks where performance is part of the requirement
- update documentation when behaviour changes
- surface uncertainty explicitly
- preserve evidence and auditability

## 4. Work protocol

Before implementation:

1. Identify the task ID from `TASK_BOARD.md`.
2. Read the linked specification and ADRs.
3. Confirm inputs, outputs, invariants and failure behaviour.
4. Propose a short implementation plan.
5. Implement only the requested scope.
6. Run tests and required benchmarks.
7. Record results.
8. Identify unresolved risks.

## 5. Agent roles

### Grok

Primary use:

- broad research
- competitive analysis
- paper and algorithm review
- open source ecosystem review
- adversarial architecture critique
- threat scenario generation
- benchmark dataset discovery
- synthetic incident catalogue design

Grok should prefer research artefacts and decision inputs over repository wide implementation.

### Codex

Primary use:

- repository aware engineering review
- implementation after specifications are accepted
- OpenTelemetry integration
- Go implementation
- test harnesses
- benchmark harnesses
- CI
- security critical review
- refactoring
- release engineering

Codex should not spend large token budgets rediscovering market or ecosystem research already assigned to Grok.

## 6. Definition of done

A task is not done because code exists.

A task is done when:

- acceptance criteria pass
- tests pass
- applicable benchmarks pass
- failure modes are handled
- documentation is updated
- security implications are considered
- there are no known silent data losses
- outputs are reproducible enough for review
