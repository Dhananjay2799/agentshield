# AgentShield Deployment and Operations Guide

## Overview

AgentShield is deployed as a distributed security control plane for autonomous AI agents.

The local reference deployment uses Kubernetes and Kustomize and includes the AgentShield application services, PostgreSQL, Redpanda, Open Policy Agent (OPA), Prometheus, Grafana, Alertmanager, and the PyTorch-based ML anomaly service.

This guide covers:

- deployment prerequisites;
- Kubernetes and Kustomize deployment;
- configuration and secrets;
- database migrations;
- Redpanda topic initialization;
- health and readiness verification;
- observability;
- NetworkPolicies;
- rolling updates;
- failure recovery;
- troubleshooting;
- production hardening considerations.

The commands in this document assume PowerShell and the repository root as the current working directory unless stated otherwise.

---

## Architecture at Deployment Time

A typical AgentShield deployment contains the following workloads:

| Component | Purpose |
| --- | --- |
| Gateway | Agent identity, sessions, authorization, approvals, grants, incidents, containment, and security-event publishing |
| Credential Broker | Issues short-lived credentials after AgentShield authorization |
| Audit | Consumes security events and persists audit information |
| Detection | Consumes security events, maintains behavioral state, detects suspicious behavior, and creates incidents |
| ML Anomaly | PyTorch-based anomaly inference service |
| OPA | Executes Rego authorization policies |
| PostgreSQL | Stores authoritative AgentShield state |
| Redpanda | Kafka-compatible security-event transport |
| Prometheus | Collects operational and security metrics |
| Grafana | Visualizes AgentShield security and operational telemetry |
| Alertmanager | Routes Prometheus alerts |
| Alert Receiver | Receives configured alert notifications |

The Kubernetes manifests are organized under:

```text
infrastructure/kubernetes/
├── base/
│   ├── audit/
│   ├── credential-broker/
│   ├── detection/
│   ├── gateway/
│   ├── jobs/
│   ├── ml-anomaly/
│   ├── network/
│   ├── networking/
│   ├── observability/
│   ├── opa/
│   ├── postgres/
│   └── redpanda/
└── overlays/
    └── local/
```

---

## Prerequisites

For the local Kubernetes deployment, install and configure:

- Git
- Docker Desktop
- Kubernetes
- `kubectl`
- Kustomize support through `kubectl`
- Go for local service development and testing
- Python where required by ML development workflows

Verify the primary tools:

```powershell
git --version
docker version
kubectl version --client
go version
```

Verify Kubernetes connectivity:

```powershell
kubectl cluster-info
kubectl get nodes
```

The local cluster must be running before AgentShield workloads can be deployed.

---

## Repository Location

Example local repository location:

```text
C:\Users\<user>\OneDrive\Desktop\project\AgentShield
```

Move to the repository root:

```powershell
cd "C:\path\to\AgentShield"
```

Verify the expected Kubernetes structure:

```powershell
Get-ChildItem ".\infrastructure\kubernetes"
```

---

## Configuration

Shared non-secret Kubernetes configuration is defined through the `agentshield-config` ConfigMap.

The current deployment uses values such as:

```text
AGENTSHIELD_GATEWAY_URL=http://gateway:8080
DATABASE_HOST=postgres
DATABASE_NAME=agentshield
DATABASE_PORT=5432
KAFKA_BROKER=redpanda:9092
KAFKA_DLQ_TOPIC=agentshield.security.dlq
KAFKA_SECURITY_TOPIC=agentshield.security.events
OPA_URL=http://opa:8181
```

Inspect the deployed ConfigMap:

```powershell
kubectl get configmap agentshield-config `
  -n agentshield `
  -o yaml
```

Configuration that differs between environments should be supplied through overlays or environment-specific deployment configuration rather than hard-coded into application source code.

---

## Secrets

Sensitive configuration is stored in the `agentshield-secrets` Kubernetes Secret.

