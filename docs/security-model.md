# AgentShield Security Model

## Overview

AgentShield applies zero-trust security principles to autonomous AI agents.

Autonomous agents are treated as untrusted computational identities. An agent does not receive permanent authority simply because it has been registered or authenticated. Every privileged operation must be evaluated using current identity, session, authorization, policy, risk, and behavioral context.

The security model follows the principle:

> Never trust an agent implicitly. Verify identity, authorization, context, and risk continuously.

---

## Security Objectives

AgentShield is designed around the following security objectives:

1. Prevent unauthorized agent actions.
2. Minimize the lifetime and scope of agent credentials.
3. Require explicit authorization for sensitive operations.
4. Introduce human approval for high-impact actions.
5. Detect abnormal or compromised agent behavior.
6. Contain compromised agents quickly.
7. Preserve an auditable record of security decisions.
8. Prevent failures in analytics infrastructure from bypassing authoritative access controls.
9. Apply least privilege to service-to-service communication.
10. Make security decisions observable and operationally measurable.

---

## Threat Model

AgentShield assumes autonomous agents may become unsafe because of:

- prompt injection;
- malicious tool output;
- compromised models;
- stolen agent sessions;
- credential leakage;
- software vulnerabilities;
- incorrect agent reasoning;
- excessive permissions;
- malicious operators;
- supply-chain compromise;
- compromised dependencies;
- abnormal automated behavior.

Therefore, successful authentication alone is not considered sufficient authorization.

---

## Protected Assets

AgentShield protects several classes of assets.

### Infrastructure Resources

Examples include:

- production services;
- databases;
- Kubernetes workloads;
- cloud infrastructure;
- deployment systems;
- internal APIs.

### Credentials

Examples include:

- API credentials;
- temporary access tokens;
- service credentials;
- authorization grants.

### Security State

Examples include:

- agent identities;
- active sessions;
- policy definitions;
- approval decisions;
- incidents;
- containment status;
- behavioral baselines.

### Audit Evidence

Security events and audit records provide evidence of agent activity and security-control decisions.

---

## Trust Boundaries

AgentShield contains multiple explicit trust boundaries.

```text
Autonomous Agent
      |
      | Untrusted request
      v
+----------------------+
| AgentShield Gateway  |
+----------------------+
      |
      +----> Session validation
      |
      +----> Risk evaluation
      |
      +----> OPA policy decision
      |
      +----> Approval workflow
      |
      +----> Credential broker
      |
      +----> Audit/event pipeline
```

The Gateway acts as the primary enforcement boundary.

Agents cannot directly bypass Gateway authorization and obtain trusted access through the AgentShield control plane.

---

## Agent Identity

Each autonomous agent receives an AgentShield identity.

An agent identity contains information such as:

```text
agent_id
name
agent_type
owner
framework
model
environment
status
```

Agent identity and agent authorization are deliberately separate concepts.

An existing agent identity does not automatically imply permission to perform an action.

---

## Short-Lived Sessions

Agents operate through short-lived sessions.

A session establishes temporary execution context and includes properties such as:

```text
session_id
agent_id
task_id
status
started_at
expires_at
ended_at
```

Sessions can become:

```text
active
expired
revoked
```

Sensitive API operations require a valid AgentShield session.

Expired or revoked sessions are rejected.

---

## Continuous Authorization

AgentShield evaluates actions individually rather than granting permanent authorization after login.

Conceptually:

```text
Agent
  |
  v
Session Validation
  |
  v
Agent Status
  |
  v
Risk Evaluation
  |
  v
OPA / Rego Policy
  |
  v
ALLOW / DENY / REQUIRE_APPROVAL
```

This limits the impact of a session or agent becoming compromised after initial authentication.

---

## Policy Enforcement

Authorization policy is implemented using OPA and Rego.

Policies can evaluate contextual attributes such as:

- action;
- resource;
- environment;
- agent type;
- agent identity;
- risk level;
- authorization state.

The policy engine produces one of three primary outcomes:

```text
ALLOW
DENY
REQUIRE_APPROVAL
```

Policy-as-code allows authorization rules to be versioned, reviewed, tested, and deployed alongside application infrastructure.

---

## Human Approval

High-impact actions may require human authorization.

When policy returns:

```text
REQUIRE_APPROVAL
```

AgentShield creates an approval workflow instead of immediately executing the operation.

An authorized human can approve or reject the request.

Approved operations may receive a narrowly scoped temporary authorization grant.

This creates a security boundary between autonomous reasoning and sensitive infrastructure changes.

---

## Just-In-Time Authorization

AgentShield avoids permanent broad privileges where possible.

Temporary grants are intended to be:

- short lived;
- narrowly scoped;
- revocable;
- auditable;
- usable only for the approved operation.

