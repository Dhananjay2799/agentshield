package agentshield.authz

import rego.v1

test_devops_production_logs_require_approval if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "logs.read",
        "resource": "production/api",
        "environment": "development"
    }

    result.decision == "REQUIRE_APPROVAL"
}

test_devops_production_push_requires_approval if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "github.push",
        "resource": "production/repository",
        "environment": "development"
    }

    result.decision == "REQUIRE_APPROVAL"
}

test_production_database_delete_is_denied if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "database.delete",
        "resource": "production/payments",
        "environment": "development"
    }

    result.decision == "DENY"
}

test_unknown_action_is_denied_by_default if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "something.unknown",
        "resource": "production/test",
        "environment": "development"
    }

    result.decision == "DENY"
}

test_managed_kubernetes_policy_requires_approval if {
    result := decision
        with input as {
            "agent_type": "devops",
            "action": "kubernetes.deploy",
            "resource": "production/payments-api",
            "environment": "production"
        }
        with data.agentshield_runtime.managed_policies as {
            "test-policy": {
                "id": "test-policy",
                "name": "Production Kubernetes Deployment",
                "effect": "REQUIRE_APPROVAL",
                "priority": 50,
                "agent_type": "devops",
                "action": "kubernetes.deploy",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            }
        }

    result.decision == "REQUIRE_APPROVAL"
}

test_managed_policy_agent_type_must_match if {
    result := decision
        with input as {
            "agent_type": "research",
            "action": "kubernetes.deploy",
            "resource": "production/payments-api",
            "environment": "production"
        }
        with data.agentshield.managed_policies as {
            "test-policy": {
                "id": "test-policy",
                "name": "Production Kubernetes Deployment",
                "effect": "REQUIRE_APPROVAL",
                "priority": 50,
                "agent_type": "devops",
                "action": "kubernetes.deploy",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            }
        }

    result.decision == "DENY"
}

test_lower_priority_number_wins if {
    result := decision
        with input as {
            "agent_type": "devops",
            "action": "kubernetes.deploy",
            "resource": "production/api",
            "environment": "production"
        }
        with data.agentshield_runtime.managed_policies as {
            "allow-policy": {
                "id": "allow-policy",
                "name": "Allow Kubernetes",
                "effect": "ALLOW",
                "priority": 100,
                "agent_type": "devops",
                "action": "kubernetes.deploy",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            },
            "approval-policy": {
                "id": "approval-policy",
                "name": "Require Kubernetes Approval",
                "effect": "REQUIRE_APPROVAL",
                "priority": 50,
                "agent_type": "devops",
                "action": "kubernetes.deploy",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            }
        }

    result.decision == "REQUIRE_APPROVAL"
}

test_deny_wins_same_priority if {
    result := decision
        with input as {
            "agent_type": "devops",
            "action": "secrets.read",
            "resource": "production/secrets",
            "environment": "production"
        }
        with data.agentshield.managed_policies as {
            "allow-policy": {
                "id": "allow-policy",
                "name": "Allow Secrets",
                "effect": "ALLOW",
                "priority": 10,
                "agent_type": "devops",
                "action": "secrets.read",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            },
            "deny-policy": {
                "id": "deny-policy",
                "name": "Deny Secrets",
                "effect": "DENY",
                "priority": 10,
                "agent_type": "devops",
                "action": "secrets.read",
                "action_match": "exact",
                "resource": "production/",
                "resource_match": "prefix",
                "environment": "production"
            }
        }

    result.decision == "DENY"
}