import { NextResponse } from "next/server"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function POST(
  _request: Request,
  context: {
    params: Promise<{ id: string }>
  }
) {
  const { id } = await context.params

  try {
    const response = await fetch(
      `${API_URL}/v1/policies/${id}/deactivate`,
      {
        method: "POST",
        cache: "no-store",
      }
    )

    const data = await response.json()

    return NextResponse.json(data, {
      status: response.status,
    })
  } catch {
    return NextResponse.json(
      {
        error:
          "Unable to connect to AgentShield Gateway.",
      },
      {
        status: 502,
      }
    )
  }
}