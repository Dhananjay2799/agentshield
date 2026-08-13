import type {
  SecurityEvent,
} from "@/types/agentshield"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function getRecentEvents(
  limit = 100
): Promise<SecurityEvent[]> {
  try {
    const response = await fetch(
      `${API_URL}/v1/events?limit=${limit}`,
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