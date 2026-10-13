# Phase 2 Safety Architecture Review

**Reviewer:** Codex
**Date:** 2026-10-13
**Scope:** Reconciliation after PF-1001 through PF-1008 and Grok's V0 adversarial review. This is not authorisation for PF-2001 and makes no active-compaction change.

## Executive conclusion

**DESIGN DECISION:** Phase 1 remains complete as a V0 laboratory milestone. Its claim is limited to a bounded, offline, redacted shadow analyser. It is not evidence that an active forwarding or compaction path is safe.

**FACT:** Grok's blocker existed at `dc11cca`: C2 calculated its supposedly beyond-C1 opportunity from weighted C1 groups, thereby counting C1 duplicates again. The current calculation uses analysed C1 representatives minus inferred C2 groups, and the new regression test passes. This restores the report-only count semantics.

**DESIGN DECISION:** PF-2001 must not start yet. A short Phase 1.5 safety-architecture gate should first settle destructive-equality semantics, C2's non-authority, active-path failure and recovery semantics, accounting/coverage semantics and evidence requirements.

## Methodology and evidence limits

I read the mandatory project, architecture, security, evaluation, data-model, roadmap, non-goal and decision-log documents; ADRs 0001 through 0011; the relevant Phase 0 research; and Grok's review in full. I inspected `3c192ef`, its complete diff from `dc11cca`, each production-code and test change, and the current input, redaction, C1, C2, accounting, report and provenance paths.

I verified `HEAD` and `origin/main` are both `3c192eff37c27b0480c984b571cd1708d3a5f824`. The expected Grok commit is present. The preserved PF-1006 stash, review branch and recovery artefact were not changed.

I ran `go test ./internal/runner -run 'TestC2OpportunityDoesNotReclaimC1ExactRepetition|TestPricedC1AndC2OpportunitiesDeclareOverlap' -v`; both passed. I did not repeat Grok's large-RSS experiments. Consequently, the absence of a whole-process budget is a verified code fact, while the reported RSS magnitudes are credible but not independently reproduced here.

Primary implementation evidence is the current repository, especially [C2](../../internal/c2/c2.go), [redaction](../../internal/redaction/redaction.go), [accounting](../../internal/accounting/accounting.go), [runner](../../internal/runner/runner.go) and [provenance](../../internal/provenance/manifest.go). Upstream Collector and ecosystem claims remain pinned in [the prior source verification](codex-otel-verification.md) and [competitive review](codex-competitive-gaps.md); this review does not recertify their moving upstream versions.

## Grok blocker: C2 beyond C1

**FACT, confirmed:** Before `3c192ef`, `AdditionalOpportunity` was `CandidateRecords - InferredTemplateGroups`. `CandidateRecords` is weighted by each C1 representative's original count. Therefore 100 identical records produced one C1 representative with weight 100, one C2 group and an incorrect C2 opportunity of 99.

**FACT, confirmed:** The current expression is `RecordsAnalysed - InferredTemplateGroups`. It gives zero for one C1 representative and one C2 group, and one for two distinct C1 representatives that collapse into one C2 group. The permanent regression test covers both cases. The policy version change to `c2-report-only-no-parameters-v2` is sufficient to distinguish the changed C2 count semantics in report provenance.

**RISK:** The field is still named `additional_counterfactual_records`, although it now counts post-C1 representatives, not original source records. It is safe for the present report-only meaning, but an active contract must name both source-event and post-C1 output units explicitly.

**DECISION:** Accept the fix for V0. Do not treat it as a solution to overlapping economics.

## High findings

### H-1: C1/C2 economic overlap

**FACT, confirmed:** Billing for accepted-record meters prices C2 as `CandidateRecords` down to `InferredTemplateGroups`. `CandidateRecords` remains C1-weighted, so its estimated C2 reduction includes exact-duplicate reduction. The current limitation correctly says the priced figures must not be added.

**DESIGN DECISION:** A warning is adequate only for V0, where neither amount authorises action. Phase 2 must use non-overlapping accounting: observed source baseline, C1 post-policy output, then any separately declared post-C1 transformation. A combined figure is permissible only where all component transformations are enabled by one versioned policy, their units and coverage align, and the result is labelled by the weakest classification. C2-inferred potential must not be priced as a savings opportunity before C2 safety is established.

### H-2: redaction changes identity