This reduces the blast radius of credential or session compromise.

---

## Credential Security

The Credential Broker separates authorization decisions from credential issuance.

Conceptually:

```text
Agent
  |
  v
Gateway Authorization
  |
  v
Approved Request
  |
  v
Credential Broker
  |
  v
Short-Lived Credential
```

Credentials should only be issued after successful authorization.

The agent should never receive long-lived infrastructure secrets merely because it is registered with AgentShield.

---

## Behavioral Detection

AgentShield continuously observes security events produced by agent activity.

Detection evaluates behavioral signals including:

- action frequency;
- denied operations;
- risk levels;
- action diversity;
- resource diversity;
- behavioral bursts;
- historical agent behavior.

Behavioral detection operates asynchronously from authoritative authorization.

This ensures detection failures cannot directly grant access.

---

## ML Anomaly Detection

AgentShield can evaluate behavioral feature vectors using the ML anomaly service.

The detection pipeline maintains per-agent behavioral baselines and sends eligible observations to the anomaly model.

Important security property:

> ML inference is advisory detection logic, not the authoritative authorization boundary.

If ML inference becomes unavailable, AgentShield records the failure while the Gateway, session validation, OPA policies, and deterministic detection continue operating.

This prevents model availability from becoming a single point of failure for access control.

---

## Agent Containment

When an agent is considered compromised or unsafe, AgentShield can contain it.

Containment performs authoritative state changes including:

1. Suspend the agent.
2. Revoke active sessions.
3. Revoke active temporary grants.
4. Persist containment audit evidence.
5. Attempt to publish the containment security event.

The critical security state is stored transactionally in PostgreSQL.

---

## Containment Invariant

AgentShield enforces the following invariant:

```text
agent.status != active
        =>
new sessions cannot be created
```

Existing sessions revoked during containment cannot continue authorizing operations.

Therefore:

```text
Contain Agent
     |
     +--> Agent suspended
     |
     +--> Sessions revoked
     |
     +--> Grants revoked
     |
     +--> New sessions blocked
     |
     +--> Revoked sessions rejected
```

This provides immediate control-plane isolation of the contained identity.

---

## Dependency Failure Isolation

Security-critical state transitions must not depend on asynchronous telemetry infrastructure.

For example:

```text
Containment
   |
   +--> PostgreSQL state transition
   |
   +--> Audit persistence
   |
   +--> Best-effort Kafka publication
```

If Redpanda/Kafka is unavailable, containment remains effective.

Similarly, ML inference failure does not bypass deterministic authorization.

This separation protects security enforcement from telemetry outages.

---

## OPA Failure Behavior

OPA is part of the authoritative authorization path.

If OPA becomes unavailable, AgentShield does not silently fall back to ALLOW.

Observed failure behavior:

```text
OPA unavailable
      |
      v
Policy evaluation fails
      |
      v
HTTP 503
      |
      v
No authorization bypass
```

This is a deliberate fail-closed boundary.

---

## ML Failure Behavior

The ML service is a supporting detection dependency rather than an authoritative authorization dependency.

Observed failure behavior:

```text
ML service unavailable
      |
      v
Inference request fails
      |
      v
ML failure metric increments
      |
      v
Adaptive and deterministic detection continue
      |
      v
Detection service remains healthy
```

This is a deliberate fail-soft boundary.

---

## Kafka Failure Handling

Security events are published to Redpanda using bounded request timeouts.

Kafka publication failures are:

- logged;
- counted in operational metrics;
- prevented from blocking indefinitely;
- prevented from rolling back completed containment.

Current bounded publish windows:

```text
normal security event: approximately 1 second
containment event: approximately 500 milliseconds
```

During failure-injection testing, emergency containment latency improved from approximately:

```text
7.25 seconds
```

to approximately:

```text
0.61 seconds
```

while the agent was still suspended, its active session revoked, and the permanent audit event persisted.

---

## Auditability

AgentShield records security-relevant operations in PostgreSQL.

Examples include:

- action evaluations;
- approval decisions;
- credential activity;
- incidents;
- containment operations.

Audit records allow operators to reconstruct security decisions and agent behavior after an incident.

---

## Network Isolation

Kubernetes NetworkPolicies implement service-level communication restrictions.

The cluster follows a default-deny approach with explicit communication paths for required dependencies.

Examples include:

```text
Gateway -> PostgreSQL
Gateway -> OPA
Gateway -> Redpanda

Detection -> PostgreSQL
Detection -> Redpanda
Detection -> ML Anomaly

Prometheus -> monitored services

Grafana -> Prometheus

Alertmanager -> Alert Receiver
```

This limits unnecessary east-west communication between workloads.

---

## Observability

Security controls expose operational metrics to Prometheus.

Examples include:

