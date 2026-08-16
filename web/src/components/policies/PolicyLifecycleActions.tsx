"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import {
  CheckCircle2,
  Loader2,
  Power,
  PowerOff,
  ShieldCheck,
} from "lucide-react"

import { Button } from "@/components/ui/button"

type PolicyStatus =
  | "draft"
  | "active"
  | "disabled"
  | "archived"

type ValidationResult = {
  valid?: boolean
  status?: string
  checks?: Record<
    string,
    {
      status?: string
      message?: string
    }
  >
  error?: string
}

export function PolicyLifecycleActions({
  policyId,
  status,
}: {
  policyId: string
  status: PolicyStatus
}) {
  const router = useRouter()

  const [pendingAction, setPendingAction] =
    useState<string | null>(null)

  const [message, setMessage] =
    useState<string | null>(null)

  const [error, setError] =
    useState<string | null>(null)

  const [validation, setValidation] =
    useState<ValidationResult | null>(null)

  async function runAction(
    action:
      | "validate"
      | "activate"
      | "deactivate"
  ) {
    setPendingAction(action)
    setMessage(null)
    setError(null)

    if (action === "validate") {
      setValidation(null)
    }

    try {
      const response = await fetch(
        `/api/policies/${policyId}/${action}`,
        {
          method: "POST",
        }
      )

      const data = await response.json()

      if (!response.ok) {
        throw new Error(
          data?.error ||
            `Policy ${action} failed.`
        )
      }

      if (action === "validate") {
        setValidation(data)

        if (data.valid) {
          setMessage(
            "Policy validation passed."
          )
        } else {
          setError(
            "Policy validation failed."
          )
        }
      }

      if (action === "activate") {
        setMessage(
          "Policy activated and synchronized with OPA."
        )
      }

      if (action === "deactivate") {
        setMessage(
          "Policy disabled and removed from OPA runtime enforcement."
        )
      }

      router.refresh()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Policy lifecycle operation failed."
      )
    } finally {
      setPendingAction(null)
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap gap-3">
        {status === "draft" && (
          <>
            <Button
              variant="outline"
              disabled={pendingAction !== null}
              onClick={() =>
                runAction("validate")
              }
            >
              {pendingAction ===
              "validate" ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <CheckCircle2 className="mr-2 h-4 w-4" />
              )}

              Validate
            </Button>

            <Button
              disabled={pendingAction !== null}
              onClick={() =>
                runAction("activate")
              }
            >
              {pendingAction ===
              "activate" ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Power className="mr-2 h-4 w-4" />
              )}

              Activate
            </Button>
          </>
        )}

        {status === "disabled" && (
          <>
            <Button
              variant="outline"
              disabled={pendingAction !== null}
              onClick={() =>
                runAction("validate")
              }
            >
              {pendingAction ===
              "validate" ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <ShieldCheck className="mr-2 h-4 w-4" />
              )}

              Validate
            </Button>

            <Button
              disabled={pendingAction !== null}
              onClick={() =>
                runAction("activate")
              }
            >
              {pendingAction ===
              "activate" ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Power className="mr-2 h-4 w-4" />
              )}

              Reactivate
            </Button>
          </>
        )}

        {status === "active" && (
          <Button
            variant="outline"
            disabled={pendingAction !== null}
            onClick={() =>
              runAction("deactivate")
            }
          >
            {pendingAction ===
            "deactivate" ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <PowerOff className="mr-2 h-4 w-4" />
            )}

            Deactivate
          </Button>
        )}

        {status === "archived" && (
          <div className="text-sm text-zinc-500">
            Archived policies cannot be activated.
          </div>
        )}
      </div>

      {message && (
        <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-4 text-sm text-emerald-300">
          {message}
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-red-500/20 bg-red-500/5 p-4 text-sm text-red-300">
          {error}
        </div>
      )}

      {validation?.checks && (
        <div className="grid gap-3 md:grid-cols-3">
          {Object.entries(
            validation.checks
          ).map(([name, check]) => (
            <div
              key={name}
              className="rounded-lg border border-zinc-800 bg-zinc-950/40 p-4"
            >
              <div className="text-xs uppercase tracking-wide text-zinc-600">
                {name.replaceAll(
                  "_",
                  " "
                )}
              </div>

              <div
                className={[
                  "mt-2 text-sm font-medium",
                  check.status === "passed"
                    ? "text-emerald-300"
                    : "text-red-300",
                ].join(" ")}
              >
                {check.status ?? "unknown"}
              </div>

              {check.message && (
                <p className="mt-2 text-xs leading-5 text-zinc-500">
                  {check.message}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}