The deployment currently expects secret values including:

```text
AGENTSHIELD_ADMIN_API_KEY
AGENTSHIELD_ANALYST_API_KEY
AGENTSHIELD_CREDENTIAL_BROKER_API_KEY
AGENTSHIELD_CREDENTIAL_SIGNING_SECRET
POSTGRES_USER
POSTGRES_PASSWORD
```

A template is maintained at:

```text
infrastructure/kubernetes/base/secret.example.yaml
```

Do not commit real production secrets to Git.

Inspect only secret key names when troubleshooting:

```powershell
kubectl get secret agentshield-secrets `
  -n agentshield `
  -o json |
ConvertFrom-Json |
Select-Object -ExpandProperty data |
Get-Member -MemberType NoteProperty |
Select-Object -ExpandProperty Name
```

Avoid printing decoded credentials into terminal logs, CI logs, screenshots, documentation, or issue reports.

### Production Secret Management

For a production environment, replace manually managed Kubernetes secrets with an external secret-management solution such as:

- Google Secret Manager;
- HashiCorp Vault;
- AWS Secrets Manager;
- Azure Key Vault;
- External Secrets Operator.

Credentials should be rotated regularly and scoped according to least privilege.

---

## Build Application Images

AgentShield services are packaged as container images.

A typical local gateway build is:

```powershell
docker build `
  -t agentshield-gateway:local `
  ".\services\gateway"
```

Other services can be built using their respective Dockerfiles.

Examples:

```powershell
docker build -t agentshield-audit:local ".\services\audit"
docker build -t agentshield-credential-broker:local ".\services\credential-broker"
docker build -t agentshield-detection:local ".\services\detection"
docker build -t agentshield-ml-anomaly:local ".\services\ml-anomaly"
```

Use immutable image versions for shared or production environments.

For example:

```text
agentshield-gateway:v1.0.0
agentshield-detection:v1.0.0
agentshield-ml-anomaly:v1.0.0
```

Do not rely on mutable tags such as `latest` for production releases.

---

## Local Kubernetes Image Availability

The exact image-loading workflow depends on the Kubernetes runtime.

Docker Desktop Kubernetes may require locally built images to be available to the container runtime used by the cluster.

Before troubleshooting an `ImagePullBackOff`, inspect the image configured on the workload:

```powershell
kubectl get deployment gateway `
  -n agentshield `
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

Then inspect pod events:

```powershell
kubectl describe pod `
  -n agentshield `
  -l app=gateway
```

For repeatable production deployments, publish versioned images to a container registry rather than depending on workstation-local images.

---

## Deploy with Kustomize

AgentShield provides a local overlay under:

```text
infrastructure/kubernetes/overlays/local
```

Preview the generated Kubernetes resources before applying them:

```powershell
kubectl kustomize ".\infrastructure\kubernetes\overlays\local"
```

Apply the local deployment:

```powershell
kubectl apply `
  -k ".\infrastructure\kubernetes\overlays\local"
```

Alternatively, if the repository-level Kubernetes kustomization is being used for the current environment:

```powershell
kubectl apply `
  -k ".\infrastructure\kubernetes"
```

Use the path appropriate to the environment being deployed.

---

## Namespace Verification

AgentShield runs in the `agentshield` namespace.

Verify it:

```powershell
kubectl get namespace agentshield
```

Inspect all resources:

```powershell
kubectl get all `
  -n agentshield
```

---

## Database Deployment

PostgreSQL stores authoritative AgentShield state, including data associated with:

- agents;
- sessions;
- audit events;
- approvals;
- authorization grants;
- incidents;
- behavioral profiles;
- policy-related state.

Verify PostgreSQL:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=postgres
```

Verify the service:

```powershell
kubectl get service postgres `
  -n agentshield
```

Connect from inside the cluster:

```powershell
kubectl exec `
  -n agentshield `
  statefulset/postgres `
  -- psql `
  -U agentshield `
  -d agentshield