- action evaluations;
- allow decisions;
- deny decisions;
- high-risk operations;
- Kafka failures;
- OPA failures;
- incidents;
- containment events;
- anomaly evaluations;
- ML failures;
- model drift.

Grafana provides security-operations visibility while Alertmanager evaluates alerting conditions.

Observability is considered part of the security architecture because degraded controls must be detectable.

---

## Fail-Open vs Fail-Closed Boundaries

AgentShield intentionally distinguishes authoritative security dependencies from asynchronous supporting systems.

### Authorization Controls

Identity, session state, agent status, and policy enforcement are security-authoritative controls.

Failures in these controls should not silently grant privileged access.

### Analytics Controls

Kafka event publication, behavioral analytics, and ML inference provide detection and telemetry.

Failures in these systems must be visible but must not invalidate already completed authoritative security state transitions.

This distinction prevents observability infrastructure from becoming an accidental dependency for critical containment.

---

## Security Invariants

The system is designed around several important invariants:

```text
Suspended agent
    -> cannot create new sessions

Revoked session
    -> cannot authorize actions

Expired session
    -> cannot authorize actions

Unapproved sensitive operation
    -> cannot receive temporary authorization

Contained agent
    -> active sessions revoked

Contained agent
    -> active grants revoked

Kafka outage
    -> does not undo containment

ML outage
    -> does not bypass deterministic authorization

OPA outage
    -> does not silently allow protected actions
```

These invariants are validated through automated tests and failure-injection scenarios.

---

## Security Validation Evidence

Phase 38 security hardening validated several real failure and abuse scenarios.

### Containment Authorization

Observed results:

```text
No credentials       -> HTTP 401
Invalid credential   -> HTTP 401
Analyst credential   -> HTTP 403
Broker credential    -> HTTP 403
Admin credential     -> containment succeeds
```

### Containment Idempotency

Repeated containment produced:

```text
First call:
agent_status      = suspended
sessions_revoked  = 1

Second call:
agent_status      = suspended
sessions_revoked  = 0
```

No privileges were restored.

### Session-Creation Race

Concurrent session-creation attempts were executed before containment committed.

After containment:

```text
agent status      = suspended
active sessions   = 0
revoked sessions  = all previously created sessions
```

New session creation then returned:

```text
HTTP 409
agent is not active
```

### Post-Containment Action Reuse

A previously valid session was reused after containment.

Observed result:

```text
HTTP 401
invalid or expired AgentShield session
```

### ML Outage

With the PyTorch anomaly service unavailable:

```text
Gateway actions continued
ML inference failures were recorded
adaptive baseline processing continued
Detection health remained healthy
```

### Kafka Outage

With Redpanda unavailable:

```text
authorization completed
containment completed
agent was suspended
session was revoked
audit persisted
Kafka failures incremented
Gateway remained healthy
```

These tests demonstrate security behavior under actual dependency failure rather than relying only on design assumptions.

---

## Automated Security Regression Tests

The Gateway repository now includes integration tests covering:

```text
containment suspends an agent
active sessions are revoked
containment is idempotent
unknown agents return ErrAgentNotFound
suspended agents cannot create sessions
```

Kafka publishing also has a regression test that verifies broker publication honors context deadlines.

Additional tests exist across Detection and ML for behavioral analysis, baseline handling, feature extraction, ML clients, model metadata, API behavior, and model drift.

---

## Defense in Depth

AgentShield does not rely on one security mechanism.

Security enforcement combines:

```text
Agent Identity
      +
Short-Lived Sessions
      +
Contextual Risk
      +
OPA Policy
      +
Human Approval
      +
Temporary Grants
      +
Credential Isolation
      +
Behavioral Detection
      +
ML Anomaly Detection
      +
Containment
      +
Audit Logging
      +
NetworkPolicies
      +
Observability
```

Compromise or failure of one layer should not automatically compromise the entire control plane.

---

## Security Philosophy

Traditional identity systems primarily secure human users and applications.

AgentShield extends zero-trust principles to autonomous machine identities.

The central assumption is:

> Autonomous execution increases the importance of continuous authorization, short-lived authority, behavioral monitoring, and rapid containment.

AgentShield therefore treats every autonomous action as a security decision rather than assuming authentication creates permanent trust.

---

## Production Security Work Still Required

AgentShield is a portfolio and security-engineering system, not a claim of production certification.

Before real production deployment, additional work would include:

- managed secret storage;
- mTLS between internal services;
- workload identities;
- secret rotation automation;
- penetration testing;
- external identity-provider integration;
- distributed tracing;
- multi-node datastore resilience;
- cloud IAM integration;
- formal disaster-recovery testing;
- supply-chain signing and provenance;
- image vulnerability scanning;
- broader adversarial testing.
