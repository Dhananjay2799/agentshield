"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import {
  CheckCircle2,
  Loader2,
  Search,
  XCircle,
} from "lucide-react"

import { Button } from "@/components/ui/button"

type IncidentActionsProps = {
  incidentId: string
  status: string
}

type ActionType = "investigate" | "resolve" | "dismiss"

export default function IncidentActions({
  incidentId,
  status,
}: IncidentActionsProps) {
  const router = useRouter()

  const [loading, setLoading] = useState<ActionType | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function updateIncident(action: ActionType) {
    try {
      setLoading(action)
      setError(null)

      const response = await fetch(
        `http://localhost:8080/v1/incidents/${incidentId}/${action}`,
        {
          method: "POST",
        }
      )

      if (!response.ok) {
        const message = await response.text()

        throw new Error(
          message || `Request failed with status ${response.status}`
        )
      }

      router.refresh()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to update incident."
      )
    } finally {
      setLoading(null)
    }
  }

  const closed =
    status === "resolved" || status === "dismissed"

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-3">
        <Button
          variant="outline"
          disabled={loading !== null || closed || status === "investigating"}
          onClick={() => updateIncident("investigate")}
          className="border-amber-500/30 bg-amber-500/10 text-amber-300 hover:bg-amber-500/20 hover:text-amber-200"
        >
          {loading === "investigate" ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Search className="mr-2 h-4 w-4" />
          )}

          Investigate
        </Button>

        <Button
          variant="outline"
          disabled={loading !== null || closed}
          onClick={() => updateIncident("resolve")}
          className="border-emerald-500/30 bg-emerald-500/10 text-emerald-300 hover:bg-emerald-500/20 hover:text-emerald-200"
        >
          {loading === "resolve" ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <CheckCircle2 className="mr-2 h-4 w-4" />
          )}

          Resolve
        </Button>

        <Button
          variant="outline"
          disabled={loading !== null || closed}
          onClick={() => updateIncident("dismiss")}
          className="border-red-500/30 bg-red-500/10 text-red-300 hover:bg-red-500/20 hover:text-red-200"
        >
          {loading === "dismiss" ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <XCircle className="mr-2 h-4 w-4" />
          )}

          Dismiss
        </Button>
      </div>

      {closed && (
        <p className="text-sm text-zinc-500">
          This incident is already {status}. No further lifecycle actions
          are available.
        </p>
      )}

      {error && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-300">
          {error}
        </div>
      )}
    </div>
  )
}