**FACT, confirmed:** Whole-string replacement follows any matching secret pattern, and key-name matching replaces complete attribute values. C1 operates only on that redacted result. Distinct source records can therefore become equal. Current documentation is technically accurate when it says C1 is redacted canonical equality, but "exact duplicate" needs the explicit qualifier "after redaction" wherever it might inform a future transformation decision.

**DESIGN DECISION:** Current C1 cannot authorise destructive compaction. The first active policy should conservatively pass through any record whose equality-relevant fields were changed by a lossy redaction operation. It should use field-aware redaction metadata and report the resulting ineligible volume. This preserves ADR 0008's privacy ordering without introducing a secret-derived identity.

**OPEN QUESTION:** A keyed equality token could preserve equality without exposing a value, but it creates key-management, rotation, cross-process determinism, equality-leakage and dictionary-attack risks. An ephemeral pre-redaction comparison also contradicts ADR 0008 as written because it is a derivation. Neither should be adopted by default. A future narrowly scoped exception requires a separate threat model, key lifecycle and an ADR that explicitly amends the boundary.

### H-3: unsafe C2 generalisation

**FACT, confirmed:** The C2 regex treats `session=active`, `transaction=rollback` and similar values as dynamic identifiers. It also masks IP:port, Go locations and URL numeric segments. Different semantics can therefore share an inferred group.

**DESIGN DECISION:** C2 inference is advisory-only permanently: it may identify patterns, prioritise investigation and propose candidate policy work, but may not itself authorise dropping, aggregation, evidence destruction or a combined savings claim. A later operator-authored deterministic policy may use a separately evaluated schema, but that is not a C2-inference decision.

### H-4: whole-process resources

**FACT, confirmed:** Stage caps do not bound process memory. Replay retains the input bytes and decoded protobuf graph; redaction clones each accepted record; C1 builds canonical byte identities and stores groups; C2 clones representatives. There is no whole-process memory, CPU or deadline budget in the current runner. The reported 51 MB to 1.39 GB RSS result is plausible from these coexisting representations but needs a reproducible profile before it becomes a numeric requirement.

**DESIGN DECISION:** An active path needs independent input, in-flight, state, queue, CPU and wall-time budgets. It must reserve capacity before allocating where practicable, bound concurrency, report every refusal and bypass unchanged telemetry on exhaustion. Go can enforce application-level admission, deadlines and allocation-aware representations; it cannot guarantee a process RSS ceiling. A container or cgroup limit is a last containment boundary and may terminate the process, so forwarding isolation and restart behaviour must be tested separately.

### H-5: telemetry-controlled source templates

**FACT, partially confirmed:** `log.record.template` is telemetry-controlled and changes `SourceTemplateRecords` and `SourceTemplateGroups`. It is not used to form the inferred C2 cluster key: `attributeShape` deliberately excludes that attribute. Thus it can poison the source-template advisory count, but it does not merge inferred C2 groups in this implementation.

**DESIGN DECISION:** Source templates need explicit provenance: `telemetry_asserted`, not PulseFold-derived and not trusted. They may be displayed only as advisory metadata with source and trust class, and must never drive a destructive policy without an authenticated producer contract.

## Medium findings

| Finding | Verdict | Severity and decision |
| --- | --- | --- |
| M-1 C2 coverage counts representatives | Confirmed. JSON has weighted C2 exclusions, but stage coverage and text use C1 representatives. | Medium. Report both representative coverage and original-record weighted reach. |
| M-2 no end-to-end coverage | Confirmed. C2 denominator is C1 representatives, not accepted source records. | High for Phase 2. Define source-to-output, C1 and C2 coverage separately. |
| M-3 capped C1 is input-order dependent | Confirmed. New groups are admitted first-seen, while later matches to admitted groups continue. | High for active mode. Overflow must bypass rather than preferentially transform first-seen groups. |
| M-4 `completed:true` despite caps | Confirmed, but partially semantic. It means execution completed, not complete coverage. | Medium. Rename or pair it with `fully_covered:false` in a future schema. |
| M-5 analyser version is `devel` | Confirmed. | Medium now, high for active auditability. Add reproducible build identity before PF-2001. |
| M-6 redaction gaps | Confirmed by direct policy inspection for the listed forms and unscanned fields. | High for any persistence or active path. Fixture safety is not a production redaction policy. |
| M-7 no time budget | Confirmed. Context cancellation is supported, but no internal default budget exists. Grok's timing needs reproduction. | High for active mode. Require deadlines and benchmarked timeout behaviour. |