```

Exit `psql` with:

```text
\q
```

---

## Database Migrations

Database migrations are stored under:

```text
infrastructure/kubernetes/base/postgres/migrations/
```

The repository currently contains migrations including:

```text
001_initial_schema.sql
002_incident_lifecycle.sql
003_agent_behavior_profiles.sql
```

Kubernetes executes migrations through the database migration Job.

Inspect it:

```powershell
kubectl get job agentshield-database-migration `
  -n agentshield
```

Inspect migration logs:

```powershell
kubectl logs `
  -n agentshield `
  job/agentshield-database-migration
```

A deployment should not be considered healthy if required migrations have failed.

### Migration Safety

Production migrations should:

- be version controlled;
- be reviewed before release;
- avoid destructive changes without a rollback plan;
- be backward compatible during rolling deployments where practical;
- be backed by tested database recovery procedures.

---

## Redpanda Deployment

AgentShield uses Redpanda as a Kafka-compatible event platform.

The primary security topic is:

```text
agentshield.security.events
```

The dead-letter topic is:

```text
agentshield.security.dlq
```

Verify Redpanda:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=redpanda
```

Check cluster health:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

Expected healthy state includes:

```text
Healthy: true
Leaderless partitions (0): []
Under-replicated partitions (0): []
```

---

## Redpanda Topic Initialization

AgentShield includes a Kubernetes Job for topic initialization.

Inspect it:

```powershell
kubectl get job agentshield-redpanda-topics `
  -n agentshield
```

View logs:

```powershell
kubectl logs `
  -n agentshield `
  job/agentshield-redpanda-topics
```

List topics directly:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk topic list
```

The security-event and DLQ topics should exist before event-dependent services are considered fully operational.

---

## OPA Deployment

OPA provides policy-as-code authorization using Rego.

Verify the OPA pod:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=opa
```

Verify the service:

```powershell
kubectl get service opa `
  -n agentshield
```

The Gateway communicates with OPA through the internal service URL:

```text
http://opa:8181
```

Authorization is an authoritative security control. OPA failures must not silently become authorization success.

---

## Gateway Deployment

The Gateway is the primary control-plane API.

Verify rollout:

```powershell
kubectl rollout status `
  deployment/gateway `
  -n agentshield `
  --timeout=180s
```

Inspect pods:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=gateway `
  -o wide
```

Check health from inside the pod:

```powershell
kubectl exec `
  -n agentshield `
  deployment/gateway `
  -- curl -fsS http://localhost:8080/health
```

Expected response:

```json
{
  "service": "agentshield-gateway",
  "status": "healthy"
}
```

---

## Local Gateway Access

For local testing, forward the Gateway service:

```powershell
kubectl port-forward `
  -n agentshield `
  service/gateway `
  18080:8080
```

Keep that terminal open.

From another PowerShell terminal:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:18080/health"
```

The Gateway API is then available locally at:

```text
http://localhost:18080
```

---

## Credential Broker

The Credential Broker issues short-lived credentials after AgentShield authorization requirements have been satisfied.

Verify the deployment:

```powershell
kubectl rollout status `
  deployment/credential-broker `
  -n agentshield `
  --timeout=180s
```

Inspect pods:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=credential-broker
```

Credential issuance should remain separate from general agent execution so long-lived infrastructure credentials do not need to be embedded in autonomous agents.

---

## Audit Service

The Audit service consumes security events and contributes to the permanent audit trail.

Verify deployment:

```powershell
kubectl rollout status `
  deployment/audit `
  -n agentshield `
  --timeout=180s
```

Inspect recent logs:

```powershell
kubectl logs `
  -n agentshield `
  deployment/audit `
  --since=5m
```

Audit records are security evidence and should be protected from unauthorized modification or deletion.

---

## Detection Service

The Detection service consumes security events and performs behavioral security analysis.

Verify deployment:

```powershell
kubectl rollout status `
  deployment/detection `
  -n agentshield `
  --timeout=180s
```

