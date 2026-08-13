import { NextResponse } from "next/server"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function GET(
  request: Request
) {
  const { searchParams } =
    new URL(request.url)

  const limit =
    searchParams.get("limit") ?? "100"

  try {
    const response = await fetch(
      `${API_URL}/v1/events?limit=${limit}`,
      {
        cache: "no-store",
      }
    )

    const payload = await response
      .json()
      .catch(() => [])

    return NextResponse.json(
      payload,
      {
        status: response.status,
      }
    )
  } catch {
    return NextResponse.json(
      {
        error:
          "Unable to connect to AgentShield Gateway",
      },
      {
        status: 502,
      }
    )
  }
}