## Low findings

| Finding | Verdict | Decision |
| --- | --- | --- |
| L-1 benchmarks stop at 4,096 | Confirmed. | Promote to a Phase 1.5 evaluation gate because configured input limits are much larger. |
| L-2 integrity wording can be quoted selectively | Confirmed wording is accurate. | Keep low for V0; place authentication state adjacent to integrity in any machine or external contract. |
| L-3 interval is telemetry-derived | Confirmed. It is derived from accepted event timestamps and zero timestamps are ignored. Manifest lineage marks capture telemetry-derived, but the interval lacks a field-level origin. | Low. Add field-level provenance when interval affects an active policy. |
| L-4 Go-only source-location masking | Confirmed. | Low. This is conservative inconsistency, another reason C2 remains advisory. |

## Additional findings

**FACT:** C2's source-template count and inferred grouping share a report stage classified as `inferred`, despite different provenance. This is a provenance-model gap, not a V0 leak. Phase 2 must distinguish `telemetry_asserted` source templates from PulseFold-derived inference.

**FACT:** The current `StageCoverage.Completed` can flow into the manifest as `C1Completed` or `C2Completed` even where coverage is incomplete. Consumers must not read it as "all eligible data was analysed".

**DESIGN DECISION:** The Phase 2 schema should prohibit one overloaded completion flag. It should carry execution outcome, source coverage, transformation eligibility, transformed count and bypassed count separately.

## C1 safety model and equality

**FACT:** The present C1 canonical form includes redacted resource, scope, severity, body, attributes, event name, flags and trace/span IDs. It excludes timestamps and observed timestamps. It checks canonical-byte equality after SHA-256 bucketing.

The terms must remain distinct:

- **byte-identical source telemetry:** every encoded source-relevant field is equal;
- **semantically equivalent telemetry:** equal for a declared consumer purpose, an application-specific judgement;
- **canonical equality:** equality of a documented projection and canonicalisation version;
- **equality after redaction:** canonical equality after a lossy privacy transformation;
- **privacy-preserving equality identity:** a separately governed equality token that does not expose its source value.

**DESIGN DECISION:** A policy may replace N source records only when every source semantic that the destination contract requires is either equal, losslessly retained in recoverable evidence, or explicitly outside an operator-approved aggregation contract. That includes resource and scope, severity, body and attributes, trace/span relationships, event IDs, timestamp/ordering semantics and any alert/query semantics. First and last timestamps plus a count are not equivalent to retaining every event timestamp. Current C1 therefore supports a shadow counterfactual, not destructive aggregation.

## C2, economics and coverage

**DESIGN DECISION:** C2 remains a non-authoritative research and prioritisation signal. It has no destructive role. C1 and C2 economics must be shown as separate, non-additive V0 claims until a policy produces a single defined output stream. Phase 2 should initially withhold C2 currency estimates.

**DESIGN DECISION:** Coverage must be a vector, not one percentage: captured/observed source records, records eligible for each policy, records analysed, records transformed, records passed through due to a safety rule, and records unknown or rejected. Every numerator needs a declared unit: source record, C1 representative or aggregate output record.

## Representative retention and evidence recovery

**DESIGN DECISION:** Do not adopt random samples alone. A minimum deterministic evidence representation should preserve counts by time bucket, first and last occurrence, resource/scope, severity, policy and build versions, trace/span relationships where present, and deterministic stratified samples for material dimensions and state transitions. Rare values, error transitions, new policy groups and tainted or redacted records must pass through. Randomised reservoir samples can supplement this only with a documented seed, selection policy and anti-poisoning evaluation.

**DESIGN DECISION:** A full raw archive forever is not required. Before replacing an authoritative destination path, the compacted class needs a verified recovery route for the rollback/reconstruction window. The smallest credible first experiment runs compaction on a secondary path while the existing backend remains authoritative, or retains a bounded encrypted raw/cold copy for only the transformed stream and policy window. Once source retention expires, reconstruction claims must become unavailable unless the retained evidence meets the declared reconstruction contract.

## Failure, rollback and provenance

