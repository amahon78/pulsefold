# ADR 0014: Active Coverage and Non-overlapping Accounting

**Status:** Accepted

## Context

V0 C1 and C2 use different units and accepted-record billing can overlap. Active decisions need source-record traceability.

## Decision

Active reports account for each source-record partition through distinct states: observed, rejected, unknown or incomplete, eligible, analysed, transformed, and bypassed with a reason. A ratio declares numerator, denominator and unit. Representative stages retain source-record weight as well as representative counts. Coverage identifies tenant/environment, time window, instance scope, lossy-redaction ineligibility and telemetry-asserted influence.

Economics follow actual stage flow: source baseline to output of one enabled versioned policy. Figures cannot overlap or be added across incompatible units, scopes or coverage. Currency identifies meter, denominator, policy, window and instance scope. C2 may report unpriced advisory volume but has no priced active-mode savings figure.

## Consequences

`completed` cannot mean complete coverage. A future schema separates execution outcome from reach and transformation.
