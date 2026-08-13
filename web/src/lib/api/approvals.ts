import type { ApprovalRequest } from "@/types/agentshield"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function getPendingApprovals(): Promise<
  ApprovalRequest[]
> {
  try {
    const response = await fetch(
      `${API_URL}/v1/approvals`,
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