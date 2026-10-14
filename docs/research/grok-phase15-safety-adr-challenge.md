# Grok Phase 1.5 Safety ADR Challenge

**Reviewer:** Grok (adversarial challenge, PULSEFOLD PROMPT 006)
**Reviewed commit:** `4e9b040` (`docs: reconcile Phase 2 safety architecture`)
**Date:** 2026-10-14
**Status of this document:** challenge input for Alex's architecture decision. It is **not** a final ADR, does not start PF-2001, does not implement active compaction and does not mark Phase 2 started. Compliance caveats are in section 30.

Labels: **FACT** = demonstrated by current code, tests or a reproduced experiment. **DESIGN DECISION** = a proposed rule (Codex's or mine), not yet accepted. **OPEN QUESTION** = needs evidence. No numeric thresholds are proposed here.

## 1. Scope and constraints

Challenge the five minimum ADR areas proposed in `codex-phase2-safety-architecture-review.md` (called ADR-A to ADR-E below), revisit H-1..H-5 and M-1..M-7, independently check Codex's H-5 correction, and identify safety requirements both reviews missed. Production code was not changed.

- ADR-A: active equality, redaction eligibility and evidence semantics.
- ADR-B: C2 advisory-only and source-template provenance.
- ADR-C: non-overlapping opportunity accounting and multi-unit coverage.
- ADR-D: active-path resource admission, pass-through and bypass semantics.
- ADR-E: policy/build provenance and retention-bounded recovery.

## 2. Executive conclusion

- **FACT:** No new Phase 1 BLOCKER was found. Phase 1 remains justified as a bounded offline V0 laboratory analyser.
- **FACT (new, MEDIUM):** the telemetry-asserted `log.record.template` attribute shares C2's single state-byte budget with inferred clustering. A flood of unique asserted templates exhausted the budget and excluded every later genuine event from inferred analysis (section 11, case H5-C). It also participates in C1 identity, so it can move opportunity from the `counterfactual` C1 class into the `inferred` C2 class (case H5-A). Codex's H-5 correction is therefore **PARTIALLY CONFIRMED**: right that the attribute never forms the inferred cluster key, incomplete about its other influence.
- **DESIGN DECISION (recommended):** accept Codex's five ADR areas as the right *topics*, but none is yet worded strongly enough to constrain an implementation. The largest gaps are: (1) equality is defined on the *analysis* redaction projection, while the destination receives a *forwarding* representation that ADR 0008 allows to differ; (2) ADR-D says nothing about records already held in an aggregation window at crash, deadline or bypass; (3) "operator-authored deterministic policy" is an escape hatch through which C2 masking can return under another name; (4) "deterministic stratified samples" conflicts with docs/03 §5 and invites sample poisoning; (5) "no loss against the policy's defined preservation metrics" lets a policy author define their own pass mark.
- **DESIGN DECISION (recommended):** PF-2001 should not be authorised until the revised ADRs and the Phase 1.5 evaluation plan are accepted. Even then it should be limited to disabled-by-default policy machinery with no forwarding integration. Active compaction is not authorised.

## 3. Repository verification

- **FACT:** the review used a clean box checkout at `4e9b04040887b9bf57b469b45dd8213ccc934cd2`. The parent agent verified that this equals Alex's local `HEAD` and `origin/main` with a clean working tree. I did not query the remote myself.
- **FACT:** `git log` shows `4e9b040` → `3c192ef` → `dc11cca` → … → `542b63f`.
- **FACT:** the full diff `3c192ef..4e9b040` adds only `docs/research/codex-phase2-safety-architecture-review.md` (173 lines). No code, test or other document changed between my V0 review and Codex's reconciliation.

## 4. Reading record

Every file below was read from first line to last with the paged Read tool. Line counts are `wc -l`.

| File | Lines |
| --- | --- |
| AGENTS.md | 111 |
| README.md | 109 |
| PROJECT_STATE.md | 73 |
| TASK_BOARD.md | 146 |
| docs/00_MASTER_PLAN.md | 224 |
| docs/01_PRODUCT_REQUIREMENTS.md | 169 |
| docs/02_ARCHITECTURE.md | 228 |
| docs/03_TELEMETRY_COMPACTION.md | 170 |
| docs/08_SECURITY_THREAT_MODEL.md | 169 |
| docs/09_EVALUATION_STRATEGY.md | 144 |
| docs/10_DATA_MODEL.md | 121 |
| docs/11_SETUP_AND_CONFIGURATION.md | 108 |
| docs/12_TEST_STRATEGY.md | 77 |
| docs/13_OBSERVABILITY_OF_PULSEFOLD.md | 54 |
| docs/15_ROADMAP.md | 82 |
| docs/18_NON_GOALS.md | 43 |
| docs/19_DECISION_LOG.md | 76 |
| docs/adr/0001 … 0011 (all accepted ADRs) | 26, 15, 11, 13, 13, 11, 32, 27, 29, 27, 38 |
| docs/research/README.md | 26 |
| docs/research/adversarial-architecture-review.md | 48 |
| docs/research/benchmark-corpus.md | 41 |
| docs/research/codex-architecture-review.md | 97 |
| docs/research/codex-competitive-gaps.md | 54 |
| docs/research/codex-otel-verification.md | 56 |
| docs/research/codex-v0-recommendation.md | 76 |
| docs/research/competitive-landscape.md | 110 |
| docs/research/log-compaction-research.md | 113 |
| docs/research/otel-extension-options.md | 66 |
| docs/research/threat-model-challenge.md | 65 |
| docs/research/grok-v0-adversarial-review.md | 197 |
| docs/research/codex-phase2-safety-architecture-review.md | 173 |

Code inspected in full or at every relevant function: `internal/c2/c2.go` (all 399 lines), `internal/redaction/redaction.go` (all 168), `internal/runner/runner.go` (all 62), `internal/accounting/accounting.go` (coverage and billing, lines 190-314), `internal/c1/c1.go` (analysis loop and finalise, lines 100-180), `internal/provenance/manifest.go` and `internal/report/report.go` (version, completion and integrity fields), `internal/runner/adversarial_test.go`.

Not read, because they were not on the required list and are outside the compaction-safety question: docs/04, 05, 06, 07, 14, 16, 17, `benchmarks/README.md`, `examples/README.md`.

## 5. Methodology

1. Read the required set (section 4) and both reviews in full.
2. Re-derived each Codex FACT from code rather than from either review.
3. For each proposed invariant, asked whether it can lose evidence, be steered by telemetry, mislead on coverage or economics, fail under pressure, be irreversible, be untestable, be too weak, block useful behaviour, cost too much, or miss a trust boundary, and whether a simpler rule exists.
4. Settled the H-5 dispute with a temporary Go test through `runner.Analyse` at default configuration (section 11). It was deleted afterwards, because committing it would freeze behaviour Phase 1.5 should change.
5. Toolchain: box, linux/amd64, `GOTOOLCHAIN=auto` (go1.27).

## 6. Phase 1 status reassessment

- **FACT:** the B-1 fix (`c2.go:256-258`) is present and covered by `TestC2OpportunityDoesNotReclaimC1ExactRepetition`.
- **FACT:** nothing in this review shows a V0 report presenting a destructive conclusion, leaking telemetry content or hiding a cap: the new H5-C exclusion sets `state_cap_reached` and is counted in `excluded_records`.
- Verdict: no new BLOCKER. Phase 1 laboratory completion remains justified with the existing limitations.

## 7. H-1 to H-5 reassessment

| ID | Current fact (re-checked) | Severity | Comment on Codex's decision |
| --- | --- | --- | --- |
| H-1 | `accounting.go:280` prices C1 over analysed records; `:285` prices C2 from `CandidateRecords` (C1-weighted) to `InferredTemplateGroups`, so C2 includes C1's reduction. A limitation warns against adding them. | HIGH | Agree. Also note a unit mismatch: report-level C2 `additional_counterfactual_records` is in C1 representatives, while billing C2 is in source records. Two "C2" numbers with different units already exist in one report. |
| H-2 | `redaction.go:157-160` replaces the whole string on any match; `:128` replaces any attribute whose key matches `secretKey` (`:18`), including `max_tokens`. | HIGH | Agree with pass-through of lossy-redacted records as the first rule. Missing piece: section 12. |
| H-3 | `c2.go:40-42` treat `session=`, `transaction=` and `user=` values as identifiers; `:38` folds IP and port; `:45` folds Go file:line. | HIGH | Agree C2 cannot authorise destruction. Challenge to the escape hatch in section 13. |
| H-4 | No whole-process budget in `runner.go:25-40`. RSS magnitudes not re-measured here. | HIGH | Agree. **OPEN QUESTION:** the RSS figures still need a reproducible profile; both reviews now rely on one unreproduced measurement. |
| H-5 | See section 11. | HIGH for active use, plus a new MEDIUM V0 finding | Codex's correction is partially right. |

## 8. M-1 to M-7 reassessment

| ID | Re-checked fact | Agree with Codex? |
| --- | --- | --- |
| M-1 | `accounting.go:224-225` measures C2 coverage in C1 representatives. | Yes. |
| M-2 | No source-to-output ratio exists. | Yes, and it is HIGH for Phase 2. |
| M-3 | `c1.go:149-153` admits new groups first-come, so later identities are excluded once a cap is hit. | Yes. It also makes cap and limit values part of the result's semantics, so they belong in policy identity (section 16). |
| M-4 | `accounting.go:223,225` always pass `complete=true`; it flows to `manifest.go:247-248`. | Yes. |
| M-5 | `runner.go:16` `AnalyserVersion = "devel"`. | Yes. |
| M-6 | The `v1-fixture-safety` policy (`redaction.go:18-23`) has no patterns for `ghp_`, Basic auth, JWT, URL userinfo or `pwd=`, and never scans event name, scope name or bytes values. | Yes. It also interacts with ADR-A: a redaction *miss* leaves a record eligible and stores the secret in any retained representative. |
| M-7 | No internal deadline. Timing not re-measured. | Yes. OPEN QUESTION on magnitude. |

## 9. Low and additional findings re-checked

- L-1 to L-4: agree with Codex's dispositions.
- **FACT:** Codex's "additional findings" about `Completed` and the shared `inferred` classification for source templates are correct.

## 10. New findings in this challenge

| ID | Severity | Finding |
| --- | --- | --- |
| N-1 | MEDIUM | Source templates and inferred clusters share one `MaxStateBytes` budget (`c2.go:195-203` vs `:238-241`). Telemetry can starve inferred analysis. Reproduced (H5-C). |
| N-2 | MEDIUM | `log.record.template` is part of C1 identity (all log attributes, `c1.go:280`) but excluded from C2 shape (`c2.go:328`). Varying it moves repetition from C1 `counterfactual` to C2 `inferred` (H5-A: C1 duplicates 0, C2 beyond-C1 99). Any rule that treats the two classes differently can be steered by a telemetry field. |
| N-3 | HIGH (Phase 2 gate) | Analysis-versus-forwarding representation gap (section 24, item 1). |
| N-4 | HIGH (Phase 2 gate) | Held-window loss on crash, deadline or bypass is undefined in ADR-D (section 24, item 2). |

## 11. Independent check of Codex's H-5 correction

Code trace:

- `c2.go:277-286` `sourceTemplate` returns the first non-empty string attribute keyed `log.record.template`.
- `c2.go:184-207` counts it into `SourceTemplateRecords`, then per partition into `SourceTemplateGroups`. Each new key costs `len(key)+80` bytes from the **shared** `result.StateBytes` (`:195-202`). Values up to `MaxTemplateBytes` (64 KiB) are accepted.
- `c2.go:226` and `:328` build the inferred cluster key from body tokens plus `attributeShape`, which skips `log.record.template`. The attribute never forms or merges an inferred group.
- `c2.go:238-241` excludes a new inferred cluster when `StateBytes+cost` exceeds the same budget.
- The redaction key pattern (`redaction.go:18`) does not match `template`, so the attribute is unredacted unless its value matches a secret pattern.

Reproduced cases (default config, one resource, severity 9; temporary test, since deleted):

| Case | Input | Result |
| --- | --- | --- |
| H5-A | 100 records, body `heartbeat ok`, 100 distinct asserted templates | C1 duplicates 0, C2 groups 1, beyond-C1 99, source-template groups 100 |
| H5-B | 100 distinct bodies, one asserted template | source-template groups 1, inferred groups 100 |
| H5-C | 260 records with unique ~64 KB templates, then 2,000 with unique ~50-byte templates, then 50 distinct `payment failed kindN` events with no template | `state_cap_reached=true`, `excluded_state_cap=50`: all 50 genuine events excluded; inferred groups 1; beyond-C1 2,259 |
| H5-D (control) | the same 50 events alone | 50 inferred groups, no exclusion |

**Verdict: PARTIALLY CONFIRMED.** Codex is correct that the attribute does not form or merge inferred C2 clusters (`c2.go:328`). Codex is incomplete in saying it only poisons the source-template count: it also (a) starves inferred analysis through the shared state budget (N-1) and (b) shifts opportunity between the C1 and C2 classes (N-2). Both are visible in V0 (`state_cap_reached`, exclusions) and non-destructive, so they are not BLOCKERs.

## 12. ADR-A challenge: equality, redaction eligibility, evidence

Codex's rule: an active policy passes through any record whose equality-relevant fields were changed by lossy redaction, and may replace N records only when every destination-required semantic is equal, recoverable or explicitly out of contract.

- **Can it lose evidence? Yes, through the representation gap.** ADR 0008 (line 27) allows the forwarding redaction policy to differ from the analysis policy. Equality is computed on the analysis projection, but the destination receives the forwarding representation. Two records equal under analysis redaction can differ under forwarding redaction, and the policy would then emit one of them as if it stood for both. **DESIGN DECISION (recommended):** equality used for an active decision must be evaluated over the representation the destination will receive, or the policy must be restricted to cases where analysis and forwarding representations are proven identical. Which one is preferable is an **OPEN QUESTION**, because the first conflicts with ADR 0008 as written.
- **Attacker manipulation:** injecting secret-shaped text makes records ineligible. That fails safe for evidence but lets an attacker suppress savings or force cost-increasing pass-through (section 15). Avoiding redaction triggers (M-6 forms) keeps records eligible and can leave a secret in a retained representative.
- **Too weak:** "changed by a lossy redaction operation" needs a field-level definition that covers the body, nested kvlists and arrays, and a counter that is reported. Without that counter the rule cannot be tested. **DESIGN DECISION (recommended):** require a reported count of records made ineligible by redaction, per field class.
- **Timestamps:** Codex correctly says first/last plus count does not equal per-event timestamps. The ADR must state that any policy discarding per-event timestamps is an *aggregation contract* the operator accepts, not "exact" equality.
- **Simpler rule:** for the first active policy, restrict eligibility to records where analysis redaction changed nothing **and** forwarding and analysis representations are byte-identical. This is narrow and testable, and it defers keyed tokens, as Codex recommends.
- Keyed tokens: agree they stay an **OPEN QUESTION** requiring an ADR 0008 amendment.

## 13. ADR-B challenge: C2 advisory-only and source-template provenance

- **"Permanently" advisory:** a permanent ban is stronger than the evidence requires. It is defensible today because C2 masking is regex-based and unevaluated on real corpora. However, Codex's escape hatch, "a later operator-authored deterministic policy may use a separately evaluated schema", lets the same masking rules come back if an operator copies C2's output into a policy. **DESIGN DECISION (recommended):** wording should be "C2 inference output can never be an input to a destructive policy. Any structural or field-masking policy, however authored, must pass the same evaluation gates as a C2-based policy would, and its provenance must state whether C2 suggested it."
- **Advisory output can still cause harm.** H5-A and H5-C show telemetry steering C2 counts. An advisory "candidate policy" ranked by those counts is a persuasion channel aimed at the operator (threat-model-challenge §3, savings-report manipulation). **DESIGN DECISION (recommended):** every advisory C2 figure carries its coverage and trust class, and C2 never emits a machine-applicable policy artefact.
- **Source-template provenance:** agree on `telemetry_asserted`. Add: **DESIGN DECISION (recommended)** separate state budgets (or separate stages) for asserted templates and inferred clustering, so asserted data cannot starve inference (N-1). Codex's "authenticated producer contract" has no OTLP mechanism at record level. Transport authentication identifies a sender, not whether an attribute's value came from user input. Whether such a contract is possible at all is an **OPEN QUESTION**.

## 14. ADR-C challenge: non-overlapping accounting and multi-unit coverage

- **Agree** with sequential source → C1 output → declared post-C1 transform, and with withholding C2 currency estimates.
- **Unit trap (FACT):** the report already holds C2 numbers in two units (section 7, H-1). The ADR must forbid a currency figure in any unit other than the source unit the meter bills.
- **Coverage vector risk:** a vector is honest but easy to quote selectively, just as "Integrity: verified" can be (L-2). **DESIGN DECISION (recommended):** every opportunity figure must be emitted together with its source-record denominator and weakest classification in the same object, so it cannot be quoted without them.
- **Attacker inflation:** opportunity created by redaction (H-2) or by asserted attributes (N-2) looks the same as genuine repetition. **DESIGN DECISION (recommended):** report the share of opportunity drawn from redaction-changed records and from partitions carrying telemetry-asserted metadata.
- **Removing priced C2 from active decisions does not block useful behaviour**, as long as unpriced C2 counts remain available for prioritisation. Agree.
- **Window representativeness** (a quiet week, single instance under load balancing) remains an **OPEN QUESTION**; the report must state the window and instance scope rather than extrapolate.

## 15. ADR-D challenge: admission, pass-through and bypass

- **Held-window records (N-4).** "Pass through unchanged on uncertainty" covers new records. It does not say what happens to records already accumulated in an aggregation window when a deadline, cap, crash, OOM kill or global bypass occurs. The `log_dedup` precedent (OTV-05) loses an aggregate on a flush failure. **DESIGN DECISION (recommended):** no record is acknowledged upstream until it, or a declared aggregate containing it, has been forwarded or durably queued; on bypass, in-flight windows flush as raw records. The failure tests must cover kill -9 during a window.
- **Double emission:** bypass after a partial aggregate can deliver both the aggregate and the raw records, double-counting at the destination. The ADR needs an explicit rule, for example marking aggregates and flushing raw only for un-emitted members.
- **Pass-through amplification:** forced pass-through increases destination volume and cost. An attacker who can trigger uncertainty conditions (new groups, secret-shaped text, caps) can deliberately cause a cost spike. This is the correct safety direction, but it must be visible: report pass-through counts by reason.
- **"Fail closed at startup":** agree, with one clarification. If PulseFold is inline, "not entering active mode" must still forward. The ADR must say which component forwards when PulseFold refuses to start.
- **cgroup limit:** agree an RSS ceiling cannot be guaranteed in-process. An OOM kill of an inline component is a black hole unless the topology isolates it (ADR 0007). The ADR should require that active compaction runs only in a topology whose kill behaviour has been tested.
- **Testability:** every uncertainty condition Codex lists needs a fault-injection test. The list is testable as written.

## 16. ADR-E challenge: build/policy identity and bounded recovery

- **Build identity:** agree (release version, commit, dirty marker, build time excluded). **Policy identity must also include** the redaction policy versions (analysis *and* forwarding), every cap and limit value (they change results under capping, M-3), the billing-model commitment and the representative-selection seed policy.
- **Bounded recovery instead of a permanent archive:** agree that "forever" is unnecessary. Challenges:
  - Incidents, especially security ones, are often discovered after a short window would have expired. **DESIGN DECISION (recommended):** security and audit classes are never compacted by default. docs/03 invariant 5 says only "may be configured", which is weaker.
  - A bounded encrypted raw copy concentrates secrets (threat-model-challenge §4c). It pulls P-007 (key management, retention) onto PF-2003's critical path. That is acceptable, but the ADR should say so explicitly.
  - Codex's "secondary path while the existing backend remains authoritative" is the safest first experiment and should be the **required** first active topology, not one option among two.
  - When the window expires, reconstruction claims must flip to unavailable in the evidence lifecycle record (ADR 0009 tombstones). Agree.

## 17. Lossy-redaction eligibility (specific challenge)

Covered in section 12. Summary: the pass-through rule is necessary and fails safe, but it is not sufficient. It must be evaluated on the destination's representation, it needs a reported counter, and it does not address redaction *misses*, which leave records eligible and secrets in retained representatives. M-6 must therefore be fixed (a production redaction policy) before any representative is persisted, independently of eligibility.

## 18. Permanent C2 advisory-only status (specific challenge)

Covered in section 13. Summary: keep it as the rule, close the operator-policy loophole by gating every structural policy equally, and treat advisory output as a persuasion channel that carries trust and coverage labels.

## 19. Non-overlapping economics and removal of priced C2 (specific challenge)

Covered in section 14. Summary: agree. Add single-unit currency, inseparable denominators and attribution of redaction- and assertion-derived opportunity.

## 20. Bounded recovery versus permanent raw archival (specific challenge)

Covered in section 16. Summary: agree with bounded recovery, with never-compact defaults for security and audit classes, the secondary-path topology as the required first experiment, and an explicit dependency on P-007.

## 21. Deferring authenticated provenance (specific challenge)

- Codex: no signing before PF-2001; authentication becomes mandatory when reports cross trust boundaries, support audit claims or control a remote action.
- Challenge: the shadow report is exactly the artefact used to justify *enabling* compaction (threat-model-challenge §3). Enablement is a trust-boundary crossing whenever the report was produced by a different party, a CI job or a machine the operator does not control. A re-sealed report is indistinguishable from an original (my V0 Observation).
- **DESIGN DECISION (recommended):** deferring signing past PF-2001 is acceptable while the machinery is disabled. Before **active enablement**, either (a) enablement requires an authenticated report, or (b) enablement re-runs the analysis locally under the build-stamped binary and decides from that run only. Choosing between (a) and (b) is an **OPEN QUESTION** for Alex.
- Policy-change audit (PR-016) is a separate requirement and should not wait for report signing: a malicious insider enabling a policy to hide activity is a threat in docs/08 §3.

## 22. Representative and evidence invariants (specific challenge)

- Codex: "deterministic stratified samples ... randomised reservoir samples can supplement".
- **Conflict (FACT):** docs/03 §5 (line 88) says selection must not be predictable from record position alone, and threat-model-challenge §3 and codex-architecture-review security item 5 both require randomised selection against poisoning. Fully deterministic selection is predictable to a telemetry producer.
- **DESIGN DECISION (recommended):** separate *coverage* from *choice*. The set of strata (at least one per distinct value of declared dimensions and state transitions) is deterministic. The member chosen within each stratum is selected unpredictably to producers, using a secret seed recorded in the manifest, so the operator can reproduce it.
- Strata can be flooded: an attacker can create many dimension values. Strata need a cap, and a cap overflow means pass-through, not dropped strata.
- Exact-duplicate laundering: collapsing exact copies keeps the count but drops per-event time. An attacker can hide *when* a real event happened among copies. That is a reason to keep per-time-bucket counts (Codex includes them; agree).

## 23. Real-corpus and incident-recall gates (specific challenge)

- **FACT:** no licensed corpus is in use yet (benchmark-corpus.md; docs/09:29). The synthetic labels were written by the same authors as the algorithm (`internal/c2/evaluation.go`).
- Codex's SAFE TO COMPACT ends with "no loss against the policy's defined preservation metrics". That is circular: the policy author defines the pass mark. **DESIGN DECISION (recommended):** the preservation metrics and the labelled incident set are fixed in the evaluation plan *before* any policy is evaluated, and they are owned independently of the policy.
- Destination-query compatibility is essential. Alerts and dashboards that count lines break under aggregation (adversarial-architecture-review §2). The gate should replay the destination's own alert and query definitions where available.
- Who labels incidents, and what corpus can legally be used, are **OPEN QUESTIONS**. No threshold is proposed here.

## 24. Safety requirements missed by both reviews

1. **Analysis-versus-forwarding representation (N-3).** Equality must hold on what the destination receives (section 12).
2. **In-flight window semantics (N-4).** Acknowledgement, crash and bypass behaviour for held records (section 15).
3. **Bypass double emission** of aggregate plus raw (section 15).
4. **Attacker-induced pass-through cost amplification** must be reported by reason (section 15).
5. **Shared C2 state budget starvation by asserted templates (N-1)**, and **telemetry-controlled class shifting (N-2)**.
6. **Multi-instance consistency.** Per-instance active state under load balancing gives different decisions for identical records, and single-instance shadow analysis overstates fleet repetition (log-compaction-research §1). Neither review sets a requirement.
7. **Timestamp semantics of emitted aggregates** (OTV-05 precedent) must be declared per policy and tested against destination time queries.
8. **Policy-change authority and audit** independent of report signing (PR-016, insider threat).
9. **Environment or tenant boundary** as a mandatory part of active identity (threat-model-challenge §7; docs/03 invariant 4). C1 includes resource attributes today, but nothing requires an environment attribute to exist.

## 25. Recommended wording changes for the ADR drafts (not final ADRs)

- ADR-A: "Active equality is evaluated over the representation forwarded to the destination. A record is ineligible if any redaction changed it or if analysis and forwarding representations differ. Ineligible counts are reported by field class."
- ADR-B: "C2 inference output is never an input to a destructive policy. Every structural policy, however authored, passes the same gates. Asserted templates are `telemetry_asserted` and have a separate state budget."
- ADR-C: "Currency is reported only in the billed source unit. Every opportunity carries its source-record denominator and weakest classification in the same object. Redaction- and assertion-derived shares are reported."
- ADR-D: "Pass-through on uncertainty covers held records. Nothing is acknowledged upstream before forwarding or durable queueing. Bypass flushes in-flight members raw without double-emitting. Pass-through counts are reported by reason. Active mode runs only in a topology whose kill behaviour is tested."
- ADR-E: "Policy identity includes both redaction policies, every cap, the billing commitment and the sampling seed policy. Security and audit classes are never compacted by default. The first active topology is a secondary path with the existing backend authoritative. Enablement requires an authenticated report or a local re-run (decision pending)."

## 26. Phase 1.5 minimum task set (proposal)

1. Draft ADR-A to ADR-E with the section 25 wording, for Alex's acceptance.
2. Decide the analysis-versus-forwarding representation question (section 12).
3. Write an active-path contract: inputs, outputs, held-window semantics, bypass, coverage vector and provenance fields.
4. Write the evaluation plan: corpus licensing, independent incident labelling, preservation metrics fixed in advance, destination-query replay.
5. Reproduce H-4 and M-7 with a committed, repeatable profile so budgets rest on reproducible measurements.
6. Record the choice between an authenticated report and a local re-run for enablement.

## 27. Authorisation position

- **DESIGN DECISION (recommended):** PF-2001 implementation: **not yet**. After the ADRs and the section 26 outputs are accepted, authorise PF-2001 only as disabled-by-default policy machinery with no forwarding integration.
- Active compaction: **not authorised**.

## 28. Open questions

1. Can active equality be computed over the forwarding representation without violating ADR 0008, or does ADR 0008 need amending?
2. Is any record-level authentication of `log.record.template` achievable in OTLP?
3. Which recovery duration per telemetry class is defensible, and who decides it?
4. Which licensed corpus and which independent labelling process?
5. Authenticated report or local re-run for enablement?
6. What fraction of real C1 opportunity is redaction- or assertion-derived?
7. Reproduced whole-process memory and CPU figures (H-4, M-7).
8. Multi-instance active-state consistency model.

## 29. Changes made and gates run

- Added this document only. No production code, tests or other documents changed.
- One temporary test was written to settle H-5, run with `go test ./internal/runner -run TestScratchH5 -v`, and deleted before committing.
- Gates for a documentation-only change: `git diff --check` and a full diff inspection. The change was scanned for credentials, secrets, private information, local machine paths and unrelated edits.

## 30. Compliance and limitations

- The sections of this document were organised by the reviewer around every area the brief lists. The reviewer's inherited copy of the brief was truncated in the middle, so the exact required section list and items 1-16 of the required final response were not visible. Alex should check this structure against the brief's list. This is disclosed as a compliance gap.
- Remote state was verified by the parent agent, not directly by the reviewer.
- H-4 and M-7 magnitudes were not re-measured.
- The non-required documents listed in section 4 were not read.
