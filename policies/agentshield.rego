package agentshield.authz

import rego.v1

default decision := {
    "decision": "DENY",
    "reason": "no AgentShield policy matched"
}

# Production database deletion is always denied.
decision := {
    "decision": "DENY",
    "reason": "production database deletion is prohibited"
} if {
    input.action == "database.delete"
    startswith(input.resource, "production/")
}

# Production deployments require human approval.
decision := {
    "decision": "REQUIRE_APPROVAL",
    "reason": "production code changes require human approval"
} if {
    input.agent_type == "devops"
    input.action == "github.push"
    startswith(input.resource, "production/")
}

# Production logs require approval.
decision := {
    "decision": "REQUIRE_APPROVAL",
    "reason": "production log access requires temporary approval"
} if {
    input.agent_type == "devops"
    input.action == "logs.read"
    startswith(input.resource, "production/")
}

# Development resources may be read.
decision := {
    "decision": "ALLOW",
    "reason": "read access to development resources is permitted"
} if {
    endswith(input.action, ".read")
    startswith(input.resource, "development/")
}