**DESIGN DECISION:** Uncertainty means pass through unchanged: cap or deadline reached, resource admission refused, unsupported input, parser ambiguity, unknown or new group, rare/error/security event, redaction/equality ambiguity, tainted source template, evidence/archive failure or provenance failure. Invalid active configuration should fail closed at startup, meaning PulseFold does not enter active mode, not that it drops telemetry.

An active path requires an immediate global bypass, per-policy disable/rollback, atomic configuration rollback, health and bypass metrics, shadow comparison against the unmodified baseline, versioned policy rollback and a tested recovery window. The forwarding process must not synchronously depend on the analyser, archive, control plane or model.

**DESIGN DECISION:** PF-2001 requires build-stamped analyser identity and policy identity. Use a reproducible release version plus source commit and a dirty/unavailable marker, with build time excluded from semantic identity. `devel` is acceptable only for local development.

**DESIGN DECISION:** Do not require PKI or report signing before PF-2001 by default. The current hash remains only consistency checking. Authentication becomes mandatory when reports, policies or evidence cross independent trust boundaries, support audit claims or control a remote action. The appropriate mechanism then depends on the deployment's control-plane and archive threat model. Transport authentication, access control and immutable audit of policy changes are separate active-path requirements.

## Evaluation gate and safe-to-compact definition

**FACT:** Current fixtures are synthetic and the public-corpus licence review remains unresolved. They cannot establish production incident recall or source-level reconstruction fidelity.

**DESIGN DECISION:** PF-2001 policy machinery may begin only after Phase 1.5 architecture decisions are accepted and remains disabled. Enabling active transformation additionally requires licensed corpus and truth-labelled incident evaluation, including duplicate precision, unsafe merges, error/rare-event preservation, redaction-created opportunity, policy bypass, resource limits, throughput, recovery and destination-query compatibility.

**SAFE TO COMPACT** means: a specific, versioned deterministic policy has a declared unit and destination contract; its eligible records have complete known coverage; it preserves or recovers every required source semantic; it bypasses unchanged telemetry for every uncertainty condition; it has a tested independent rollback and recovery route; its provenance identifies build, policy and redaction semantics; and representative incident replay demonstrates no loss against the policy's defined preservation metrics. No C2 inference alone can satisfy this definition.

## Proposed Phase 1.5

**DESIGN DECISION:** Introduce Phase 1.5: Active-compaction safety architecture, with three small outputs:

1. Accept the minimum ADR set below, including a precise equality and C2 boundary.
2. Specify active-policy input/output, coverage, bypass, recovery and provenance contracts, plus PF-2001 acceptance criteria.
3. Define the corpus, adversarial and resource evaluation plan, including licensed-data handling and no-enable gate metrics.

## ADR disposition

| Grok proposal | Disposition | Reason |
| --- | --- | --- |
| Equality under redaction | Accept concept, modify | Start with conservative exclusion of lossy-redacted records. Defer keyed tokens and any ADR 0008 exception. |
| C2 advisory-only | Accept | C2 inference must never authorise destructive action. |
| Non-overlap accounting and end-to-end coverage | Split | Accounting semantics and coverage schema are related but independently material. |
| Pass-through limits plus process budgets | Accept concept, modify | Include forwarding isolation, admission and timeout semantics. Do not promise an in-process RSS guarantee. |
| Build stamp and authenticated provenance | Split and defer authentication | Build and policy identity are required before PF-2001. Signing depends on an external trust-boundary threat model. |
| Raw/cold archive prerequisite | Modify | Require a verified, retention-bounded recovery path, not an archive forever. |

Minimum ADRs before PF-2001:

1. Active equality, redaction eligibility and evidence semantics.
2. C2 advisory-only and source-template provenance.
3. Non-overlapping opportunity accounting and multi-unit coverage.
4. Active-path resource admission, pass-through and bypass semantics.
5. Policy/build provenance and retention-bounded recovery requirements.

## Exact PF-2001 prerequisites and unresolved questions

PF-2001 is not authorised until those ADRs and Phase 1.5 outputs are accepted, including a redaction policy suitable for an active path, a policy/evidence schema, an isolated forwarding topology, explicit budget behaviour and an evaluation plan with enable gates.

Open questions requiring evidence include the fraction of C1 opportunity caused by redaction, whether a safe keyed-token exception is worthwhile, the minimum recovery duration by telemetry class, destination query/alert compatibility for aggregates, public-corpus licensing, and numeric CPU/memory/latency budgets for supported deployments.
