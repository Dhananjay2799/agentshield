package agentshield.authz

import rego.v1

default decision := {
    "decision": "DENY",
    "reason": "no AgentShield policy matched"
}

#
# ============================================================
# IMMUTABLE STATIC GUARDRAILS
# ============================================================
#

# Production database deletion is always denied.
decision := {
    "decision": "DENY",
    "reason": "production database deletion is prohibited"
} if {
    input.action == "database.delete"
    startswith(input.resource, "production/")
}

#
# ============================================================
# MANAGED POLICY CONTROL PLANE
# ============================================================
#

# Managed policies are stored dynamically under:
#
# data.agentshield_runtime.managed_policies
#
managed_policies := object.get(
    data.agentshield_runtime,
    "managed_policies",
    {}
)

#
# Determine whether a managed policy matches the request.
#
policy_matches(policy) if {
    agent_type_matches(policy)
    environment_matches(policy)
    action_matches(policy)
    resource_matches(policy)
}

#
# Agent type
#
# Missing/null agent_type means "all agent types".
#
agent_type_matches(policy) if {
    object.get(policy, "agent_type", null) == null
}

agent_type_matches(policy) if {
    policy.agent_type == input.agent_type
}

#
# Environment
#
# Missing/null environment means "all environments".
#
environment_matches(policy) if {
    object.get(policy, "environment", null) == null
}

environment_matches(policy) if {
    policy.environment == input.environment
}

#
# Action matcher
#
action_matches(policy) if {
    policy.action_match == "exact"
    input.action == policy.action
}

action_matches(policy) if {
    policy.action_match == "prefix"
    startswith(input.action, policy.action)
}

action_matches(policy) if {
    policy.action_match == "suffix"
    endswith(input.action, policy.action)
}

#
# Resource matcher
#
resource_matches(policy) if {
    policy.resource_match == "exact"
    input.resource == policy.resource
}

resource_matches(policy) if {
    policy.resource_match == "prefix"
    startswith(input.resource, policy.resource)
}

resource_matches(policy) if {
    policy.resource_match == "suffix"
    endswith(input.resource, policy.resource)
}

#
# Gather all matching managed policies.
#
matching_policies := [
    policy |
    some policy_id
    policy := managed_policies[policy_id]
    policy_matches(policy)
]

#
# Gather matching priorities.
#
matching_priorities := [
    policy.priority |
    some policy in matching_policies
]

#
# The lowest numeric value represents the strongest priority.
#
winning_priority := min(matching_priorities) if {
    count(matching_priorities) > 0
}

#
# Gather policies at the winning priority.
#
winning_policies := [
    policy |
    some policy in matching_policies
    policy.priority == winning_priority
]

#
# Explicit conflict resolution for policies sharing a priority.
#
# DENY > REQUIRE_APPROVAL > ALLOW
#
winning_effect := "DENY" if {
    some policy in winning_policies
    policy.effect == "DENY"
}

winning_effect := "REQUIRE_APPROVAL" if {
    not winning_effect_deny
    some policy in winning_policies
    policy.effect == "REQUIRE_APPROVAL"
}

winning_effect := "ALLOW" if {
    not winning_effect_deny
    not winning_effect_require_approval
    some policy in winning_policies
    policy.effect == "ALLOW"
}

winning_effect_deny if {
    some policy in winning_policies
    policy.effect == "DENY"
}

winning_effect_require_approval if {
    some policy in winning_policies
    policy.effect == "REQUIRE_APPROVAL"
}

#
# Use one representative winning policy for the decision trace.
#
winning_policy := policy if {
    some policy in winning_policies
    policy.effect == winning_effect
}

#
# Managed policy decision.
#
# The matched_policy block provides structured evidence to the
# AgentShield Gateway for explainable authorization decisions.
#
decision := {
    "decision": winning_effect,
    "reason": sprintf(
        "managed policy: %s",
        [
            object.get(
                winning_policy,
                "name",
                "unnamed managed policy"
            )
        ]
    ),
    "matched_policy": {
        "id": object.get(
            winning_policy,
            "id",
            ""
        ),
        "name": object.get(
            winning_policy,
            "name",
            "unnamed managed policy"
        ),
        "priority": object.get(
            winning_policy,
            "priority",
            0
        ),
        "effect": object.get(
            winning_policy,
            "effect",
            winning_effect
        ),
        "version": object.get(
            winning_policy,
            "version",
            0
        ),
        "source": object.get(
            winning_policy,
            "source",
            "managed_policy"
        )
    }
} if {
    count(matching_policies) > 0
}

#
# ============================================================
# LEGACY STATIC POLICIES
# ============================================================
#
# These remain while the control plane is being migrated.
#

#
# Production code changes require human approval.
#
decision := {
    "decision": "REQUIRE_APPROVAL",
    "reason": "production code changes require human approval"
} if {
    input.agent_type == "devops"
    input.action == "github.push"
    startswith(input.resource, "production/")
    count(matching_policies) == 0
}

#
# Production logs require temporary human approval.
#
decision := {
    "decision": "REQUIRE_APPROVAL",
    "reason": "production log access requires temporary approval"
} if {
    input.agent_type == "devops"
    input.action == "logs.read"
    startswith(input.resource, "production/")
    count(matching_policies) == 0
}

#
# Development resources may be read.
#
decision := {
    "decision": "ALLOW",
    "reason": "read access to development resources is permitted"
} if {
    endswith(input.action, ".read")
    startswith(input.resource, "development/")
    count(matching_policies) == 0
}