Check health:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/health
```

Inspect detection metrics:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/debug/metrics
```

Metrics include operational and security signals such as:

- processed events;
- denied events;
- high-risk events;
- incidents triggered;
- behavioral detections;
- containment events;
- rejected events;
- fetch errors;
- commit failures;
- DLQ publications;
- anomaly evaluations;
- anomaly detections;
- ML predictions;
- ML failures.

---

## ML Anomaly Service

The ML anomaly service provides PyTorch-based behavioral anomaly inference.

Verify deployment:

```powershell
kubectl rollout status `
  deployment/ml-anomaly `
  -n agentshield `
  --timeout=180s
```

Verify readiness from the Detection service:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

A ready response includes model information such as:

```json
{
  "service": "agentshield-ml-anomaly",
  "status": "ready",
  "model_loaded": true,
  "model_version": "v1"
}
```

The exact threshold and feature count are model-dependent and should not be hard-coded into operational expectations.

---

## ML Failure Behavior

ML anomaly detection is supplementary security analytics rather than the authoritative authorization layer.

If the ML service is unavailable:

- Gateway authorization must continue to enforce identity, session, risk, and policy controls;
- Detection should record inference failures;
- ML failure metrics should increase;
- the failure must not silently become an authorization bypass.

A controlled local resilience test can scale the service down:

```powershell
kubectl scale deployment/ml-anomaly `
  -n agentshield `
  --replicas=0
```

After testing, restore it immediately:

```powershell
kubectl scale deployment/ml-anomaly `
  -n agentshield `
  --replicas=1

kubectl rollout status `
  deployment/ml-anomaly `
  -n agentshield `
  --timeout=180s
```

Do not perform failure injection against production without an approved resilience-testing procedure.

---

## Prometheus

Prometheus collects AgentShield metrics.

Verify deployment:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=prometheus
```

Inspect the service:

```powershell
kubectl get service prometheus `
  -n agentshield
```

For local access:

```powershell
kubectl port-forward `
  -n agentshield `
  service/prometheus `
  19090:9090
```

Prometheus is then available through the local forwarded port.

---

## Grafana

Grafana provides dashboards for security and operational visibility.

Dashboard configuration is stored under:

```text
infrastructure/kubernetes/base/observability/grafana/
```

The AgentShield security operations dashboard is stored at:

```text
infrastructure/kubernetes/base/observability/grafana/dashboards/agentshield-security-operations.json
```

Verify Grafana:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=grafana
```

For local access:

```powershell
kubectl port-forward `
  -n agentshield `
  service/grafana `
  13000:3000
```

---

## Alertmanager

Alertmanager receives alerts generated from Prometheus alert rules and routes them according to the configured notification pipeline.

Verify it:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=alertmanager
```

Inspect Prometheus alert rules:

```text
infrastructure/kubernetes/base/observability/prometheus/alerts.yaml
```

Inspect Alertmanager configuration under:

```text
infrastructure/kubernetes/base/observability/alertmanager/
```

---

## NetworkPolicies

AgentShield applies Kubernetes NetworkPolicies to reduce unnecessary service-to-service connectivity.

The deployment includes policies for components such as:

- Gateway;
- Audit;
- Credential Broker;
- Detection;
- ML Anomaly;
- OPA;
- PostgreSQL;
- Redpanda;
- Prometheus;
- Grafana;
- Alertmanager;
- database migration jobs;
- Redpanda topic jobs;
- DNS.

A default-deny policy forms the baseline.

Inspect policies:

```powershell
kubectl get networkpolicy `
  -n agentshield
```

Describe a specific policy:

```powershell
kubectl describe networkpolicy gateway-access `
  -n agentshield
```

The intended principle is:

```text
default deny
      +
explicitly required communication
      =
least-privilege service networking
```

Production deployments should verify that the Kubernetes networking implementation actually enforces NetworkPolicy.

