# Grok V0 Adversarial Safety Review

**Reviewer:** Grok (independent adversarial review, PULSEFOLD PROMPT 004)
**Reviewed commit:** `dc11cca` (`test: add V0 fault and performance harness`)
**Date:** 2026-10-13
**Scope:** Phase 1 V0 laboratory implementation, PF-1001 to PF-1008. This review does not start PF-2001 or any Phase 2 mechanism.

## 1. Executive conclusion

Facts:

- I found one **BLOCKER**, and it is fixed in this change. C2's "beyond C1" opportunity re-counted C1's exact-duplicate reduction. A replay of 100 identical lines reported `Candidate records beyond C1: 99` (`additional_counterfactual_records`), when the true structural opportunity beyond C1 was 0. This contradicted README.md:83 ("C2 is an inferred structural candidate beyond C1"). No existing test caught it.
- I found one closely related **HIGH** economics finding. The priced C2 "inferred potential" uses the same C1-weighted subset (`internal/accounting/accounting.go:285`). On a 1,000-record replay, C1 and C2 each showed USD 99.90 against a USD 100.00 baseline, so adding the two lines gives 199.8% of the baseline. I added a committed limitation saying the two must not be added. I left the underlying metric semantics alone because changing them is an economics decision.
- Redaction (`v1-fixture-safety`) replaces a **whole string** if any secret pattern matches (`internal/redaction/redaction.go:157-160`). It also replaces the value of any attribute whose key contains `token`, `secret`, `password` and similar (`:18`, `:128`). In both cases, 100 operationally distinct records collapsed into **one** C1 identity: a 99% "exact duplicate" rate created entirely by redaction. This matches the documented semantics ("redacted canonical equality"), and ADR 0008 acknowledges that redaction changes grouping. It is still the single most dangerous input for Phase 2 design.
- C2 merges operator-distinct records. Examples: `transaction:commit` with `transaction:rollback`, `session=expired` with `session=active`, `user=root` with `user=alice`, different IPs and ports, different Go file:line locations, and `/admin/0` with `/admin/1050`. C2 is non-destructive, so this is Phase 2 evidence rather than a V0 failure.
- Memory is **not** bounded by the configured state limits. A 49 MB input of tiny unique records, under the 64 MiB default input cap, peaked at about 1.39 GB RSS. The PF-1008 benchmarks stop at 4,096 records and do not exercise this.

Verdict: once the BLOCKER fix is in, the narrow claim "V0 laboratory implementation complete, Phase 2 not started" is justified. It is **not** justified as evidence for designing active compaction. I list 13 Phase 2 gates (section 15). PF-2001 should **not** be authorised until the proposed ADRs (section 16) are accepted.

## 2. Methodology and honesty about reading

- **Read in full:** AGENTS.md; docs/03_TELEMETRY_COMPACTION.md; docs/08_SECURITY_THREAT_MODEL.md; docs/18_NON_GOALS.md; docs/19_DECISION_LOG.md; ADRs 0004, 0006, 0007 and 0008; all production Go source under `internal/` and `cmd/` (redaction, c1, c2, input, runner, accounting, report, provenance; I read main.go diagnostics and the manifest at the level of their structs and functions).
- **Read in part or skimmed:** README; PROJECT_STATE; TASK_BOARD (the Phase 0 portion); docs/09 (sections 20-63); docs/10, 12 and 02 (keyword search for determinism, memory and coverage claims); ADR 0009 (via the PF-1006 work and the manifest code); the harness and the test function inventory.
- **Not read for this review:** docs/00, 01, 11 and 13 in full; ADRs 0001-0003, 0005, 0010 and 0011; most Phase 0 research. Parts of earlier sessions covered some of this material, but I did not re-read it here. Treat any finding that depends on those documents as unverified.
- I did not rely on the PF-1008 completion report.
- Experiments ran as temporary Go tests in `internal/runner` and as a built CLI on synthetic files. They were removed afterwards; the cases are reproduced below. Environment: box, linux/amd64, go1.27.0 (GOTOOLCHAIN=auto), 8 vCPU.

