# Grok Phase 1.5 Safety ADR Challenge

**Reviewer:** Grok (adversarial challenge, PULSEFOLD PROMPT 006)
**Reviewed commit:** `4e9b040` (`docs: reconcile Phase 2 safety architecture`)
**Date:** 2026-10-14
**Status:** challenge input for Alex's architecture decision. This is **not** a final ADR. It does not start PF-2001, implement active compaction or mark Phase 2 started.

Labels: **FACT** = demonstrated by current code, tests or a reproduced experiment. **DESIGN DECISION** = a proposed rule (Codex's or mine), not yet accepted. **OPEN QUESTION** = needs evidence. No numeric thresholds are proposed.

The five proposed ADR areas from `codex-phase2-safety-architecture-review.md` are referred to as:

- ADR-A: active equality, redaction eligibility and evidence semantics.
- ADR-B: C2 advisory-only and source-template provenance.
- ADR-C: non-overlapping opportunity accounting and multi-unit coverage.
- ADR-D: active-path resource admission, pass-through and bypass semantics.
- ADR-E: policy/build provenance and retention-bounded recovery.

## 1. Executive verdict

- **FACT:** No new Phase 1 BLOCKER was found. Phase 1 remains justified as a bounded offline V0 laboratory analyser.
- **FACT (new, MEDIUM):** the telemetry-asserted `log.record.template` attribute shares C2's single state-byte budget with inferred clustering. A flood of unique asserted templates exhausted the budget and excluded every later genuine event from inferred analysis (section 10, case H5-C). The attribute is also part of C1 identity, so it can move opportunity from the `counterfactual` C1 class into the `inferred` C2 class (case H5-A).
- **DESIGN DECISION (recommended):** Codex's five ADR areas are the right topics. None is yet worded strongly enough to constrain an implementation. The largest gaps:
  1. Equality is defined on the *analysis* redaction projection, while the destination receives a *forwarding* representation that ADR 0008 allows to differ.
  2. ADR-D says nothing about records already held in an aggregation window at a crash, deadline or bypass.
  3. "Operator-authored deterministic policy" is an escape hatch through which C2 masking can return under another name.
  4. "Deterministic stratified samples" conflicts with docs/03 §5 and invites sample poisoning.
  5. "No loss against the policy's defined preservation metrics" lets a policy author set their own pass mark.
- **DESIGN DECISION (recommended):** PF-2001 is not authorised until the revised ADRs and the Phase 1.5 outputs are accepted. Even then, authorise it only as disabled-by-default machinery with no forwarding integration. Active compaction is not authorised.

## 2. Methodology

1. Read the required set (section 4) and both reviews in full.
2. Re-derived each Codex FACT from code rather than from either review.
3. For each proposed invariant, asked whether it:
   - can lose evidence;
   - can be steered by telemetry;
   - can mislead on coverage or economics;
   - can fail under resource pressure;
   - is irreversible;
   - is testable;
   - is too weak to constrain an implementation;
   - blocks useful behaviour or costs too much;
   - misses a trust boundary;
   - has a simpler, safer alternative.
4. Settled the H-5 dispute with a temporary Go test through `runner.Analyse` at default configuration (section 10). The test was deleted afterwards, because committing it would freeze behaviour Phase 1.5 should change.
5. Toolchain: box, linux/amd64, `GOTOOLCHAIN=auto` (go1.27).

## 3. Repository state

- **FACT:** the review used a clean box checkout at `4e9b04040887b9bf57b469b45dd8213ccc934cd2`. The parent agent verified that this equals Alex's local `HEAD` and `origin/main` with a clean working tree. I did not query the remote myself.
- **FACT:** `git log` shows `4e9b040` → `3c192ef` → `dc11cca` → … → `542b63f`.
- **FACT:** the full diff `3c192ef..4e9b040` adds only `docs/research/codex-phase2-safety-architecture-review.md` (173 lines). No code, test or other document changed between the V0 review and Codex's reconciliation.

## 4. Required-reading completion

Every file below was read from first line to last with the paged Read tool. Line counts are from `wc -l`.

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
| docs/adr/0001 to 0011 (all accepted ADRs) | 26, 15, 11, 13, 13, 11, 32, 27, 29, 27, 38 |
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

Code inspected:

- in full: `internal/c2/c2.go` (399 lines), `internal/redaction/redaction.go` (168 lines), `internal/runner/runner.go` (62 lines);
- at every relevant function: `internal/accounting/accounting.go` lines 190-314 (coverage and billing), `internal/c1/c1.go` lines 100-180 (analysis loop and finalise), the version, completion and integrity fields in `internal/provenance/manifest.go` and `internal/report/report.go`;
- `internal/runner/adversarial_test.go`.

Not read, because they are not on the required list and fall outside the compaction-safety question: docs/04, 05, 06, 07, 14, 16 and 17, `benchmarks/README.md` and `examples/README.md`.

## 5. Codex review assessment

- **FACT:** Codex's confirmations of B-1, H-1 to H-4, M-1 to M-7 and L-1 to L-4 reproduce against the current code (sections 6 and 7).
- **FACT:** Codex's additional findings are correct: the overloaded `Completed` flag (`accounting.go:223,225` → `manifest.go:247-248`) and source templates sharing the `inferred` classification.
- **FACT:** Codex's H-5 correction is incomplete (section 10).
- **FACT:** the B-1 fix (`c2.go:256-258`) is present and covered by `TestC2OpportunityDoesNotReclaimC1ExactRepetition`. No V0 report presents a destructive conclusion, leaks content or hides a cap. The H5-C exclusion sets `state_cap_reached` and is counted in `excluded_records`.
- Codex is strongest on terminology (the five kinds of equality), the uncertainty list and the "no RSS guarantee" realism.
- Codex is weakest on:
  - the analysis-versus-forwarding representation gap;
  - records held in an aggregation window;
  - the operator-policy loophole for C2;
  - deterministic sample selection;
  - a self-defined evaluation pass mark.

## 6. H-1 through H-5 reassessment

| ID | Current fact (re-checked) | Severity | Comment |
| --- | --- | --- | --- |
| H-1 | `accounting.go:280` prices C1 over analysed records. `:285` prices C2 from `CandidateRecords` (C1-weighted) down to `InferredTemplateGroups`, so the C2 figure includes C1's reduction. A limitation warns against adding them. | HIGH | Agree. Unit mismatch: the report-level C2 `additional_counterfactual_records` counts C1 representatives, while billing C2 counts source records. |
| H-2 | `redaction.go:157-160` replaces the whole string on any match. `:128` replaces any attribute whose key matches `secretKey` (`:18`), including `max_tokens`. | HIGH | Agree with pass-through as the first rule. It is not sufficient on its own (section 14). |
| H-3 | `c2.go:40-42` treat `session=`, `transaction=` and `user=` values as identifiers. `:38` folds an IP together with its port. `:45` folds Go file:line locations. | HIGH | Agree C2 cannot authorise destruction. Escape hatch: section 9. |
| H-4 | `runner.go:25-40` has no whole-process budget. RSS was not re-measured here. | HIGH | Agree. **OPEN QUESTION:** the RSS magnitudes need a reproducible profile. |
| H-5 | Section 10. | HIGH for active use, plus new MEDIUM findings N-1 and N-2 | Codex's correction is only partly right. |

## 7. M-1 through M-7 reassessment

| ID | Re-checked fact | Agree with Codex? |
| --- | --- | --- |
| M-1 | `accounting.go:224-225` measures C2 coverage in C1 representatives. | Yes. |
| M-2 | No source-to-output ratio exists. | Yes; HIGH for Phase 2. |
| M-3 | `c1.go:149-153` admits new groups first-come, first-served. | Yes. Cap values therefore belong in policy identity. |
| M-4 | `Completed` is always `true` for completed runs and flows into the manifest. | Yes. |
| M-5 | `runner.go:16` sets `AnalyserVersion = "devel"`. | Yes. |
| M-6 | `redaction.go:18-23` has no patterns for `ghp_`, Basic auth, JWTs, URL userinfo or `pwd=`. Event names, scope names and bytes values are never scanned. | Yes. A missed secret leaves its record eligible and stores the secret in any retained representative. |
| M-7 | There is no internal deadline. Timing was not re-measured. | Yes. **OPEN QUESTION** on the magnitude. |

## 8. ADR-A challenge

Codex's rule: an active policy passes through any record whose equality-relevant fields were changed by lossy redaction. It may replace N records only when every destination-required semantic is equal, recoverable or explicitly outside the contract.

- **Lose evidence? Yes, through the representation gap (N-3).** ADR 0008 (line 27) allows the forwarding redaction policy to differ from the analysis policy. Two records that are equal under analysis redaction can differ in what the destination receives, so emitting one would misrepresent the other.
- **Attacker manipulation:**
  - Injecting secret-shaped text makes records ineligible. That fails safe, but it forces pass-through and therefore cost.
  - Avoiding the redaction triggers (the M-6 forms) keeps records eligible and leaves secrets in retained representatives.
- **Too weak:**
  - "Changed by a lossy redaction operation" needs a field-level definition that covers the body, kvlists and arrays.
  - It needs a reported counter, or it cannot be tested.
  - "Destination-relevant" leaves the policy author to decide what is relevant.
- **Timestamps:** first and last occurrence plus a count do not replace per-event timestamps. A policy that drops them is an aggregation contract, not exact equality.
- **Simpler rule:** for the first active policy, a record is eligible only if analysis redaction changed nothing and the analysis and forwarding representations are byte-identical.
- Keyed equality tokens remain an **OPEN QUESTION** that requires an ADR 0008 amendment.

## 9. ADR-B challenge

- **"Permanently advisory"** is defensible today, because the masking is regex-based and unevaluated on real corpora. However, the "operator-authored deterministic policy" escape hatch lets the same masking return if an operator copies C2's output into a policy.
- **Advisory output can still harm.** H5-A and H5-C show telemetry steering C2 counts. Candidate policies ranked by those counts become a persuasion channel aimed at the operator (threat-model-challenge §3).
- **Source templates:**
  - Agree they should be labelled `telemetry_asserted`.
  - Their state budget must be separate from inferred clustering (N-1).
  - Codex's "authenticated producer contract" has no record-level mechanism in OTLP. Transport authentication identifies the sender, not whether an attribute's value came from user input.

## 10. Corrected H-5 analysis

Code trace:

- `c2.go:277-286`: `sourceTemplate` returns the first non-empty string attribute keyed `log.record.template`.
- `c2.go:184-207`: the template is counted into `SourceTemplateRecords` and, per partition, into `SourceTemplateGroups`. Each new key costs `len(key)+80` bytes from the **shared** `result.StateBytes` (`:195-202`). Values up to `MaxTemplateBytes` (64 KiB) are accepted.
- `c2.go:226` and `:328`: the inferred cluster key is built from body tokens plus `attributeShape`, which skips `log.record.template`. The attribute never forms or merges an inferred group.
- `c2.go:238-241`: a new inferred cluster is excluded when `StateBytes+cost` would exceed the same shared budget.
- `c1.go:280`: all log attributes, including this one, are part of C1 identity.
- `redaction.go:18`: the key pattern does not match `template`, so the value is not redacted unless it happens to match a secret pattern.

Reproduced cases (default configuration, one resource, severity 9; temporary test, since deleted):

| Case | Input | Result |
| --- | --- | --- |
| H5-A | 100 records with body `heartbeat ok` and 100 distinct asserted templates | C1 duplicates 0; C2 groups 1; beyond-C1 99; source-template groups 100 |
| H5-B | 100 distinct bodies, one asserted template | source-template groups 1; inferred groups 100 |
| H5-C | 260 records with unique ~64 KB templates, then 2,000 with unique ~50-byte templates, then 50 distinct `payment failed kindN` events with no template | `state_cap_reached=true`; `excluded_state_cap=50` (all 50 genuine events excluded); inferred groups 1; beyond-C1 2,259 |
| H5-D (control) | the same 50 events on their own | 50 inferred groups; no exclusion |

**ORIGINAL H-5: PARTIALLY CONFIRMED.**

Exact threat:

- The attribute cannot choose or merge an inferred C2 group. On that point the original H-5 wording was wrong and Codex is right.
- Any log producer can set `log.record.template` to an arbitrary unredacted string of up to 64 KiB, which causes three effects:
  1. It inflates the source-template counts.
  2. It **exhausts the shared C2 state budget**, so later genuine events are excluded from inferred analysis (N-1, MEDIUM). This is visible in V0 but steerable by telemetry.
  3. It **splits C1 exact identity without affecting C2**, so repetition moves from the C1 `counterfactual` class to the C2 `inferred` class (N-2, MEDIUM).
- All three effects are non-destructive in V0. None is a BLOCKER.

## 11. ADR-C challenge

- **Agree** with a sequential baseline (source → C1 output → declared post-C1 transform) and with withholding C2 currency estimates.
- **Unit trap (FACT):** one report already holds C2 numbers in two different units (section 6, H-1).
- **Selective quoting:** a coverage vector is honest, but it can be quoted selectively, as with L-2.
- **Attacker inflation:** opportunity derived from redaction (H-2) or from asserted attributes (N-2) looks identical to genuine repetition.
- **Useful behaviour preserved:** removing priced C2 does not block prioritisation, provided unpriced C2 counts remain.
- **OPEN QUESTION:** whether the analysis window and the instance scope (a quiet week, one instance behind a load balancer) are representative.

## 12. ADR-D challenge

- **Held-window records (N-4):** the rule covers new records but not records already held at a deadline, cap, crash, OOM kill or bypass. The `log_dedup` precedent (OTV-05) loses an aggregate when a flush fails.
- **Double emission:** a bypass after a partial aggregate can deliver both the aggregate and its raw members.
- **Pass-through amplification:** an attacker who can trigger uncertainty conditions can force pass-through and a cost spike. This is the correct safety direction, but it must be visible.
- **"Fail closed at startup":** if PulseFold is inline, something must still forward when it declines to enter active mode.
- **cgroup limit:** an OOM kill of an inline component is a black hole unless the topology isolates it (ADR 0007).
- **Testability:** each of Codex's uncertainty conditions can be tested by fault injection.

## 13. ADR-E challenge

- **Build identity:** agree. Policy identity must also include:
  - both redaction policies (analysis and forwarding);
  - every cap and limit value, since these change results (M-3);
  - the billing-model commitment;
  - the sample-seed policy.
- **Bounded recovery:**
  - Incidents, security incidents especially, are often found after a short window has expired.
  - A bounded encrypted raw copy concentrates secrets (threat-model-challenge §4c), which puts P-007 on the critical path.
  - Codex's secondary-path option is the safest first experiment.
- **Different dependencies:** provenance/build identity and recovery depend on different things. Build identity is ready now; recovery needs P-007 and an archive design. They should not share one ADR.

## 14. Redaction/equality recommendation

**DESIGN DECISION (recommended).** Destructive equality invariant:

> A policy may replace N source records with fewer output records only if, for each of the N records, every field of the representation the destination would have received without PulseFold is either (a) byte-identical to the corresponding field of the emitted output, (b) retained losslessly in verified recoverable evidence for the declared recovery window, or (c) named explicitly in an operator-approved, versioned aggregation contract. The fields are resource, scope, schema URLs, severity, body, attributes, event name, flags, trace ID, span ID, timestamp and observed timestamp. Equality after analysis redaction, canonical-projection equality and semantic equivalence never satisfy this invariant.

Redaction eligibility:

- A record is ineligible for any destructive policy if any redaction operation changed any field of it, or if its analysis and forwarding representations differ.
- Ineligible counts are reported per field class.
- A production redaction policy that closes the M-6 gaps must exist before any representative is persisted.
- Keyed tokens stay deferred (an **OPEN QUESTION** that requires an ADR 0008 amendment).

## 15. C2 final role recommendation

**DESIGN DECISION (recommended):**

- C2 inference output is never an input to a destructive policy.
- Any structural or field-masking policy, however authored, must pass the same evaluation gates as a C2-based policy would. Its provenance records whether C2 suggested it.
- C2 never emits a machine-applicable policy artefact.
- Every advisory C2 figure carries its coverage and trust class.

## 16. Telemetry taint/provenance recommendation

**DESIGN DECISION (recommended):**

- Asserted templates are labelled `telemetry_asserted`, separately from PulseFold-derived inference, and have their own state budget or stage.
- Telemetry-asserted metadata never drives grouping, eligibility or ranking for a destructive decision.
- Partitions that carry asserted metadata are reported.
- The analysis interval carries a field-level `telemetry_derived` origin (L-3).
- **OPEN QUESTION:** whether any record-level producer authentication is achievable in OTLP.

## 17. Coverage model recommendation

**DESIGN DECISION (recommended):**

- Adopt Codex's coverage vector: observed, eligible, analysed, transformed, passed through by reason, and unknown or rejected. Every numerator declares its unit.
- Every opportunity figure is emitted in the same object as its source-record denominator and its weakest classification, so it cannot be quoted without them.
- The single `Completed` flag is replaced by separate execution outcome, coverage, eligibility, transformed and bypassed fields.
- Report the share of opportunity that comes from redaction-changed records and from asserted-metadata partitions.

## 18. Economics recommendation

**DESIGN DECISION (recommended):**

- Accounting is sequential and non-overlapping.
- Currency is shown only in the source unit the meter bills.
- No priced C2.
- A combined figure is allowed only when one versioned policy enables every component.
- Each figure states its window and instance scope and is never extrapolated.

## 19. Resource/fail-open recommendation

**DESIGN DECISION (recommended):**

- Admission control and deadlines are enforced in the application, with independent input, in-flight, state, queue, CPU and wall-time budgets. There is no in-process RSS promise.
- Pass-through on uncertainty also covers held records.
- Nothing is acknowledged upstream before it is forwarded or durably queued.
- A bypass flushes held members raw, with no double emission.
- Pass-through counts are reported by reason.
- Invalid configuration means PulseFold does not enter active mode, and forwarding continues.
- Active mode runs only in a topology whose kill -9 and OOM behaviour has been tested.

## 20. Representative/evidence invariants

- **Conflict (FACT):** docs/03 §5 (line 88), threat-model-challenge §3 and codex-architecture-review security item 5 all require selection that cannot be predicted from record position. Codex's "deterministic stratified samples" would be predictable to a telemetry producer.
- **DESIGN DECISION (recommended):**
  - Separate *coverage* from *choice*. The strata are deterministic: at least one per distinct value of each declared dimension and per state transition.
  - The member chosen within each stratum is selected unpredictably to producers, using a secret seed recorded in the manifest so the operator can reproduce the choice.
  - Strata are capped. Cap overflow means pass-through, never dropped strata.
- Also retain:
  - counts per time bucket (so a real event's time cannot be hidden among exact copies);
  - first and last occurrence;
  - resource and scope;
  - severity;
  - trace and span relationships;
  - every policy, build and redaction version.
- Rare, error, novel, tainted and redaction-changed records always pass through.

## 21. Recovery/rollback recommendation

**DESIGN DECISION (recommended):**

- Use a retention-bounded recovery window, not a permanent archive.
- Security and audit classes are never compacted by default. docs/03 invariant 5 currently says only "may be configured".
- The **required** first active topology is a secondary path, with the existing backend remaining authoritative.
- A bounded encrypted raw copy requires P-007 decisions first.
- When the window expires, reconstruction claims flip to unavailable through ADR 0009 tombstones.
- Rollback requires:
  - a global bypass;
  - per-policy disable;
  - atomic configuration rollback;
  - the held-window flush rule;
  - an append-only audit of policy changes (PR-016), independent of report signing.

## 22. Authentication recommendation

- The shadow report is the artefact used to justify *enabling* compaction (threat-model-challenge §3).
- A re-sealed report is indistinguishable from an original (V0 review, Observation).
- Enablement crosses a trust boundary whenever the report was produced elsewhere, for example in CI or by another party.
- Signing before PF-2001 adds cost without benefit while the machinery is disabled.
- **Dependency:** Alex may instead choose to require that enablement re-runs the analysis locally under the build-stamped binary and decides from that run only. That choice would move the conclusion to "deferred until external trust boundary". The safe default is the conclusion below.

**AUTHENTICATION REQUIRED BEFORE ACTIVE ENABLEMENT**

## 23. Evaluation gate

- **FACT:** no licensed corpus is in use (benchmark-corpus.md; docs/09:29). The synthetic labels were written by the same authors as the algorithm (`internal/c2/evaluation.go`).
- **DESIGN DECISION (recommended):** the preservation metrics and the labelled incident set are fixed in the evaluation plan *before* any policy is evaluated, and are owned independently of the policy author. Codex's "policy's defined preservation metrics" is circular.
- Required measures:
  - incident recall on that set;
  - unsafe-merge rate;
  - error and rare-event preservation;
  - redaction- and assertion-derived opportunity share;
  - pass-through by reason;
  - resource and timeout behaviour at the configured limits;
  - held-window crash recovery;
  - replay of the destination's own alert and query definitions.
- **OPEN QUESTIONS:** who labels incidents, and which corpus can legally be used. No thresholds are proposed here.

## 24. Safety requirements both reviews missed

1. **Analysis-versus-forwarding representation (N-3).**
2. **Held-window semantics on crash, deadline or bypass (N-4).**
3. **Bypass double emission** of an aggregate plus its raw members.
4. **Attacker-forced pass-through cost amplification.**
5. **Asserted templates starving C2 state (N-1)**, and **telemetry-controlled class shifting (N-2)**.
6. **Multi-instance consistency.** Per-instance active state behind a load balancer makes different decisions for identical records, and single-instance shadow analysis overstates fleet-wide repetition (log-compaction-research §1).
7. **Timestamp semantics of emitted aggregates** (OTV-05 precedent).
8. **Policy-change authority and audit**, independent of report signing (PR-016, insider threat).
9. **Environment or tenant boundary as a mandatory part of active identity** (threat-model-challenge §7; docs/03 invariant 4).

## 25. Phase 1.5 scope recommendation

Minimum task set:

1. Revise ADR-A to ADR-E using sections 14 to 22 and the disposition in section 26.
2. Decide the analysis-versus-forwarding representation question.
3. Write an active-path contract covering inputs and outputs, held-window semantics, bypass, the coverage vector and provenance fields.
4. Write the evaluation plan: corpus licensing, independent incident labelling, metrics fixed in advance and destination-query replay.
5. Produce a reproducible H-4/M-7 profile with a committed, repeatable command.
6. Record Alex's choice for enablement: authenticated report or local re-run.

Out of scope: any implementation of these mechanisms.

## 26. Five-ADR disposition

| ADR | Disposition | Reason |
| --- | --- | --- |
| ADR-A | ACCEPT WITH MODIFICATION | Right direction. Needs the section 14 invariant (destination representation, byte-level fields, named aggregation contract), a reported ineligibility counter and a production redaction prerequisite. |
| ADR-B | ACCEPT WITH MODIFICATION | Close the operator-policy loophole, separate the asserted-template state budget and treat advisory output as a labelled persuasion channel. |
| ADR-C | ACCEPT WITH MODIFICATION | Add single-unit currency, inseparable denominators and redaction- and assertion-derived attribution. |
| ADR-D | ACCEPT WITH MODIFICATION | Add held-window, acknowledgement, double-emission and forced-pass-through rules, and a tested-topology requirement. |
| ADR-E | SPLIT | Build/policy identity is needed before PF-2001 and is ready now. Recovery depends on P-007 and an archive design, and gates active enablement. |

## 27. PF-2001 implementation gate

PF-2001 may begin only when all of the following are true:

- the revised ADR-A to ADR-D and the identity half of ADR-E are accepted;
- the active-path contract and the evaluation plan are accepted;
- the representation question is decided;
- build-stamped analyser and policy identity is specified;
- the H-4/M-7 profile exists.

Scope even then: disabled-by-default policy machinery, with no forwarding integration and no persisted representatives until a production redaction policy exists.

## 28. Active-enablement gate

Active transformation may be enabled only when all of the following are true:

- the recovery half of ADR-E is accepted and implemented on the secondary-path topology;
- the P-007 key-management and retention decisions are made;
- a production redaction policy is in place;
- authenticated reports exist, or Alex has chosen the local re-run alternative (section 22);
- the section 23 evaluation passes on a licensed corpus with independently fixed metrics;
- the kill, OOM and held-window fault tests pass;
- policy-change audit exists;
- security and audit classes are excluded by default.

## 29. Unresolved questions

1. Can active equality be computed over the forwarding representation without violating ADR 0008, or does ADR 0008 need amending?
2. Is any record-level authentication of `log.record.template` achievable in OTLP?
3. What recovery duration per telemetry class is defensible, and who decides it?
4. Which licensed corpus, and which independent labelling process?
5. Authenticated report, or local re-run, for enablement?
6. What fraction of real C1 opportunity is derived from redaction or from asserted attributes?
7. What are the reproducible whole-process memory and CPU figures (H-4, M-7)?
8. What consistency model applies to active state across multiple instances?

## 30. Final recommendation

- Keep Phase 1 complete.
- Accept ADR-A to ADR-D with the modifications above, and split ADR-E.
- Run the six Phase 1.5 tasks (section 25).
- Do not authorise PF-2001 until the section 27 gate passes.
- Do not authorise active compaction until the section 28 gate passes.
- Next action: Alex decides the representation question (section 8) and the authentication-or-re-run question (section 22). The ADR drafts follow from those decisions.

Changes in this document: research documentation only. No production code, tests or other documents changed. One temporary H-5 test was run and deleted. Gates run: `git diff --check`, a full diff inspection, and a scan for credentials, private information and local paths.
