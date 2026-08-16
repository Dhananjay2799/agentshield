export type Policy = {
  id: string
  name: string
  description: string
  effect: "ALLOW" | "REQUIRE_APPROVAL" | "DENY"
  status: "draft" | "active" | "disabled" | "archived"
  priority: number
  agent_type?: string | null
  action: string
  action_match: "exact" | "prefix" | "suffix"
  resource: string
  resource_match: "exact" | "prefix" | "suffix"
  environment?: string | null
  version: number
  source: string
  created_by: string
  created_at: string
  updated_at: string
}

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function getPolicies(): Promise<Policy[]> {
  try {
    const response = await fetch(
      `${API_URL}/v1/policies`,
      {
        cache: "no-store",
      }
    )

    if (!response.ok) {
      return []
    }

    return response.json()
  } catch {
    return []
  }
}

export async function getPolicy(
  id: string
): Promise<Policy | null> {
  try {
    const response = await fetch(
      `${API_URL}/v1/policies/${id}`,
      {
        cache: "no-store",
      }
    )

    if (!response.ok) {
      return null
    }

    return response.json()
  } catch {
    return null
  }
}