## 3. Findings by severity

### BLOCKER

**B-1 (FIXED): C2 "beyond C1" opportunity double-counted C1.**
- Where: `internal/c2/c2.go` (formerly `AdditionalOpportunity = CandidateRecords - InferredTemplateGroups`). `CandidateRecords` is weighted by C1 group counts.
- Smallest case: 100 copies of `"identical heartbeat"`. Before the fix, C1 reported `duplicates=99` and C2 reported `beyond C1=99`. After the fix, C2 reports `0`.
- Fix: the opportunity is now `RecordsAnalysed (representatives) - InferredTemplateGroups`. The C2 `CandidatePolicyVersion` moves to `c2-report-only-no-parameters-v2`, so reports produced under the old semantics can be told apart (the analyser version is still `devel`; see M-5).
- Regression test: `TestC2OpportunityDoesNotReclaimC1ExactRepetition`.

### HIGH (Phase 2 gates; they do not invalidate shadow V0)

- **H-1 Overlapping priced opportunities.** The billing C2 potential uses baseline = C1-weighted candidates and counterfactual = groups (`accounting.go:285`). It therefore contains the C1 reduction. A manager adding "C1 counterfactual opportunity 9990" and "C2 inferred potential 9990" gets 19,980 against a 10,000 baseline. Mitigated in this change by a committed limitation (`TestPricedC1AndC2OpportunitiesDeclareOverlap`). The metric semantics are left for a decision.
- **H-2 Whole-value redaction creates false C1 duplicates and erases incident-distinguishing context.**
  - Bodies: `payment failed for account N: token=sk_N amount=…` for N = 0..99 collapse from 100 distinct records to 1 identity (99 "exact duplicates").
  - Attribute keys `auth_token_status` and `max_tokens`, which are not secrets, have their values wiped. 100 distinct records become 1 identity.
  - The report has no count of records altered by redaction, so redaction-induced equality cannot be told apart from real repetition.
- **H-3 C2 masking crosses semantic boundaries.** Measured with 2-record pairs at the same severity and resource:

  | Pair | Groups | Masked by |
  |---|---|---|
  | `transaction:commit` / `transaction:rollback` | 1 | `idValue` |
  | `session=expired` / `session=active` | 1 | `idValue` |
  | `user=root` / `user=alice` | 1 | `idValue` |
  | `10.0.0.5` / `203.0.113.9` | 1 | `ipToken` |
  | `10.0.0.1:5432` / `:22` / `8.8.8.8:53` | 1 of 3 | `ipToken` (includes the port) |
  | `main.go:10` / `auth.go:999` / `billing/charge.go:42` | 1 of 3 | `goLocation` (whole path) |
  | `/admin/0` + 50 `/admin/10xx` | 1 of 51 | `urlID` |
  | `/facade12` / `/deadbeef` | 1 | `urlID` hex |
  | cert expiry `2026-…` / `1970-…` | 1 | `timeToken` |
  | nil UUID / real UUID | 1 | `uuidToken` |
  | HTTP 200 / 500, `SQLSTATE 40001` / `23505`, AZ `us-east-1a` / `eu-west-1b`, `HTTP/1.1` / `HTTP/2`, `50/100` / `100/100` | kept separate | — |

  Large-opportunity case: 10,000 `transaction:commit applied` plus 1 `transaction:rollback applied` form a single group, and the rare rollback is absorbed. `idValue` is at `c2.go:40`. The regex treats any `user|session|transaction…[=:]` value as an identifier, but those values are often state words.
- **H-4 Memory amplification beyond the configured bounds.** CLI peak RSS:

  | Input | Size | Peak RSS | Time |
  |---|---|---|---|
  | 1.45M tiny unique records | 51 MB | 1,386 MB | 6.8 s |
  | 20k records with depth-31 nested kvlists | 64 MB | 1,064 MB | 4.5 s |
  | 600k unique records | 34 MB | 807 MB | 4.4 s |
  | 200k records, one resource each | 31 MB | 526 MB | 2.5 s |

  The C1 (64 MiB) and C2 (16 MiB) state limits do not bound protojson decoding, the cloned OTLP messages from redaction or canonicalisation. Docs/02:60 lists "bounded memory" as an architecture property. For V0 this is acceptable *only* as an offline tool. DoS resistance is not demonstrated.
