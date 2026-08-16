import { NextResponse } from "next/server"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function POST(
  request: Request
) {
  try {
    const body = await request.json()

    const response = await fetch(
      `${API_URL}/v1/policies`,
      {
        method: "POST",
        cache: "no-store",
        headers: {
          "Content-Type":
            "application/json",
        },
        body: JSON.stringify(body),
      }
    )

    const data = await response.json()

    return NextResponse.json(
      data,
      {
        status: response.status,
      }
    )
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