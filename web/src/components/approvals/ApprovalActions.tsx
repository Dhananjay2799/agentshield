"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"

import {
  CheckCircle2,
  Loader2,
  XCircle,
} from "lucide-react"

import { Button } from "@/components/ui/button"

type ApprovalActionsProps = {
  approvalId: string
}

type ApprovalAction = "approve" | "deny"

export default function ApprovalActions({
  approvalId,
}: ApprovalActionsProps) {
  const router = useRouter()

  const [loading, setLoading] =
    useState<ApprovalAction | null>(null)

  const [error, setError] =
    useState<string | null>(null)

  async function updateApproval(
    action: ApprovalAction
  ) {
    try {
      setLoading(action)
      setError(null)

      const response = await fetch(
        `/api/approvals/${approvalId}/${action}`,
        {
          method: "POST",
        }
      )

      if (!response.ok) {
        const payload = await response
          .json()
          .catch(() => null)

        throw new Error(
          payload?.error ??
            `Request failed with status ${response.status}`
        )
      }

      router.refresh()
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Unable to update approval."
      )
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-end gap-2">
        <Button
          size="sm"
          variant="outline"
          disabled={loading !== null}
          onClick={() =>
            updateApproval("approve")
          }
          className="border-emerald-500/30 bg-emerald-500/10 text-emerald-300 hover:bg-emerald-500/20 hover:text-emerald-200"
        >
          {loading === "approve" ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <CheckCircle2 className="mr-2 h-4 w-4" />
          )}

          Approve
        </Button>

        <Button
          size="sm"
          variant="outline"
          disabled={loading !== null}
          onClick={() =>
            updateApproval("deny")
          }
          className="border-red-500/30 bg-red-500/10 text-red-300 hover:bg-red-500/20 hover:text-red-200"
        >
          {loading === "deny" ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <XCircle className="mr-2 h-4 w-4" />
          )}

          Deny
        </Button>
      </div>

      {error && (
        <div className="max-w-xs text-right text-xs text-red-400">
          {error}
        </div>
      )}
    </div>
  )
}