- **H-5 C2 accepts telemetry-chosen grouping hints.** Records carrying `log.record.template` are counted as source-template groups (`c2.go`, `sourceTemplate`). An attacker can make any records claim one template. This is only an advisory count in V0, but it must never drive Phase 2 decisions without trust.

### MEDIUM

- **M-1 C2 coverage is in representatives, not records.** 10,000 identical structured-body records plus 1 string record are reported as `Analysed representatives: 1 of 2 eligible`, while 10,000 of 10,001 records were excluded. The text renderer does not show the weighted exclusion count. The JSON does carry `excluded_records=10000`, and a limitation line appears.
- **M-2 There is no end-to-end coverage.** The C2 denominator is C1's representatives, so when C1 hits its state cap, C2 can read near-complete over a tiny subset. In `unique.json`, C1 analysed 100,000 of 600,000 and C2 analysed 10,000 of 100,000. The C1 limitation is present, but no single ratio says "C2 analysed 1.7% of input".
- **M-3 Capped C1 results depend on input order and favour first-seen identities.** With MaxGroups=3, 5 rare errors followed by 1,000 heartbeats gave `analysed=3, excluded=1002`. Reversed, the same input gave `analysed=1002, dup=999, excluded=3`. Without caps, the manifest ID is identical under reversal. Under caps, rare events arriving after the cap are the ones excluded. This is visible but biased.
- **M-4 `StageCoverage.Completed` is always `true` for completed runs** (`accounting.go:223-231`), even when the state cap was reached. The ratio and `state_cap_reached` carry the truth, but a reader of `completed: true` could be misled.
- **M-5 `AnalyserVersion` is the constant `"devel"`** (`internal/runner/runner.go:16`). Manifests cannot distinguish builds. B-1 shows why this matters: two builds with different semantics produced identical analyser versions. I mitigated this here by bumping the policy version; a build-stamped version is still needed.
- **M-6 Redaction misses common secret forms.** These passed through unchanged into the C1 canonical form and the hash bucket: `ghp_…`, `Authorization: Basic …`, JWTs, `postgres://admin:hunter2@…` and `pwd=…`. Bytes values, event names and scope names are never scanned. These values never reach reports in V0, but Phase 2 storage of canonical forms would persist them.
- **M-7 CPU cost of large bodies.** 1,000 bodies of 60 KB took 16 s, about 16 ms per record of regex work. There is no time cap, despite safety invariant 9 in docs/03 ("hard time, entry and byte caps").

### LOW

- **L-1** Benchmarks (`internal/harness/harness_test.go:185-220`) scale only to 4,096 records with the defaults. The 64 MiB default input cap allows about 1.4M records, so the cliffs in H-4 and M-7 are not covered.
- **L-2** The text renderer prints `Integrity: verified SHA-256 consistency check`. Followed by `Authentication: not provided`, that is accurate, but a future agent could quote only the first line.
- **L-3** Observed timestamps and timestamp 0 are ignored for the analysis interval. The interval is attacker-controlled (any `timeUnixNano`), yet it appears in the manifest without a telemetry-derived marker on the field itself.
- **L-4** `goLocation` masks only `.go` files, so stack locations in other languages stay literal. That is conservative, but it makes C2 behaviour language-dependent.

### OBSERVATION

