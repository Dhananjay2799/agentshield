const API_URL =
  process.env.AGENTSHIELD_API_URL ??
  "http://localhost:8080"

export const dynamic = "force-dynamic"

export async function GET(
  request: Request
) {
  const controller =
    new AbortController()

  request.signal.addEventListener(
    "abort",
    () => {
      controller.abort()
    }
  )

  try {
    const upstream = await fetch(
      `${API_URL}/v1/events/stream`,
      {
        method: "GET",
        cache: "no-store",
        signal: controller.signal,
        headers: {
          Accept: "text/event-stream",
        },
      }
    )

    if (
      !upstream.ok ||
      !upstream.body
    ) {
      return new Response(
        "Unable to open AgentShield event stream",
        {
          status:
            upstream.status || 502,
        }
      )
    }

    return new Response(
      upstream.body,
      {
        status: 200,
        headers: {
          "Content-Type":
            "text/event-stream",
          "Cache-Control":
            "no-cache, no-transform",
          Connection: "keep-alive",
          "X-Accel-Buffering":
            "no",
        },
      }
    )
  } catch (error) {
    if (
      error instanceof Error &&
      error.name === "AbortError"
    ) {
      return new Response(null, {
        status: 499,
      })
    }

    return new Response(
      "Unable to connect to AgentShield Gateway",
      {
        status: 502,
      }
    )
  }
}