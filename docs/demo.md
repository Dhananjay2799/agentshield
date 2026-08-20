# AgentShield Demo Guide

## Overview

This demo presents AgentShield as a zero-trust security control plane for autonomous AI agents.

The goal is to demonstrate the project in approximately 5–10 minutes while focusing on the security controls and engineering decisions that matter most to an interviewer, recruiter, security engineer, or platform engineer.

The demo highlights:

- Kubernetes-based distributed deployment;
- autonomous agent identity;
- short-lived sessions;
- contextual risk evaluation;
- OPA/Rego policy-as-code;
- ALLOW, DENY, and REQUIRE_APPROVAL decisions;
- human approval and temporary authorization;
- Kafka-compatible security-event streaming with Redpanda;
- behavioral threat detection;
- PyTorch anomaly inference;
- incident visibility;
- Prometheus and Grafana observability;
- emergency containment;
- immediate session revocation;
- idempotent containment;
- dependency-failure resilience.

---

## Demo Story

Use one simple security story throughout the presentation:

> An autonomous AI agent is allowed to perform normal low-risk work. AgentShield continuously evaluates its identity, session, policy, risk, and behavior. If the agent becomes suspicious or compromised, AgentShield detects the behavior, creates security evidence, and allows an administrator to contain the agent immediately.

This keeps the demonstration focused on the problem AgentShield solves instead of turning it into a collection of unrelated API calls.

---

## Recommended Demo Order

```text
1. Show the architecture
2. Verify Kubernetes
3. Verify Gateway health
4. Register an autonomous agent
5. Create a short-lived session
6. Evaluate an agent action
7. Explain OPA/Rego authorization
8. Show security-event streaming
9. Show behavioral/ML detection
10. Show observability
11. Contain the agent
12. Prove the session is revoked
13. Explain failure resilience
14. Show automated tests and CI
```

---

# Part 1 — Architecture

Before running commands, show:

```text
docs/architecture.md
```

Use the architecture diagram to explain the request path:

```text
Autonomous AI Agent
        |
        v
AgentShield Gateway
        |
        v
Session Validation
        |
        v
Risk Engine
        |
        v
OPA / Rego
        |
        +------> ALLOW
        |
        +------> DENY
        |
        +------> REQUIRE_APPROVAL
                        |
                        v
                 Human Approval
                        |
                        v
                 Temporary Grant
```

Then explain the asynchronous security path:

```text
Gateway
   |
   v
Redpanda / Kafka
   |
   +------> Audit
   |
   +------> Detection
                  |
                  v
          Behavioral Engine
                  |
                  v
          PyTorch ML Anomaly
                  |
                  v
              Incident
                  |
                  v
             Containment
```

### What to Say

AgentShield separates authoritative access control from asynchronous analytics.

Identity, session validation, risk, policy, approval, and grants determine whether an operation is authorized.

Kafka, behavioral analytics, and machine learning provide detection and security telemetry, but failures in those analytical systems cannot silently bypass the authoritative authorization controls.

---

# Part 2 — Verify the Kubernetes Control Plane

Run:

```powershell
kubectl get pods `
  -n agentshield
```

The exact pod names and ages will vary.

The important result is that the core components are running:

```text
gateway
audit
credential-broker
detection
ml-anomaly
opa
postgres
redpanda
prometheus
grafana
alertmanager
alert-receiver
```

You can also show services:

```powershell
kubectl get services `
  -n agentshield
```

### What This Proves

AgentShield is not a single-process demo application.

It is a distributed security platform deployed on Kubernetes with separate authorization, credential, audit, detection, ML, policy, persistence, event-streaming, and observability components.

---

# Part 3 — Verify Gateway Health

If the Gateway is not already forwarded locally, open a separate PowerShell terminal and run:

```powershell
kubectl port-forward `
  -n agentshield `
  service/gateway `
  18080:8080
```

Keep that terminal open.

In the demo terminal:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:18080/health"
```

Expected result:

```text
service             status
-------             ------
agentshield-gateway healthy
```

### What This Proves

The AgentShield control-plane API is healthy and reachable.

---

# Part 4 — Load the Administrator Credential

For the containment portion of the demo, load the development administrator API key from Kubernetes without printing the credential itself.

```powershell
$adminKeyB64 = kubectl get secret agentshield-secrets `
  -n agentshield `
  -o jsonpath='{.data.AGENTSHIELD_ADMIN_API_KEY}'

$adminKey = [System.Text.Encoding]::UTF8.GetString(
    [System.Convert]::FromBase64String($adminKeyB64)
)

Write-Host "Admin key loaded:" ($adminKey.Length -gt 0)
```