- The integrity check worked as designed. Editing `C1.DuplicateRecords` to 999999 and deleting all limitations was detected. Calling `report.IntegrityCommitment` and `provenance.AttachIntegrity` then re-sealed the report, and `VerifyIntegrity` returned nil while the text renderer printed "Integrity: verified". `created_at` moved by 5 years still verifies (it is excluded by design). This is the documented "not authentication" behaviour (report.go:300 and onwards; README:83). It does mean that **"integrity verified" must never be read as "evidence trusted"**.
- Telemetry taint: ESC/OSC terminal sequences, "IGNORE PREVIOUS INSTRUCTIONS", a fake `git push --force`, `/Users/…/.ssh/id_rsa` and `https://u:p@x` injected in bodies and attributes appeared in **neither** the text nor the JSON report. Telemetry can influence only counts, the interval and exclusion reasons, never free text. Source labels are allow-listed (`input/replay.go`, `SafeSourceLabel`).
- C1 canonicalisation keeps type distinctions: zigzag for ints, `Float64bits` for doubles (so -0 ≠ 0 and NaN payloads differ), length-prefixed strings and bytes, sorted attributes and nil-vs-empty markers. It excludes both timestamps by design. I found no collision between operationally distinct, *unredacted* records. Duplicate attribute keys are kept as separate sorted entries.
- C2 keeps HTTP status codes, SQLSTATE codes, AZs, protocol versions and numeric ratios literal, as documented.

## 4. Reproducible cases

All cases use one resource (`service.name=svc`), scope `s`, severity 9 unless stated, and the default configuration through `runner.Analyse`.

1. **B-1:** 100 identical bodies. Expect C1 duplicates = 99 and C2 beyond-C1 = 0 (permanent test).
2. **H-1:** 1,000 identical bodies; model: accepted_records, fixed 10 minor units per record. Baseline 10,000; C1 9,990; C2 9,990 (permanent test).
3. **H-2:** `fmt.Sprintf("payment failed for account %d: token=sk_%d amount=%d", i, i, 7*i)` for i < 100, severity 17. Result: C1 unique = 1.
4. **H-3:** the pairs in the table above, severity 13; count `inferred_template_groups`.
5. **H-4:** `{"resourceLogs":[{"scopeLogs":[{"logRecords":[{"body":{"stringValue":"e<i>"}} …1.45M]}]}]}` (compact JSON, 51 MB) run with `pulsefold analyse f.json --format json`; read the peak RSS.
6. **Forgery:** take a complete report, change any counter, then call `IntegrityCommitment` and `AttachIntegrity`. `VerifyIntegrity` returns nil.

## 5. C1 assessment

C1 is exact over the **redacted** canonical form, and its hash is used only for bucketing, with full byte equality checked afterwards. As a description of the code, "exact repetition" is correct. As a description of the telemetry, it is not: redaction is lossy and many-to-one (H-2). Timestamps are excluded, so C1 "duplicates" are events that recur within the window, not replays. Trace and span IDs are included, so per-request logs rarely deduplicate, which is good. Verdict: **correct for its documented definition; unsafe as a basis for dropping records until redaction-induced equality is measurable and avoidable.**

## 6. C2 assessment

C2 is deterministic and order-independent without caps (exact template-key clusters, no Drain similarity merge). Partitioning by resource, scope and severity holds. The masking regexes overreach (H-3), and attacker-controlled prefixes and suffixes can choose their group (H-5). The synthetic evaluation in `internal/c2/evaluation.go` measures only fixture labels that the authors chose. Verdict: **acceptable as an inferred, non-destructive indicator. Not safe for any destructive decision.**

## 7. Redaction interaction

The pipeline order is input → redaction → C1 → C2, and it is enforced by types: `redaction.Batch` has private records. The interaction is many-to-one. One matching substring blanks the whole body, so novelty, account IDs and amounts disappear, and C1 and C2 both report repetition that does not exist in the source. Severity and resource are not redacted, so partitions survive. Phase 2 needs a design decision among the following options. I implemented none of them.

- (a) Keyed equality tokens, such as HMAC with a per-deployment key, so distinct secrets remain distinct without being revealed.
- (b) Span-level rather than whole-value redaction.
- (c) Field-specific semantics, with an allow-list instead of the substring key `token`.
- (d) Ephemeral pre-redaction equality checks that never persist raw values.
- (e) A report counter of records and attributes changed by redaction.