---

## Deployment Verification

After applying AgentShield, perform a structured verification.

### 1. Check workloads

```powershell
kubectl get pods `
  -n agentshield
```

Look for unexpected states such as:

```text
CrashLoopBackOff
ImagePullBackOff
Error
Pending
```

### 2. Check services

```powershell
kubectl get services `
  -n agentshield
```

### 3. Check jobs

```powershell
kubectl get jobs `
  -n agentshield
```

Database migration and Redpanda topic jobs should complete successfully.

### 4. Check Gateway

```powershell
kubectl exec `
  -n agentshield `
  deployment/gateway `
  -- curl -fsS http://localhost:8080/health
```

### 5. Check Detection

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://localhost:8083/health
```

### 6. Check ML readiness

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

### 7. Check Redpanda

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

### 8. Check recent errors

```powershell
kubectl logs `
  -n agentshield `
  deployment/gateway `
  --since=5m

kubectl logs `
  -n agentshield `
  deployment/detection `
  --since=5m
```

---

## Rolling Updates

After building and making a new image available to the cluster, update the deployment image through the appropriate Kustomize configuration or manifest.

Then apply the configuration:

```powershell
kubectl apply `
  -k ".\infrastructure\kubernetes\overlays\local"
```

Monitor rollout:

```powershell
kubectl rollout status `
  deployment/gateway `
  -n agentshield `
  --timeout=180s
```

Inspect the deployed image:

```powershell
kubectl get deployment gateway `
  -n agentshield `
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

For a failed deployment, inspect rollout history:

```powershell
kubectl rollout history `
  deployment/gateway `
  -n agentshield
```

Rollback only after understanding whether database or configuration changes are compatible with the previous application version.

---

## Security Containment Operational Behavior

Containment is an authoritative security operation.

When an agent is contained, AgentShield is designed to:

1. change the agent state to suspended;
2. revoke active sessions;
3. revoke applicable temporary grants;
4. preserve audit evidence;
5. publish a containment security event when the event platform is available.

Containment state is committed independently of Kafka availability. A security-event publication failure must not roll back successful containment.

This is important because event infrastructure is telemetry and coordination infrastructure, while containment changes authoritative security state.

---

## Dependency Failure Principles

AgentShield distinguishes authoritative dependencies from asynchronous or analytical dependencies.

### Authorization dependencies

Failures in controls required to authorize a privileged operation should fail safely.

Examples include:

- invalid or expired session;
- unavailable required policy decision;
- missing approval;
- invalid authorization grant.

These conditions must not become implicit ALLOW decisions.

### Event-stream failures

Security-state changes that have already succeeded should not be undone merely because a Kafka/Redpanda publication fails.

The Gateway uses bounded event-publishing behavior and records publication failures.

### ML failures

ML inference failures must be observable, but ML availability must not determine whether core zero-trust authorization is enforced.

This separation prevents analytics outages from becoming authorization bypasses.

---

## Kafka/Redpanda Outage Recovery

If Redpanda is unavailable:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=redpanda
```

Inspect Redpanda logs:

```powershell
kubectl logs `
  -n agentshield `
  redpanda-0
```

After recovery:

```powershell
kubectl rollout status `
  statefulset/redpanda `
  -n agentshield `
  --timeout=180s

kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

Also inspect Gateway and Detection metrics/logs for event publication or consumption failures.

---

## PostgreSQL Recovery Checks

If PostgreSQL becomes unavailable, inspect:

```powershell
kubectl get pod `
  -n agentshield `
  -l app=postgres
```

Then:

```powershell
kubectl describe pod `
  -n agentshield `
  -l app=postgres
```

And:

```powershell
kubectl logs `
  -n agentshield `
  statefulset/postgres
```

Check persistent storage:

```powershell
kubectl get pvc `
  -n agentshield
```

Because PostgreSQL stores authoritative security state, production deployments require tested backup, restore, and disaster-recovery procedures.

