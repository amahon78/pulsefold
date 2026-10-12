# Agent and Model Integration

## 1. Two distinct concepts

PulseFold must distinguish:

### Model provider
A model endpoint used by PulseFold's built in investigator.

Examples:
- OpenAI
- Anthropic
- xAI
- Google
- local/OpenAI compatible endpoint

### External agent
An existing organisational bot or agent that PulseFold proactively informs.

Examples:
- internal Grok bot
- Claude based SRE agent
- custom incident bot
- workflow platform

These are not the same abstraction.

## 2. Built in model provider contract

Conceptual interface:

```text
Generate(
  system_policy,
  structured_context,
  tool_contracts,
  budget
) -> model_response
```

The provider layer handles:
- authentication
- endpoint
- model identifier
- timeout
- retry
- token/accounting metadata

The core investigation layer handles:
- evidence
- tool permissions
- claim validation
- state

## 3. External agent event contract

PulseFold should emit a stable structured package.

Example:

```json
{
  "event": "pulsefold.incident.investigation_requested",
  "incident_id": "pf_...",
  "severity": "high",
  "summary": "...",
  "signals": ["..."],
  "evidence": ["pfe_1", "pfe_2"],
  "tool_endpoint": "...",
  "callback": "..."
}
```

The external agent may then query safe PulseFold tools.

## 4. Agent tools

Read only initial tools:

- get_incident
- list_incident_signals
- get_evidence
- search_logs
- get_metric_window
- get_trace
- list_recent_deployments
- get_deployment
- get_code_change_metadata
- find_similar_incidents
- get_runbook

Every tool should:
- scope access
- limit result volume
- record invocation
- return evidence IDs
- distinguish unavailable from empty

## 5. MCP

MCP is a strong candidate for exposing investigation tools but must not be the only integration surface.

HTTP/webhook contracts should remain available.

## 6. Context minimisation

Do not send every log to a model.

Investigation should:
1. start with incident summary and evidence index
2. let tools retrieve bounded detail
3. expand only where required

## 7. Model independence test

A provider adapter passes if:
- core tests run unchanged
- evidence semantics remain unchanged
- incident state remains unchanged
- model specific features are optional