## 8. Coverage assessment

I found no unknown→0 or unknown→100% substitution. Zero eligible records gives `known:false` with a reason. An unknown observed count gives "denominator unknown". The byte limit fails the whole input instead of analysing a prefix. Cancellation produces a failure report without partial counts. The weaknesses are M-1, M-2 and M-4: the ratios are per stage and C2's are unweighted.

## 9. Economics assessment

The arithmetic is sound: exact integers, ceiling on indexed fractions, overflow reported as unavailable, monotonic tier validation, and zero rates allowed. There are no commitments, minimums, fixed fees or retention modelling. Canonical bytes are not source or wire bytes, and the docs say so. The terminology "gross observability cost opportunity" is conservative. The misreading risk was H-1 (now carries a limitation) and B-1 (fixed). An engineering manager would still read "C1 counterfactual opportunity: USD 99.90 of USD 100.00" on a heartbeat-only replay as near-total saving. Under H-2, much of that can be redaction artefact.

## 10. Provenance and integrity assessment

The commitment covers report content and the deterministic manifest fields. It excludes the integrity block, `manifest_id` and `created_at`; lifecycle records live outside it. SHA-256 is change detection only, and anyone can recompute it (Observation). The docs state this clearly in README:83 and in the threat model under "Unsafe report interpretation". I found no mutation of a *committed* field that passed without re-sealing. Gaps: the `devel` analyser version (M-5), and the fact that a re-sealed report is indistinguishable from an original. Verdict: **fit for accidental-change detection only.**

## 11. Privacy and taint assessment

Reports contain only counts, versions, references and fixed strings. Attacker text did not reach the outputs in any test. CLI errors are fixed strings. Source labels are allow-listed and paths are reduced to their base name. Telemetry can still influence the analysis interval (L-3), all counts and therefore every opportunity figure, and C2 source-template counts (H-5). Verdict: **no content leakage found. The numeric outputs remain attacker-influenceable by construction.**

## 12. Resource assessment

The state caps hold and are visible (`state_cap_reached`). Total process memory is about 20-27× the input size in the worst cases measured (H-4). CPU on large bodies is about 16 ms per 60 KB record (M-7). The C1 identity-size check happens after the full canonical form is built (`c1.go:148`); the input byte cap bounds it, but it costs memory. Verdict: **bounded per stage, not bounded per process. No DoS-resistance claim is supported.**

## 13. Reproducibility assessment

The same input, configuration and binary give an identical manifest ID; `created_at` is the only volatile field. Reversing the input order without caps gives the same manifest ID; I measured this. Group ordinals and C1 representative indices depend on order but are not serialised. Under caps, the results depend on order (M-3). The docs claim determinism "for identical input, configuration and analyser version" (docs/10:111). That is accurate, but "analyser version" is the constant `devel`, so the claim is weaker than it reads. Verdict: **accurate but under-specified.**

## 14. Benchmark assessment

The PF-1008 matrix runs the full pipeline (decode, redaction, C1, C2, accounting, manifest) on synthetic fixtures. Setup is generated outside the timed loop and allocations are reported with `-benchmem`. The workloads are small and favourable: at most 4,096 records, 256 KiB large bodies and few partitions. They do not reach the configured limits where the H-4 and M-7 cliffs appear. They show relative cost and the absence of regressions at small scale only. They do not show throughput or memory at the configured limits, under production distributions or for live forwarding.

## 15. Phase 2 safety gates (mandatory before active compaction)

1. **Safe to compact** means:
   - the transformation is reversible, or the dropped information is provably reconstructable from the retained representatives plus metadata, *in source terms, not redacted terms*;
   - the incident-replay metrics in gate 13 show no loss.
