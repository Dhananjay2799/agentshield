import { NextResponse } from "next/server"

const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export async function POST(
  _request: Request,
  {
    params,
  }: {
    params: Promise<{ id: string }>
  }
) {
  const { id } = await params

  try {
    const response = await fetch(
      `${API_URL}/v1/approvals/${id}/deny`,
      {
        method: "POST",
        cache: "no-store",
      }
    )

    const payload = await response
      .json()
      .catch(() => ({
        error: "Gateway returned an invalid response",
      }))

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