Expected:

```text
Admin key loaded: True
```

Do not display the actual key during a recorded demo.

---

# Part 5 — Register an Autonomous Agent

Create a dedicated demo agent:

```powershell
$demoAgent = Invoke-RestMethod `
  -Uri "http://localhost:18080/v1/agents" `
  -Method POST `
  -ContentType "application/json" `
  -Body (@{
      name        = "portfolio-demo-agent"
      agent_type  = "devops"
      owner       = "platform-engineering"
      framework   = "langgraph"
      model       = "gpt-5"
      environment = "development"
  } | ConvertTo-Json)

$demoAgent | ConvertTo-Json -Depth 10
```

Store the ID:

```powershell
$demoAgentId = $demoAgent.id

Write-Host "Demo Agent ID:" $demoAgentId
```

### What to Point Out

The response should contain an AgentShield identity with fields such as:

```text
id
name
agent_type
owner
framework
model
environment
status
```

### What This Proves

Autonomous agents are represented as explicit security identities rather than being treated as anonymous application processes.

---

# Part 6 — Create a Short-Lived Agent Session

Create a session:

```powershell
$demoSession = Invoke-RestMethod `
  -Uri "http://localhost:18080/v1/agents/$demoAgentId/sessions" `
  -Method POST `
  -ContentType "application/json" `
  -Body (@{
      task_id     = "portfolio-demo-task"
      ttl_minutes = 30
  } | ConvertTo-Json)

$demoSession | ConvertTo-Json -Depth 10
```

Store the session ID:

```powershell
$demoSessionId = $demoSession.id
```

### What to Point Out

The session includes:

```text
agent_id
task_id
status
started_at
expires_at
```

### What This Proves

Agent registration does not provide permanent execution authority.

The agent operates through a time-bounded session that can expire or be revoked.

---

# Part 7 — Evaluate a Normal Agent Action

Evaluate a low-risk development operation:

```powershell
$allowResult = Invoke-RestMethod `
  -Uri "http://localhost:18080/v1/actions/evaluate" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
      "X-AgentShield-Session" = $demoSessionId
  } `
  -Body (@{
      action   = "logs.read"
      resource = "development/service-a"
      reason   = "Investigate application health during portfolio demo"
  } | ConvertTo-Json)

$allowResult | ConvertTo-Json -Depth 10
```

For the currently validated development policy, this type of operation should produce an ALLOW decision.

### What to Say

The Gateway does more than authenticate the agent.

The request is evaluated using the current session, agent context, risk logic, and OPA/Rego policy before AgentShield returns an authorization decision.

---

# Part 8 — Show OPA/Rego Policy-as-Code

Open:

```text
policies/agentshield.rego
```

Also show:

```text
policies/agentshield_test.rego
```

Explain that AgentShield supports authorization outcomes:

```text
ALLOW
DENY
REQUIRE_APPROVAL
```

### What to Say

Authorization policy is externalized from application code.

OPA and Rego make security decisions reviewable, testable, version-controlled, and suitable for CI.

Sensitive operations can require human approval rather than being represented only as a binary allow/deny decision.

---

# Part 9 — Show the Human Approval Architecture

For a concise interview demo, it is usually better to explain the already implemented approval path rather than spend several minutes constructing a new high-risk scenario.

The flow is:

```text
Agent Action
     |
     v
Risk + OPA
     |
     v
REQUIRE_APPROVAL
     |
     v
Approval Request
     |
     v
Human Decision
     |
     v
Temporary Authorization Grant
     |
     v
Authorized Retry
```

### What to Say

AgentShield can place a human checkpoint between autonomous reasoning and a sensitive real-world operation.

Approved authority is temporary rather than permanently added to the agent.

This reduces the blast radius of both compromised agents and incorrect autonomous decisions.

---

# Part 10 — Show the Security Event Pipeline

After evaluating actions, inspect recent Gateway logs:

```powershell
kubectl logs `
  -n agentshield `
  deployment/gateway `
  --since=5m |
Select-String `
  -Pattern "publishing security event|security event published|failed to publish"
```

Then verify Redpanda:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

Expected healthy state includes:

```text
Healthy: true
```