2. **C1:** only exact duplicates in the pre-redaction source (or under keyed equality tokens) may be collapsed, and the count and first/last timestamps must be kept. Redaction-induced equality must never count as duplication (H-2).
3. **C2:** no destructive decision may rely on C2 until masking is field-typed, operator-overridable and validated on real labelled corpora. C2 may at most suggest candidates for human-approved policy (H-3, H-5).
4. **Representative evidence:** the first and last event per group, randomised reservoir samples (docs/03 §5), and one example per distinct value of the resource, severity, status and error-code dimensions.
5. **Metadata that must survive:** resource and scope identity, severity, trace and span IDs (or their cardinality plus samples), the time range, counts, and the policy, analyser (build-stamped) and redaction versions.
6. **Reconstructable:** per-group counts over time, distinct-value sets or sketches for masked fields, and a pointer to raw or cold evidence.
7. **Uncertainty:** forward unchanged. Uncertainty includes coverage that is not known or complete, a cap that was reached, or any masking ambiguity.
8. **State limits:** fail open to pass-through and report it. Never drop records because state is full, and never favour first-seen identities.
9. **New, rare or error events:** always pass through. Events at or above a severity threshold, novel groups, and groups below a rarity threshold are never compacted.
10. **Rollback and bypass:** a per-policy kill switch, a global bypass to pass-through, and a tested failure-domain separation (ADR 0007).
11. **Raw or cold evidence:** a retention-bounded raw archive or replica must exist and be verified before any active transformation. ADR 0009's lifecycle needs to become a real store first.
12. **Privacy and keys:** a keyed equality-token design (key generation, rotation, scope per tenant or deployment, no keys in reports), plus a redaction policy beyond fixture safety (M-6).
13. **Metrics:** recall of incident-relevant events on replayed labelled incidents (target 100% for error and rare events); unsafe-merge rate on real corpora; the share of opportunity attributable to redaction; per-process memory and CPU at the configured limits; and bypass-trigger counts.

## 16. Recommended ADRs before PF-2001

1. **Equality under redaction:** keyed tokens versus span redaction versus pre-redaction ephemeral comparison (gates 2 and 12).
2. **C2 trust boundary:** C2 is advisory only, and operator-declared field semantics are required before any C2-driven action (gates 3 and 9).
3. **Opportunity accounting semantics:** non-overlapping, incremental C1 → C2 figures, end-to-end weighted coverage, and the billing C2 baseline (H-1, M-1, M-2).
4. **Fail-open state-limit behaviour and process resource budgets**, including a time cap and a whole-process memory bound (H-4, M-7, gate 8).
5. **Build-stamped analyser version and authenticated provenance** (signing or attestation) before evidence is relied on across trust boundaries (M-5, Observation).
6. **Raw and cold evidence prerequisite** for active transformation (gate 11; extends ADR 0009).

## 17. Unresolved questions

- What fraction of real-world C1 opportunity is redaction-induced? This needs a licensed public corpus, which docs/09 notes is still pending licence review.
- Should the billing C2 line remain cumulative (C1 plus C2) and be relabelled, or become incremental? That is a product decision.
- Is a whole-process memory bound a V0 requirement, or only a Phase 2 gateway requirement?
- I did not re-verify the ADRs, docs and research I did not read in full (section 2).

## 18. Changes in this commit and why the tests are permanent

- `internal/c2/c2.go`: B-1 fix, and the C2 candidate policy version bumped to v2.
- `internal/report/report.go`: a committed overlap limitation when both C1 and C2 are priced (H-1).
- `internal/runner/adversarial_test.go`:
  - `TestC2OpportunityDoesNotReclaimC1ExactRepetition` protects the stage-incrementality invariant. Without it, the "beyond C1" claim silently regresses, and no existing test noticed the original defect.
  - `TestPricedC1AndC2OpportunitiesDeclareOverlap` protects the committed limitation that prevents summing overlapping money figures. It skips itself if the semantics later become non-overlapping.
  - Both tests fail on `dc11cca` and pass on this change.
- Scratch experiments (H-2 to H-5, M-1 to M-7, forgery, taint) are summarised above rather than committed. They characterise behaviour that Phase 2 must change, so freezing them as tests would lock in the defects.
