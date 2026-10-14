# Active-path safety and delivery contract

## Status

Phase 1.5 architecture. It does not authorise active compaction, forwarding integration or PF-2001.

## Contract

```text
input -> admission -> eligibility -> bounded analysis -> decision
      -> held ownership, if a window is used -> output or bypass
      -> evidence and accounting
```

Each record is partitioned by tenant/environment and policy version. Admission either accepts the record or records an observable refusal. Eligibility applies ADR 0012 before any transformation. Analysis uncertainty, a cap, timeout, asserted-metadata influence, a rare/error/security rule, invalid evidence or recovery state, or any unknown coverage state selects bypass.

The output state for every admitted record is exactly one of: unchanged forwarding, durably owned pending delivery, committed transformed output, rejected before admission, or unknown/incomplete. No state may be silently omitted.

## Held-window ownership

A held member has one owner and one policy/configuration version. Configuration changes do not reinterpret existing held members. A bypass, deadline or rollback either forwards held members unchanged before an aggregate is committed, or follows the already committed delivery/recovery record. It must never emit both an aggregate and its members without an explicit at-least-once duplicate record.

Process crash, downstream failure and multi-instance failover have no implied lossless guarantee. A supported topology must define durable ownership, retry and acknowledgement behaviour and demonstrate it with fault tests before active enablement. The secondary-path first experiment leaves the existing backend authoritative.

## Invariants

- No silent loss, unreported transformation or hidden partial transformation.
- Invalid configuration prevents destructive mode, not ordinary forwarding.
- Resource exhaustion and uncertainty bypass unchanged telemetry where forwarding remains possible.
- Bypass reason, queue/admission state, held count, duplicate outcome and destination failure are observable.
- C2 and telemetry assertions cannot authorise a decision.
- Accounting and evidence use the same policy, tenant/environment, window and instance scope as delivery.
