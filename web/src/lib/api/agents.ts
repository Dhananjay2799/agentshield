import type {
  Agent,
  AgentSession,
  ApprovalRequest,
  AuditEvent,
  Incident,
  SessionSecurity,
} from "@/types/agentshield"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

async function getJSON<T>(
  path: string
): Promise<T> {
  const response = await fetch(
    `${API_URL}${path}`,
    {
      cache: "no-store",
    }
  )

  if (!response.ok) {
    throw new Error(
      `AgentShield API request failed: ${response.status} ${response.statusText}`
    )
  }

  return response.json()
}

export async function getAgent(
  agentId: string
): Promise<Agent | null> {
  try {
    return await getJSON<Agent>(
      `/v1/agents/${agentId}`
    )
  } catch {
    return null
  }
}

export async function getAgentSessions(
  agentId: string
): Promise<AgentSession[]> {
  try {
    return await getJSON<AgentSession[]>(
      `/v1/agents/${agentId}/sessions`
    )
  } catch {
    return []
  }
}

export async function getAgentSessionSecurity(
  agentId: string
): Promise<SessionSecurity[]> {
  try {
    return await getJSON<SessionSecurity[]>(
      `/v1/agents/${agentId}/sessions/security`
    )
  } catch {
    return []
  }
}

export async function getAgentActions(
  agentId: string
): Promise<AuditEvent[]> {
  try {
    return await getJSON<AuditEvent[]>(
      `/v1/agents/${agentId}/actions`
    )
  } catch {
    return []
  }
}

export async function getAgentApprovals(
  agentId: string
): Promise<ApprovalRequest[]> {
  try {
    return await getJSON<ApprovalRequest[]>(
      `/v1/agents/${agentId}/approvals`
    )
  } catch {
    return []
  }
}

export async function getIncidents(): Promise<Incident[]> {
  try {
    return await getJSON<Incident[]>(
      "/v1/incidents"
    )
  } catch {
    return []
  }
}