---

## Troubleshooting

### Pod stuck in `ImagePullBackOff`

Inspect the pod:

```powershell
kubectl describe pod `
  -n agentshield `
  <pod-name>
```

Confirm:

- the image name is correct;
- the tag exists;
- the cluster can access the registry;
- required registry credentials are configured;
- the local Kubernetes runtime can see locally built images when using a workstation-only deployment.

---

### Pod in `CrashLoopBackOff`

Inspect current logs:

```powershell
kubectl logs `
  -n agentshield `
  <pod-name>
```

Inspect the previous container instance:

```powershell
kubectl logs `
  -n agentshield `
  <pod-name> `
  --previous
```

Check configuration, secrets, database connectivity, service DNS, and startup dependencies.

---

### Gateway is unhealthy

Check:

```powershell
kubectl logs `
  -n agentshield `
  deployment/gateway `
  --since=10m
```

Verify PostgreSQL:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=postgres
```

Verify OPA:

```powershell
kubectl get pods `
  -n agentshield `
  -l app=opa
```

Verify configuration:

```powershell
kubectl get configmap agentshield-config `
  -n agentshield `
  -o yaml
```

---

### Security events are not appearing

Check Gateway logs:

```powershell
kubectl logs `
  -n agentshield `
  deployment/gateway `
  --since=10m
```

Check Redpanda:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk cluster health
```

Check topics:

```powershell
kubectl exec `
  -n agentshield `
  redpanda-0 `
  -- rpk topic list
```

Check Detection:

```powershell
kubectl logs `
  -n agentshield `
  deployment/detection `
  --since=10m
```

---

### ML inference failures

Verify ML readiness:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

Inspect ML logs:

```powershell
kubectl logs `
  -n agentshield `
  deployment/ml-anomaly `
  --since=10m
```

Inspect Detection logs:

```powershell
kubectl logs `
  -n agentshield `
  deployment/detection `
  --since=10m |
Select-String `
  -Pattern "ML ANOMALY INFERENCE FAILED|ML ANOMALY SKIPPED|ANOMALY SCORE"
```

---

### Network connectivity failures

Inspect NetworkPolicies:

```powershell
kubectl get networkpolicy `
  -n agentshield
```

Check service DNS and connectivity from the calling workload.

For example:

```powershell
kubectl exec `
  -n agentshield `
  deployment/detection `
  -- curl -fsS http://ml-anomaly:8085/ready
```

Do not solve connectivity problems by permanently removing the default-deny security model. Add the minimum required network permission instead.

---

## Operational Metrics

Operators should monitor both system reliability and security behavior.

Important metric categories include:

### Gateway

- action evaluations;
- ALLOW decisions;
- DENY decisions;
- approval-required decisions;
- high-risk actions;
- OPA errors;
- Kafka publication failures;
- incident state counts.

### Detection

- processed events;
- rejected events;
- consumer fetch errors;
- commit failures;
- behavioral detections;
- incidents triggered;
- containment events;
- DLQ publications;
- anomaly evaluations;
- ML predictions;
- ML failures;
- anomaly detections.

### Infrastructure

- pod availability;
- container restarts;
- CPU and memory pressure;
- PostgreSQL availability;
- Redpanda health;
- topic health;
- persistent-volume capacity;
- Prometheus target health;
- alert delivery status.

---

## Logs

Use Kubernetes logs for component-level investigation.

Gateway:

```powershell
kubectl logs `
  -n agentshield `
  deployment/gateway `
  --since=10m
```

Detection:

```powershell
kubectl logs `
  -n agentshield `
  deployment/detection `
  --since=10m
```

Credential Broker:

```powershell
kubectl logs `
  -n agentshield `
  deployment/credential-broker `
  --since=10m
```

Audit:

```powershell
kubectl logs `
  -n agentshield `
  deployment/audit `
  --since=10m
```

ML Anomaly:

