# Notification and Routing

## 1. Goal

Route useful incident information to humans and agents without coupling PulseFold to one collaboration product.

## 2. Event types

Initial event catalogue:

- incident.opened
- incident.updated
- incident.severity_changed
- investigation.started
- investigation.completed
- recommendation.created
- incident.recovered
- incident.resolved

## 3. Destinations

Initial priority:
1. generic webhook
2. Slack

Later candidates:
- Microsoft Teams
- email
- PagerDuty
- Opsgenie
- Discord
- message queues
- custom agent endpoints

## 4. Notification quality

A useful notification contains:
- incident ID
- what changed
- impact if known
- current severity
- evidence summary
- investigation status
- link/command for more detail

Avoid:
- raw telemetry floods
- unsupported AI conclusions
- duplicate messages for the same incident
- constant edits with no material change

## 5. Progressive updates

Example sequence:

```text
OPENED
Checkout error rate rose from 0.4% to 12.9%.
PulseFold has opened PF-127 and started investigation.

UPDATED
A deployment occurred 96 seconds before failures began.
96% of failures are on revision 231.

INVESTIGATION
Probable cause: connection pool regression.
Confidence: 0.89.
Evidence: E14, E22, E29.

RESOLVED
Failure rate returned to baseline after revision 230 became active.
```

## 6. Routing policy

Possible routing dimensions:
- environment
- service
- severity
- team
- incident type
- business hours
- destination health
