"use client"

import {
  FormEvent,
  useState,
} from "react"

import { useRouter } from "next/navigation"

import {
  ArrowLeft,
  Loader2,
  ShieldPlus,
} from "lucide-react"

import Link from "next/link"

import { Button } from "@/components/ui/button"

type Effect =
  | "ALLOW"
  | "REQUIRE_APPROVAL"
  | "DENY"

type MatchType =
  | "exact"
  | "prefix"
  | "suffix"

type FormState = {
  name: string
  description: string
  effect: Effect
  priority: string
  agentType: string
  action: string
  actionMatch: MatchType
  resource: string
  resourceMatch: MatchType
  environment: string
  createdBy: string
}

const initialState: FormState = {
  name: "",
  description: "",
  effect: "REQUIRE_APPROVAL",
  priority: "100",
  agentType: "",
  action: "",
  actionMatch: "exact",
  resource: "",
  resourceMatch: "prefix",
  environment: "",
  createdBy: "soc-admin",
}

export function CreatePolicyForm() {
  const router = useRouter()

  const [form, setForm] =
    useState<FormState>(initialState)

  const [submitting, setSubmitting] =
    useState(false)

  const [error, setError] =
    useState<string | null>(null)

  function updateField<K extends keyof FormState>(
    field: K,
    value: FormState[K]
  ) {
    setForm((current) => ({
      ...current,
      [field]: value,
    }))
  }

  async function handleSubmit(
    event: FormEvent<HTMLFormElement>
  ) {
    event.preventDefault()

    setSubmitting(true)
    setError(null)

    try {
      const priority =
        Number.parseInt(
          form.priority,
          10
        )

      if (
        !Number.isFinite(priority) ||
        priority <= 0
      ) {
        throw new Error(
          "Priority must be greater than zero."
        )
      }

      const response = await fetch(
        "/api/policies",
        {
          method: "POST",
          headers: {
            "Content-Type":
              "application/json",
          },
          body: JSON.stringify({
            name: form.name.trim(),

            description:
              form.description.trim(),

            effect: form.effect,

            priority,

            agent_type:
              form.agentType.trim() ||
              null,

            action:
              form.action.trim(),

            action_match:
              form.actionMatch,

            resource:
              form.resource.trim(),

            resource_match:
              form.resourceMatch,

            environment:
              form.environment.trim() ||
              null,

            created_by:
              form.createdBy.trim() ||
              "soc-admin",
          }),
        }
      )

      const data = await response.json()

      if (!response.ok) {
        throw new Error(
          data?.error ||
            "Unable to create policy."
        )
      }

      if (!data?.id) {
        throw new Error(
          "Policy was created but no policy ID was returned."
        )
      }

      router.push(
        `/policies/${data.id}`
      )

      router.refresh()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to create policy."
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <Link
        href="/policies"
        className="inline-flex items-center gap-2 text-sm text-zinc-500 transition hover:text-zinc-200"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Policies
      </Link>

      <div>
        <div className="flex items-center gap-3">
          <ShieldPlus className="h-6 w-6 text-emerald-400" />

          <h2 className="text-2xl font-semibold tracking-tight">
            Create Policy
          </h2>
        </div>

        <p className="mt-2 max-w-3xl text-sm leading-6 text-zinc-500">
          Define an autonomous-agent authorization
          policy. New policies are always created as
          drafts and must be validated before runtime
          activation.
        </p>
      </div>

      <form
        onSubmit={handleSubmit}
        className="space-y-6"
      >
        <section className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-6">
          <h3 className="font-medium text-zinc-200">
            Policy Identity
          </h3>

          <div className="mt-5 grid gap-5">
            <Field label="Policy Name">
              <input
                required
                value={form.name}
                onChange={(event) =>
                  updateField(
                    "name",
                    event.target.value
                  )
                }
                placeholder="Production Database Protection"
                className={inputClasses}
              />
            </Field>

            <Field label="Description">
              <textarea
                required
                rows={4}
                value={form.description}
                onChange={(event) =>
                  updateField(
                    "description",
                    event.target.value
                  )
                }
                placeholder="Describe what this policy protects and why it exists."
                className={inputClasses}
              />
            </Field>

            <div className="grid gap-5 md:grid-cols-2">
              <Field label="Created By">
                <input
                  value={form.createdBy}
                  onChange={(event) =>
                    updateField(
                      "createdBy",
                      event.target.value
                    )
                  }
                  className={inputClasses}
                />
              </Field>

              <Field label="Priority">
                <input
                  required
                  type="number"
                  min="1"
                  value={form.priority}
                  onChange={(event) =>
                    updateField(
                      "priority",
                      event.target.value
                    )
                  }
                  className={inputClasses}
                />

                <p className="mt-2 text-xs text-zinc-600">
                  Lower numeric values have
                  stronger priority.
                </p>
              </Field>
            </div>
          </div>
        </section>

        <section className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-6">
          <h3 className="font-medium text-zinc-200">
            Authorization Effect
          </h3>

          <div className="mt-5">
            <Field label="Decision">
              <select
                value={form.effect}
                onChange={(event) =>
                  updateField(
                    "effect",
                    event.target
                      .value as Effect
                  )
                }
                className={inputClasses}
              >
                <option value="ALLOW">
                  ALLOW
                </option>

                <option value="REQUIRE_APPROVAL">
                  REQUIRE_APPROVAL
                </option>

                <option value="DENY">
                  DENY
                </option>
              </select>
            </Field>
          </div>
        </section>

        <section className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-6">
          <h3 className="font-medium text-zinc-200">
            Match Conditions
          </h3>

          <div className="mt-5 grid gap-5">
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="Agent Type">
                <input
                  value={form.agentType}
                  onChange={(event) =>
                    updateField(
                      "agentType",
                      event.target.value
                    )
                  }
                  placeholder="devops"
                  className={inputClasses}
                />

                <p className="mt-2 text-xs text-zinc-600">
                  Leave blank to match any
                  agent type.
                </p>
              </Field>

              <Field label="Environment">
                <input
                  value={form.environment}
                  onChange={(event) =>
                    updateField(
                      "environment",
                      event.target.value
                    )
                  }
                  placeholder="production"
                  className={inputClasses}
                />

                <p className="mt-2 text-xs text-zinc-600">
                  Leave blank to match any
                  environment.
                </p>
              </Field>
            </div>

            <div className="grid gap-5 md:grid-cols-[1fr_220px]">
              <Field label="Action">
                <input
                  required
                  value={form.action}
                  onChange={(event) =>
                    updateField(
                      "action",
                      event.target.value
                    )
                  }
                  placeholder="database.delete"
                  className={inputClasses}
                />
              </Field>

              <Field label="Action Matcher">
                <select
                  value={form.actionMatch}
                  onChange={(event) =>
                    updateField(
                      "actionMatch",
                      event.target
                        .value as MatchType
                    )
                  }
                  className={inputClasses}
                >
                  <option value="exact">
                    Exact
                  </option>

                  <option value="prefix">
                    Prefix
                  </option>

                  <option value="suffix">
                    Suffix
                  </option>
                </select>
              </Field>
            </div>

            <div className="grid gap-5 md:grid-cols-[1fr_220px]">
              <Field label="Resource">
                <input
                  required
                  value={form.resource}
                  onChange={(event) =>
                    updateField(
                      "resource",
                      event.target.value
                    )
                  }
                  placeholder="production/"
                  className={inputClasses}
                />
              </Field>

              <Field label="Resource Matcher">
                <select
                  value={form.resourceMatch}
                  onChange={(event) =>
                    updateField(
                      "resourceMatch",
                      event.target
                        .value as MatchType
                    )
                  }
                  className={inputClasses}
                >
                  <option value="exact">
                    Exact
                  </option>

                  <option value="prefix">
                    Prefix
                  </option>

                  <option value="suffix">
                    Suffix
                  </option>
                </select>
              </Field>
            </div>
          </div>
        </section>

        {error && (
          <div className="rounded-lg border border-red-500/20 bg-red-500/5 p-4 text-sm text-red-300">
            {error}
          </div>
        )}

        <div className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-zinc-800 bg-zinc-900/70 p-5">
          <div>
            <div className="text-sm font-medium text-zinc-300">
              Draft creation only
            </div>

            <div className="mt-1 text-xs text-zinc-600">
              Creating this policy will not
              change runtime authorization.
            </div>
          </div>

          <Button
            type="submit"
            disabled={submitting}
          >
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <ShieldPlus className="mr-2 h-4 w-4" />
            )}

            Create Draft Policy
          </Button>
        </div>
      </form>
    </div>
  )
}

const inputClasses =
  "w-full rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-sm text-zinc-200 outline-none transition placeholder:text-zinc-700 focus:border-zinc-600"

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-medium text-zinc-400">
        {label}
      </span>

      {children}
    </label>
  )
}