```powershell
kubectl logs `
  -n agentshield `
  deployment/ml-anomaly `
  --since=10m
```

Production environments should aggregate logs centrally and define appropriate retention and access controls.

---

## CI/CD

AgentShield includes GitHub Actions workflows under:

```text
.github/workflows/
```

Current workflows include:

```text
ci.yml
gateway-security-ci.yml
kubernetes-ci.yml
```

CI should verify applicable checks such as:

- Go formatting;
- Go tests;
- `go vet`;
- security-sensitive gateway tests;
- Rego policy tests;
- Kubernetes manifest generation or validation;
- configuration consistency.

Production delivery should extend this with:

- immutable container builds;
- vulnerability scanning;
- software bill of materials generation;
- artifact signing;
- provenance/attestation;
- registry publishing;
- environment promotion;
- deployment approval controls;
- post-deployment health verification.

---

## Local Development Validation

Before committing Go changes, run the appropriate service tests.

For the Gateway:

```powershell
cd ".\services\gateway"

go fmt ./...
go test ./... -count=1
go vet ./...

cd "..\.."
git diff --check
```

Repository integration tests that require PostgreSQL should use an explicitly configured test database URL and must clean up the data they create.

Do not point destructive tests at a production database.

---

## Production Hardening

The local Kubernetes deployment demonstrates AgentShield architecture and security controls, but a production deployment requires additional controls.

### High Availability

Production should consider:

- multiple Gateway replicas;
- multiple Detection replicas where consumer semantics support them;
- highly available PostgreSQL;
- multi-node Kafka/Redpanda;
- replicated observability infrastructure;
- disruption budgets;
- anti-affinity and topology-spread policies.

### TLS

Production traffic should use encrypted transport.

Apply TLS to:

- external Gateway traffic;
- service-to-service traffic where required by the threat model;
- database connections;
- Kafka/Redpanda connections;
- observability endpoints containing sensitive telemetry.

### Authentication

Development API keys should be replaced with stronger workload and operator identity mechanisms.

Potential approaches include:

- workload identity;
- service accounts with short-lived tokens;
- SPIFFE/SPIRE;
- OIDC;
- cloud-native identity;
- mTLS identities.

### Secret Management

Use a dedicated secret-management system rather than storing long-lived production credentials directly in manifests.

### Container Security

Production workloads should use:

- non-root containers;
- read-only root filesystems where possible;
- dropped Linux capabilities;
- seccomp profiles;
- resource requests and limits;
- vulnerability-scanned images;
- minimal runtime images.

### Kubernetes Security

Production clusters should implement:

- RBAC least privilege;
- Pod Security Standards;
- NetworkPolicy enforcement;
- admission controls;
- restricted administrative access;
- audit logging;
- namespace isolation;
- controlled image registries.

### Database Protection

PostgreSQL should have:

- encrypted storage;
- encrypted connections;
- automated backups;
- restore testing;
- restricted network access;
- least-privilege database users;
- retention policies appropriate to audit evidence.

### Event Platform Protection

Kafka/Redpanda should have:

- authentication;
- TLS;
- authorization;
- controlled topic access;
- retention policies;
- replication;
- monitoring for under-replicated or unavailable partitions.

### Observability Protection

Prometheus, Grafana, and Alertmanager expose sensitive operational information.

Production deployments should require authentication and restrict network access.

---

## Google Cloud Deployment Direction

The local deployment is designed so the architecture can be moved to a managed cloud environment.

A future Google Cloud production architecture can map components approximately as follows:

```text
AgentShield Kubernetes workloads
        |
        v
Google Kubernetes Engine (GKE)

PostgreSQL
        |
        v
Cloud SQL for PostgreSQL

Container images
        |
        v
Artifact Registry

Secrets
        |
        v
Secret Manager

Metrics / logs
        |
        v
