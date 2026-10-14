# ADR 0016: Active Behaviour Identity and Recovery

**Status:** Accepted

## Context

Destructive outcomes depend on software, policy, redaction and limits. Recovery and identity are distinct concerns but must appear together in lineage.

## Decision

Every destructive lineage identifies release/version, source commit, dirty or unavailable source state, destructive-policy identity, analysis and forwarding redaction identities, behaviour-changing configuration and cap identity, tenant/environment identity, and billing-model commitment when economics are emitted. Build time is not deterministic identity. Policy changes require authorised, append-only auditable records; concurrent versions remain distinguishable in held state and evidence.

Each active policy declares a verified recovery mechanism and period. Permanent raw archive is not required. The initial active experiment uses a secondary path with the existing observability backend authoritative. A bounded archive or another recovery mechanism requires its own retention, access and key-management design before it becomes authoritative.

V0 SHA-256 integrity remains unauthenticated consistency checking. Local reproduction and verification are the initial trust mechanism. Authentication is reconsidered when a destructive decision relies on a report, policy or evidence assertion crossing an independent trust boundary that cannot be locally reproduced.
