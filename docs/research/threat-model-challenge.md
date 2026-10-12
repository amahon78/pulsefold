# PF-0005: Threat Model Challenge (review of docs/08_SECURITY_THREAT_MODEL.md)

**Status:** Draft (Grok, 2026-10-12). Does not modify 08; proposes additions.

## 1. Overall verdict

08 is directionally right but **under-specifies the most likely real attack** (telemetry-borne reward hacking of the investigator), **misses compaction-specific integrity attacks**, and treats several mitigations ("telemetry is data, never instruction") as if they were enforceable controls. They are aspirations; LLMs cannot reliably separate data from instructions.

## 2. Telemetry / prompt injection

- FACT: "When AIOps Become 'AI Oops'" (Pasquini et al., USENIX Security 2026) demonstrates **AIOpsDoom**: attackers fuzz public endpoints to make services log attacker-controlled strings; the payload is a plausible *false root cause* plus a matching *remediation* ("adversarial reward hacking"), steering LLM agents to unsafe actions (e.g. adding a malicious package repo). They propose **AIOpsShield** — taint tracking of untrusted telemetry fields and sanitisation before agent consumption [1].
- Implications:
  1. The attack doesn't need "ignore all instructions" text; it looks like a normal error message. 08's example and "injection test corpus" understate this. Keyword filters won't catch it.
  2. **Read-only does not neutralise it.** A read-only agent that *recommends* the attacker's remediation to a human, or posts it to Slack with PulseFold's credibility, is a successful attack. ADR 0003 reduces blast radius but not the harm path through humans.
  3. **Compaction amplifies or hides it**: an attacker who spams a template can make their payload the "representative sample" (first/last/reservoir), or make a real error look like the ten-thousandth instance of a benign template.
- Proposed additions:
  - Taint labels on every evidence field: `origin = attacker-influenceable | system-generated | operator`. HTTP paths, user agents, request bodies, exception messages from user input = tainted. Render tainted text to models inside explicit delimiters and **forbid tainted text from appearing in recommendations verbatim** (URLs, commands, package names, hostnames).
  - Recommendations containing commands/URLs/repos must be validated against an operator allowlist or flagged "unverified, sourced from telemetry".
  - Eval suite must include reward-hacking payloads modelled on AIOpsDoom, not only instruction overrides.

## 3. Compaction integrity (missing from 08)

| Threat | Description | Mitigation proposal |
|---|---|---|
| **Evidence laundering** | Attacker crafts events matching a high-volume benign template so their malicious line is aggregated and only counted | Never compact security/audit categories (exists, invariant 5) **by default**, not optional; keep raw for tainted-origin + error severity; sample selection must not be attacker-predictable (seeded random, not first/last only) |
| **Template poisoning** | Flood of crafted lines shifts Drain clusters so later real errors merge into wrong template / wildcard swallows severity or error code | Severity + status code in identity; cap wildcard generalisation for numeric error codes; monitor template merge events |
| **Memory DoS** | High-cardinality junk exhausts dedup maps, Drain clusters (`max_clusters` default unlimited [2]), HLL trackers, `metadata_cardinality_limit` default unlimited in logdedup [3] | Hard caps on every stateful structure with overflow → **pass-through**, not drop; counters exported |
| **Savings-report manipulation** | Shadow report used to justify enabling compaction; poisoned input inflates "safe" estimates | Report must show confidence & adversarial-volume share; signed/versioned report |
| **Hash collision** | Covered by invariant 1–2 | Full-compare after hash |

## 4. Secrets and PII

- FACT: OTel `redaction` processor exists, allowlist-based and fail-closed by default on keys, plus blocked-value regex [4]. Its own README says it is "one line of defence".
- Gaps in 08: (a) secrets in **log bodies** (not attributes) — regex-only; (b) secrets **persist in compaction state** (dedup keys, Drain templates — a token that never varies becomes part of the *template string* and is then shipped as `log.record.template` and stored in the evidence layer); (c) **raw evidence archive** concentrates secrets in one place with long retention; (d) redaction ordering — must run **before** fingerprinting, or the template/hashes leak secrets (03 §8.8 leaves ordering open).
- Proposals: redaction before fingerprint (state as invariant); entropy+pattern detectors (gitleaks rule set as reference) on bodies before model boundary; evidence archive encrypted at rest with retention TTL; never send templates to models without redaction; PII: email/IP/phone detectors and GDPR right-to-erasure implications for an immutable evidence store (**immutability in 02 §6 conflicts with erasure obligations**).

## 5. SSRF / egress

- Add: webhook/notification destinations are operator config — validate at config time, block link-local/metadata IPs (169.254.169.254, fd00:ec2::254), resolve-then-connect pinning (DNS rebinding).
- External-agent contract includes `tool_endpoint` and `callback` (06 §3): callbacks must be pre-registered, signed (HMAC with rotation), and never derived from telemetry or model output.
- Archive refs (`raw_reference`) must be opaque IDs, not URLs.

## 6. Tool abuse & credential scope

- Read-only ≠ low-risk: `search_logs`, `get_code_change_metadata` can exfiltrate secrets *to the model provider*. Data egress to a third-party model is itself the risk. Need per-tool **sensitivity ceiling** and per-provider data-classification policy (e.g. "external provider: no raw bodies").
- External agents calling PulseFold tools need scoped, short-lived tokens bound to `incident_id` (capability tokens), not a global API key.
- PulseFold's own credentials (backend export keys, Slack, Git, cloud APIs) — the control plane should hold only what it needs; separate the exporter credentials (telemetry plane) from investigation credentials (control plane) into different processes.
- MCP: treat MCP server tool descriptions as untrusted if PulseFold *consumes* MCP tools; if it *exposes* MCP, audit every call (06 says this) and rate-limit.

## 7. Multi-tenant

08 defers multi-tenancy. Challenge: even single-org deployments are **multi-team / multi-environment**. Prod and staging evidence, or two teams' logs, in one evidence store with one agent context is a confidentiality problem *now*. Proposal: environment/tenant key as a mandatory partition in dedup identity (logdedup's `metadata_keys` is precedent [3]), evidence store, and agent scopes from V0 — "tenant = 1" by default, but the field exists.

## 8. Other gaps

- **Supply chain:** a Collector distro pulls hundreds of modules; need SBOM, `govulncheck`, signed images (cosign) — not only "dependency scan".
- **Report/notification integrity:** Slack messages are a persuasion channel; include evidence links and a clear "AI-generated hypothesis" label.
- **Audit log tampering:** append-only, hash-chained.
- **Model provider compromise/logging:** providers may retain prompts; per-provider retention settings must be documented.

## Sources
1. Pasquini et al., USENIX Security 2026, https://www.usenix.org/conference/usenixsecurity26/presentation/pasquini ; https://arxiv.org/abs/2508.06394
2. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/drainprocessor
3. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor
4. https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/redactionprocessor