Managed Prometheus / Cloud Monitoring / Cloud Logging
```

Redpanda may remain Kubernetes-hosted or be replaced with an appropriate managed Kafka-compatible deployment depending on operational requirements.

Cloud deployment should preserve the same security boundaries rather than weakening them for convenience.

---

## Backup and Recovery

A production recovery plan should cover at minimum:

- PostgreSQL backups;
- PostgreSQL point-in-time recovery where appropriate;
- Kubernetes configuration stored in Git;
- policy source stored in Git;
- container images stored in a registry;
- ML model artifacts and version metadata;
- Grafana dashboard definitions;
- Prometheus alert rules;
- secret recovery/rotation procedures.

Recovery procedures should be tested rather than assumed.

---

## Upgrade Strategy

Recommended release process:

```text
Code change
   |
   v
Automated tests
   |
   v
Security and policy checks
   |
   v
Container build
   |
   v
Image scan
   |
   v
Versioned artifact
   |
   v
Staging deployment
   |
   v
Integration / resilience validation
   |
   v
Production approval
   |
   v
Rolling deployment
   |
   v
Health and security verification
```

Database migrations and application rollbacks must be designed together.

---

## Deployment Security Invariants

The following properties should remain true across environments:

1. Suspended agents cannot create usable new sessions.
2. Revoked or expired sessions cannot authorize actions.
3. High-impact actions cannot bypass required approval.
4. Temporary grants remain limited in scope and lifetime.
5. Credential issuance remains downstream of authorization.
6. OPA failures cannot silently become authorization success.
7. Kafka/Redpanda failures cannot undo completed containment.
8. ML failures cannot bypass authoritative access control.
9. Security-relevant state changes remain auditable.
10. Network access follows least privilege.
11. Production secrets are not committed to source control.
12. Security failures are observable through logs, metrics, or alerts.

These invariants are more important than any individual deployment technology.

---

## Quick Operational Checklist

Before considering an AgentShield deployment ready:

- [ ] Kubernetes cluster is reachable.
- [ ] `agentshield` namespace exists.
- [ ] required ConfigMaps are present.
- [ ] required Secrets are present.
- [ ] PostgreSQL is healthy.
- [ ] database migrations completed.
- [ ] Redpanda is healthy.
- [ ] required topics exist.
- [ ] OPA is running.
- [ ] Gateway is healthy.
- [ ] Credential Broker is running.
- [ ] Audit service is running.
- [ ] Detection service is healthy.
- [ ] ML model reports ready.
- [ ] Prometheus is collecting metrics.
- [ ] Grafana is available.
- [ ] Alertmanager is running.
- [ ] NetworkPolicies are installed.
- [ ] no unexpected `CrashLoopBackOff` or `ImagePullBackOff` exists.
- [ ] security-event publishing works.
- [ ] authorization tests pass.
- [ ] containment tests pass.
- [ ] dependency-failure behavior has been validated in the target environment.

---

## Related Documentation

See:

```text
README.md
docs/architecture.md
docs/security-model.md
docs/demo.md
```

Together these documents describe the project overview, architecture, security model, demonstration workflow, and deployment/operations model.

---

## Summary

AgentShield is deployed as a defense-in-depth security control plane rather than a single authorization service.

The operational architecture combines:

```text
Agent Identity
      |
      v
Short-Lived Session
      |
      v
Risk Evaluation
      |
      v
OPA / Rego Authorization
      |
      +----> Human Approval / Temporary Grant
      |
      v
Credential Broker
      |
      v
Security Event Stream
      |
      v
Audit + Behavioral Detection + ML Anomaly Detection
      |
      v
Incident Response
      |
      v
Containment
```

Kubernetes provides workload orchestration and network isolation, PostgreSQL stores authoritative security state, Redpanda transports security events, OPA enforces policy-as-code, PyTorch provides anomaly inference, and Prometheus/Grafana/Alertmanager provide operational visibility.

The central deployment principle is that failures in asynchronous telemetry or analytical systems must never weaken AgentShield's authoritative zero-trust controls.
