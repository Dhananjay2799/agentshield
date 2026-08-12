\# AgentShield



Zero-Trust Security Control Plane for Autonomous AI Agents.



AgentShield governs autonomous AI agent access to sensitive tools and infrastructure using short-lived agent sessions, contextual risk evaluation, OPA/Rego policy-as-code, human approval workflows, temporary authorization grants, and permanent audit trails.



\## Current Capabilities



\- Agent registration

\- Short-lived agent sessions

\- Session revocation and expiration

\- Protected API enforcement

\- Contextual risk scoring

\- OPA/Rego policy evaluation

\- ALLOW / DENY / REQUIRE\_APPROVAL decisions

\- Human approval workflow

\- One-time temporary authorization grants

\- PostgreSQL audit trail

\- Rego policy tests

\- Go formatting, testing, and vet checks

\- Docker Compose infrastructure



\## Current Architecture



```text

AI Agent

&#x20;  ↓

AgentShield Go Gateway

&#x20;  ↓

Session Validation

&#x20;  ↓

Agent Identity

&#x20;  ↓

Risk Engine

&#x20;  ↓

OPA / Rego

&#x20;  ↓

ALLOW | DENY | REQUIRE\_APPROVAL

&#x20;  ↓

Human Approval / JIT Grant

&#x20;  ↓

Audit Trail

&#x20;  ↓

PostgreSQL

