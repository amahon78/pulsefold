# ADR 0015: Active Delivery, Resource and Bypass Safety

**Status:** Accepted

## Context

State caps do not bound the current process. An inline aggregation window can lose, duplicate or partially transform held records during failure or bypass.

## Decision

An active topology has explicit admission, in-flight, state, queue, concurrency, CPU/deadline and memory controls. PulseFold measures refusals and bypasses, enforces application-level admission and deadlines, and relies on deployment limits only as external containment, not as an in-process RSS guarantee.

Resource exhaustion, analysis uncertainty, evidence/recovery failure or invalid eligibility never authorise destructive loss. Where forwarding is possible, affected records bypass unchanged with a reason. Invalid destructive configuration prevents active mode from starting while forwarding remains available.

Before upstream acknowledgement, each held record is either forwarded unchanged, durably owned by a declared delivery mechanism, or represented by one committed transformation whose evidence and destination delivery state are recoverable. Bypass or rollback flushes held members raw only when no aggregate was delivered or committed for them. A later implementation must make duplicate delivery, retry, crash, destination outage, configuration change and multi-instance ownership observable and testable.

## Consequences

"Fail open" means a defined delivery outcome, not merely dropping transformation work. The initial experiment is secondary-path only and must test kill, OOM and held-window behaviour.