List topics:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk topic list
```

Point out:

```text
agentshield.security.events
agentshield.security.dlq
```

### What This Proves

Authorization activity is converted into a security-event stream that can be consumed independently by audit and detection services.

This decouples real-time authorization from downstream analytics.

---

# Part 11 — Show Detection

Generate a few additional normal actions:

```powershell
1..4 | ForEach-Object {
    $result = Invoke-RestMethod `
      -Uri "http://localhost:18080/v1/actions/evaluate" `
      -Method POST `
      -ContentType "application/json" `
      -Headers @{
          "X-AgentShield-Session" = $demoSessionId
      } `
      -Body (@{
          action   = "logs.read"
          resource = "development/service-a"
          reason   = "Portfolio demo detection event $_"
      } | ConvertTo-Json)

    Write-Host "event $_ -> $($result.decision)"
}
```

Wait briefly for asynchronous processing:

```powershell
Start-Sleep -Seconds 5
```

Inspect Detection logs for this agent:

```powershell
kubectl logs `
  -n agentshield `
  deployment/detection `
  --since=5m |
Select-String `
  -Pattern "$demoAgentId|ANOMALY SCORE|AGENT BASELINE|ML ANOMALY"
```

### What to Point Out

Depending on the current behavioral window, logs can show:

```text
security event
AGENT BASELINE
ANOMALY SCORE
ML ANOMALY SKIPPED
```

Do not claim that an anomaly occurred unless the logs actually show one.

### What This Proves

AgentShield does not rely only on static authorization rules.

It also observes how an agent behaves over time.

---

# Part 12 — Show the PyTorch ML Service

Verify ML readiness:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

A healthy response includes information such as:

```json
{
  "service": "agentshield-ml-anomaly",
  "status": "ready",
  "model_loaded": true,
  "model_version": "v1"
}
```

The exact model threshold and feature count can change between model versions.

### What to Say

The Detection service extracts behavioral features and sends eligible windows to the ML anomaly service.

The model supplements deterministic behavioral rules.

Machine learning is intentionally not placed in the authoritative authorization path, so an ML outage cannot become an authorization bypass.

---

# Part 13 — Show Detection Metrics

Run:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/debug/metrics
```

Useful fields include:

```text
processed_events
incidents_triggered
behavioral_detections
containment_events
rejected_events
fetch_errors
commit_failures
dlq_published
anomaly_evaluations
anomaly_detections
baseline_warmups
ml_predictions
ml_anomalies
ml_failures
ml_skipped_insufficient_window
```

### What This Proves

The detection pipeline is measurable.

Operators can distinguish normal processing, behavioral detections, ML inference, dependency failures, and event-processing failures.

---

# Part 14 — Show Gateway Security Metrics

Run:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:18080/debug/metrics" |
ConvertTo-Json -Depth 10
```

Useful fields include:

```text
action_evaluations
allow
deny
high_risk_actions
require_approval
opa_errors
kafka_failures
incidents_total
incidents_open
incidents_critical_open
incidents_resolved
```

### What to Say

AgentShield exposes security-control metrics, not only generic application CPU and memory telemetry.

---

# Part 15 — Show Grafana and Prometheus

If needed, forward Prometheus:

```powershell
kubectl port-forward `
  -n agentshield `
  service/prometheus `
  19090:9090
```

In another terminal, forward Grafana:

```powershell
kubectl port-forward `
  -n agentshield `
  service/grafana `
  13000:3000
```

The repository contains the AgentShield security operations dashboard at:

```text
infrastructure/kubernetes/base/observability/grafana/dashboards/agentshield-security-operations.json
```

### What to Show

Use the dashboard to discuss:

- authorization activity;
- security decisions;
- incident counts;
- behavioral detections;
- anomaly activity;
- Kafka failures;
- ML failures;
- containment events.

### What This Proves

AgentShield is designed to be operated as a security platform, not merely executed as an API.

---

# Part 16 — Emergency Containment

Now demonstrate the strongest end-to-end security control.

First confirm the demo agent is active:

```powershell
kubectl exec `
  -n agentshield `
  statefulset/postgres `
  -- psql `
  -U agentshield `
  -d agentshield `
  -c "
SELECT id, name, status
FROM agents
WHERE id = '$demoAgentId';
"
```

Contain the agent:

```powershell
$containment = Invoke-RestMethod `
  -Uri "http://localhost:18080/v1/agents/$demoAgentId/contain" `
  -Method POST `
  -Headers @{
      "Authorization" = "Bearer $adminKey"
  }

$containment | ConvertTo-Json -Depth 10
```

Expected shape:

```json
{
  "agent_id": "...",
  "agent_status": "suspended",
  "grants_revoked": 0,
  "sessions_revoked": 1,
  "status": "contained"
}
```

Counts depend on the current agent state.

### What to Say

Containment is an authoritative state transition.

AgentShield suspends the agent and revokes active execution authority immediately.

---

# Part 17 — Prove the Session Was Revoked

Verify the database:

```powershell
kubectl exec `
  -n agentshield `
  statefulset/postgres `
  -- psql `
  -U agentshield `
  -d agentshield `
  -c "
SELECT id, name, status
FROM agents
WHERE id = '$demoAgentId';

SELECT id, task_id, status, started_at, ended_at
FROM agent_sessions
WHERE agent_id = '$demoAgentId'
ORDER BY started_at;

SELECT event_type, action, decision, metadata, created_at
FROM audit_events
WHERE agent_id = '$demoAgentId'
ORDER BY created_at DESC
LIMIT 10;
"
```

The important state should be:

```text
agent status   = suspended
session status = revoked
```

---

# Part 18 — Prove the Old Session Cannot Execute

Try the previous session again:

```powershell
$postContainBody = @{
    action   = "logs.read"
    resource = "development/service-a"
    reason   = "Attempt action after containment"
} | ConvertTo-Json

try {
    Invoke-WebRequest `
      -Uri "http://localhost:18080/v1/actions/evaluate" `
      -Method POST `
      -ContentType "application/json" `
      -Headers @{
          "X-AgentShield-Session" = $demoSessionId
      } `
      -Body $postContainBody `
      -UseBasicParsing
}
catch {
    if ($_.Exception.Response) {
        Write-Host "HTTP STATUS:" ([int]$_.Exception.Response.StatusCode)
    }

    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}
```

The validated containment behavior rejects the revoked session with an authentication/session error.

A previously observed response is:

```text
401
{"error":"invalid or expired AgentShield session"}
```

### What This Proves

Containment is not cosmetic.

The existing execution session becomes unusable.

---

# Part 19 — Prove New Sessions Are Blocked

Try to create another session for the suspended agent:

```powershell
try {
    Invoke-WebRequest `
      -Uri "http://localhost:18080/v1/agents/$demoAgentId/sessions" `
      -Method POST `
      -ContentType "application/json" `
      -Body (@{
          task_id     = "post-containment-session"
          ttl_minutes = 30
      } | ConvertTo-Json) `
      -UseBasicParsing
}
catch {
    if ($_.Exception.Response) {
        Write-Host "HTTP STATUS:" ([int]$_.Exception.Response.StatusCode)
    }

    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}
```

Validated behavior:

```text
409
{"error":"agent is not active"}
```

### What This Proves

The agent cannot simply obtain a fresh session after containment.

---

# Part 20 — Demonstrate Idempotent Containment

Run containment again:

```powershell
$secondContainment = Invoke-RestMethod `
  -Uri "http://localhost:18080/v1/agents/$demoAgentId/contain" `
  -Method POST `
  -Headers @{
      "Authorization" = "Bearer $adminKey"
  }

$secondContainment | ConvertTo-Json -Depth 10
```

The agent remains suspended.

Already-revoked sessions and grants should not be counted again.

Typical second-call behavior is:

```text
agent_status     = suspended
sessions_revoked = 0
grants_revoked   = 0
status           = contained
```

### What This Proves

Containment can safely be retried by automation or an operator without repeatedly mutating already-contained state.

---

# Part 21 — Explain RBAC on Containment

AgentShield's containment endpoint is protected.

Previously validated behavior includes:

```text
No credentials       -> 401
Invalid credentials  -> 401
Analyst credential   -> 403
Broker credential    -> 403
Admin credential     -> containment allowed
```

### What to Say

Emergency security controls are themselves access controlled.

An agent, analyst, or credential-broker identity cannot arbitrarily suspend agents merely because it can reach the Gateway.

---

# Part 22 — Dependency Failure Resilience

This is an excellent engineering discussion if the interviewer wants deeper detail.

You do not need to intentionally break the cluster during every demo.

Explain the failure tests already implemented and validated.

## ML Failure

During resilience testing, the ML service was scaled to zero.

Detection recorded:

```text
ML ANOMALY INFERENCE FAILED
```

while the core Gateway remained healthy and normal authorization continued.

Detection metrics recorded ML failures.

### Security Property

```text
ML unavailable
      !=
authorization bypass
```

---

## Kafka/Redpanda Failure

During event-platform failure testing, Redpanda was made unavailable.

Containment still:

```text
suspended the agent
revoked the session
persisted the audit state
```

while event publication failed separately.

The Gateway uses bounded publishing behavior so a Kafka outage does not indefinitely block the authoritative containment operation.

### Security Property

```text
Kafka unavailable
      !=
containment rollback
```

This distinction is one of the most important design decisions in AgentShield.

---

# Part 23 — Show the Containment Integration Tests

Open:

```text
services/gateway/internal/repository/containment_repository_test.go
```

The repository contains integration coverage for security invariants including:

```text
containment suspends the agent
active sessions are revoked
containment is idempotent
unknown agents return not found
session creation rejects suspended agents
```

With the integration-test database configured, the validated test suite includes:

```text
TestContainAgentSuspendsAgentAndRevokesActiveSessions
TestContainAgentIsIdempotent
TestContainAgentReturnsNotFoundForUnknownAgent
TestSessionCreationRejectsSuspendedAgent
```

### What to Say

The project does not rely only on manual demonstrations for containment correctness.

Critical security-state transitions have automated integration coverage.

---

# Part 24 — Show Kafka Producer Tests

Open:

```text
services/gateway/internal/events/producer_test.go
```

Explain that event publication behavior is tested independently from the higher-level API path.

This is especially important because AgentShield intentionally separates authoritative state changes from event-stream delivery.

---

# Part 25 — Show CI

Open:

```text
.github/workflows/
```

The repository includes:

```text
ci.yml
gateway-security-ci.yml
kubernetes-ci.yml
```

Explain that CI validates application code, security-sensitive Gateway behavior, and Kubernetes configuration.

For local Gateway validation:

```powershell
cd ".\services\gateway"

go fmt ./...
go test ./... -count=1
go vet ./...

cd "..\.."

git diff --check
```

### What This Proves

AgentShield treats security controls as software that must be continuously tested rather than as architecture claims that exist only in documentation.

---

# Part 26 — Optional Deep-Dive: Concurrent Containment

If the interviewer asks about race conditions, discuss the Phase 38 concurrency validation.

AgentShield was tested while multiple session-creation requests were running concurrently with containment.

The resulting database state showed all sessions revoked and the agent suspended.

After containment, a new session request returned:

```text
409
{"error":"agent is not active"}
```

### Why This Matters

Security containment must remain correct under concurrent activity.

A compromised autonomous agent may continue issuing requests while an operator is attempting to stop it.

---

# Part 27 — Optional Deep-Dive: Model Drift

If discussing ML operations, show the ML/detection observability path rather than claiming a specific drift state unless the current system actually reports one.

Explain:

```text
Behavioral events
      |
      v
Feature extraction
      |
      v
Agent behavioral baseline
      |
      v
PyTorch anomaly inference
      |
      v
Anomaly score
      |
      v
Operational metrics / monitoring
```

Model versions, thresholds, prediction failures, anomaly evaluations, and behavioral baselines provide the foundation for monitoring ML behavior over time.

### Important Demo Rule

Do not manufacture an anomaly or drift claim just for presentation.

Show the actual metrics/logs generated by the running system and explain how the architecture supports monitoring.

---

# Part 28 — Five-Minute Interview Version

If time is limited, use only these steps.

## Minute 1 — Problem and Architecture

Say:

> Autonomous AI agents can call APIs, infrastructure, databases, and deployment systems. AgentShield treats each agent as a zero-trust identity and evaluates every sensitive action using short-lived sessions, contextual risk, OPA/Rego policy, human approval, and behavioral monitoring.

Show:

```text
docs/architecture.md
```

## Minute 2 — Live Agent Authorization

Run:

```text
Create agent
Create session
Evaluate logs.read
```

Show the ALLOW result.

## Minute 3 — Detection and ML

Show:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/debug/metrics
```

Then:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

Explain Redpanda + behavioral detection + PyTorch.

## Minute 4 — Containment

Contain the agent.

Show:

```text
agent_status = suspended
sessions_revoked >= 1
```

Then retry the old session and show rejection.

## Minute 5 — Engineering Depth

Show:

```text
Kubernetes
NetworkPolicies
Prometheus
Grafana
GitHub Actions
Go integration tests
Rego tests
```

Finish with the failure-resilience design.

---

# Part 29 — Recruiter-Friendly Explanation

For a non-security recruiter, describe AgentShield like this:

> AgentShield is a security platform for autonomous AI agents. It gives each AI agent an identity, limits it to short-lived sessions, checks every sensitive action against policy and risk, can require human approval, monitors behavior using event streaming and machine learning, and can immediately suspend a compromised agent and revoke its active access.

Avoid starting with implementation details.

Explain the problem first, then the architecture.

---

# Part 30 — Security Engineer Explanation

For a security-focused interviewer:

> AgentShield applies zero-trust principles to autonomous agent identities. Authorization is enforced synchronously through session validation, contextual risk, and OPA/Rego. High-impact actions can require human approval and scoped temporary grants. Security events are streamed asynchronously through Redpanda to audit and detection services. Behavioral rules and PyTorch anomaly inference supplement deterministic controls. Containment changes authoritative PostgreSQL state, revokes active sessions and grants, and remains effective even when event publication fails.

---

# Part 31 — Platform Engineer Explanation

For a platform or DevOps interviewer:

> AgentShield is deployed as a Kubernetes-native distributed control plane. It uses Go microservices, PostgreSQL for authoritative state, Redpanda for Kafka-compatible events, OPA for policy-as-code, PyTorch for anomaly inference, NetworkPolicies for service isolation, and Prometheus/Grafana/Alertmanager for observability. CI validates Go code, security behavior, policy, and Kubernetes configuration.

---

# Part 32 — Key Design Decisions to Emphasize

During interviews, emphasize these decisions:

1. Agents are first-class security identities.
2. Authentication does not imply authorization.
3. Sessions are short lived and revocable.
4. Policy is externalized through OPA/Rego.
5. Sensitive operations can require humans.
6. Temporary grants avoid permanent privilege expansion.
7. Credentials are issued downstream of authorization.
8. Security events are asynchronous.
9. ML supplements rather than replaces deterministic controls.
10. Containment modifies authoritative state.
11. Containment is idempotent.
12. Suspended agents cannot create new sessions.
13. Kafka failure does not roll back containment.
14. ML failure does not bypass authorization.
15. Network access follows default-deny/least-privilege principles.
16. Critical security invariants have automated tests.

---

# Part 33 — What Not to Do During the Demo

Avoid:

- exposing Kubernetes secrets;
- printing API keys;
- spending most of the demo waiting for Docker builds;
- rebuilding every service live;
- deleting production-like persistent data;
- intentionally corrupting the database;
- claiming an anomaly that the system did not detect;
- claiming model drift without actual evidence;
- describing asynchronous ML as the authorization mechanism;
- presenting Kafka availability as required for successful containment;
- showing 20 minutes of terminal output without explaining the security property being demonstrated.

The demo should focus on security outcomes.

---

# Part 34 — Recommended Screens to Keep Open

Before an interview, prepare:

```text
Terminal 1
Gateway port-forward

Terminal 2
Demo PowerShell commands

Terminal 3
kubectl logs / health checks

Browser Tab 1
GitHub README

Browser Tab 2
architecture.md

Browser Tab 3
Grafana

Editor
agentshield.rego
containment_repository_test.go
```

This avoids wasting interview time navigating the repository.

---

# Part 35 — Pre-Demo Checklist

Before presenting:

```powershell
kubectl get pods `
  -n agentshield
```

Verify Gateway:

```powershell
kubectl exec `
  -n agentshield `
  deployment/gateway `
  -- curl -fsS http://localhost:8080/health
```

Verify Detection:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/health
```

Verify ML:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

Verify Redpanda:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

Verify repository state:

```powershell
git status
```

Verify tests before the interview rather than running the entire setup from scratch during the presentation.

---

# Part 36 — Demo Cleanup

The demo intentionally ends with the demo agent suspended.

You can leave the security evidence in the local development database if useful for later demonstrations, or clean up test data using an explicit development-only procedure.

Never run ad hoc destructive cleanup against a production database.

Clear sensitive PowerShell variables after the demo:

```powershell
$adminKey = $null
$adminKeyB64 = $null
```

---

# Part 37 — Closing Statement

A concise closing statement:

> AgentShield demonstrates how autonomous AI agents can be governed like zero-trust machine identities. Instead of giving an agent permanent credentials and trusting its reasoning, AgentShield continuously evaluates identity, session, risk, policy, approvals, and behavior. It combines Go, Kubernetes, OPA/Rego, PostgreSQL, Redpanda/Kafka, PyTorch, Prometheus, Grafana, and automated security testing to provide authorization, detection, auditability, and rapid containment in one control plane.

---

## Related Documentation

For deeper details, see:

```text
README.md
docs/architecture.md
docs/security-model.md
docs/deployment.md
```

Use this demo guide for the live walkthrough and the other documents for architectural, security, and operational depth.
