package agentshield.authz

import rego.v1

test_devops_production_logs_require_approval if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "logs.read",
        "resource": "production/api"
    }

    result.decision == "REQUIRE_APPROVAL"
}

test_devops_production_push_requires_approval if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "github.push",
        "resource": "production/repository"
    }

    result.decision == "REQUIRE_APPROVAL"
}

test_production_database_delete_is_denied if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "database.delete",
        "resource": "production/payments"
    }

    result.decision == "DENY"
}

test_unknown_action_is_denied_by_default if {
    result := decision with input as {
        "agent_type": "devops",
        "action": "something.unknown",
        "resource": "production/test"
    }

    result.decision == "DENY"
}