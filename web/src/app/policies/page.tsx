import Link from "next/link"

import {
  Plus,
  ShieldCheck,
  ShieldOff,
  ShieldQuestion,
} from "lucide-react"

import {
  getPolicies,
  type Policy,
} from "@/lib/api/policies"

import { Badge } from "@/components/ui/badge"

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function statusClasses(
  status: Policy["status"]
) {
  switch (status) {
    case "active":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "draft":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "disabled":
      return "border-zinc-700 bg-zinc-900 text-zinc-400"

    case "archived":
      return "border-zinc-800 bg-zinc-950 text-zinc-600"
  }
}

function effectClasses(
  effect: Policy["effect"]
) {
  switch (effect) {
    case "ALLOW":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "REQUIRE_APPROVAL":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "DENY":
      return "border-red-500/30 bg-red-500/10 text-red-300"
  }
}

export default async function PoliciesPage() {
  const policies =
    await getPolicies()

  const activeCount =
    policies.filter(
      (policy) =>
        policy.status === "active"
    ).length

  const draftCount =
    policies.filter(
      (policy) =>
        policy.status === "draft"
    ).length

  const disabledCount =
    policies.filter(
      (policy) =>
        policy.status === "disabled"
    ).length

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
  	<div>
    	  <h2 className="text-2xl font-semibold tracking-tight">
      	    Policies
    	  </h2>

    	  <p className="mt-2 text-sm text-zinc-500">
      	     Manage autonomous-agent authorization
      	     rules synchronized with OPA.
    	  </p>
  	</div>

  	<Link
  	  href="/policies/new"
  	  className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-xs transition-all hover:bg-primary/90"
	>
  	  <Plus className="mr-2 h-4 w-4" />
  	  Create Policy
	</Link>
   </div>

      <section className="grid gap-4 md:grid-cols-3">
        <MetricCard
          title="Active"
          value={activeCount}
          description="Currently enforced"
          icon={
            <ShieldCheck className="h-4 w-4 text-emerald-400" />
          }
        />

        <MetricCard
          title="Draft"
          value={draftCount}
          description="Awaiting validation or activation"
          icon={
            <ShieldQuestion className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Disabled"
          value={disabledCount}
          description="Stored but not enforced"
          icon={
            <ShieldOff className="h-4 w-4 text-zinc-400" />
          }
        />
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Managed Policies
          </CardTitle>
        </CardHeader>

        <CardContent>
          {policies.length === 0 ? (
            <div className="py-16 text-center text-sm text-zinc-500">
              No managed policies found.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[1000px] text-left text-sm">
                <thead className="border-b border-zinc-800 text-xs uppercase tracking-wide text-zinc-600">
                  <tr>
                    <th className="px-3 py-3">
                      Policy
                    </th>
                    <th className="px-3 py-3">
                      Effect
                    </th>
                    <th className="px-3 py-3">
                      Status
                    </th>
                    <th className="px-3 py-3">
                      Action
                    </th>
                    <th className="px-3 py-3">
                      Resource
                    </th>
                    <th className="px-3 py-3">
                      Environment
                    </th>
                    <th className="px-3 py-3">
                      Priority
                    </th>
                    <th className="px-3 py-3">
                      Version
                    </th>
                  </tr>
                </thead>

                <tbody>
                  {policies.map(
                    (policy) => (
                      <tr
                        key={policy.id}
                        className="border-b border-zinc-900 transition hover:bg-zinc-900/60"
                      >
                        <td className="px-3 py-4">
                          <Link
                            href={`/policies/${policy.id}`}
                            className="font-medium text-zinc-200 transition hover:text-white"
                          >
                            {
                              policy.name
                            }
                          </Link>

                          <div className="mt-1 max-w-sm text-xs text-zinc-600">
                            {
                              policy.description
                            }
                          </div>
                        </td>

                        <td className="px-3 py-4">
                          <Badge
                            variant="outline"
                            className={effectClasses(
                              policy.effect
                            )}
                          >
                            {
                              policy.effect
                            }
                          </Badge>
                        </td>

                        <td className="px-3 py-4">
                          <Badge
                            variant="outline"
                            className={statusClasses(
                              policy.status
                            )}
                          >
                            {
                              policy.status.toUpperCase()
                            }
                          </Badge>
                        </td>

                        <td className="px-3 py-4 font-mono text-xs text-zinc-400">
                          {policy.action}
                        </td>

                        <td className="px-3 py-4 font-mono text-xs text-zinc-400">
                          {policy.resource}
                        </td>

                        <td className="px-3 py-4 text-zinc-400">
                          {policy.environment ??
                            "Any"}
                        </td>

                        <td className="px-3 py-4 text-zinc-400">
                          {
                            policy.priority
                          }
                        </td>

                        <td className="px-3 py-4 text-zinc-400">
                          v{
                            policy.version
                          }
                        </td>
                      </tr>
                    )
                  )}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function MetricCard({
  title,
  value,
  description,
  icon,
}: {
  title: string
  value: number
  description: string
  icon: React.ReactNode
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900/70">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-zinc-300">
          {title}
        </CardTitle>

        {icon}
      </CardHeader>

      <CardContent>
        <div className="text-3xl font-semibold">
          {value}
        </div>

        <p className="mt-1 text-xs text-zinc-500">
          {description}
        </p>
      </CardContent>
    </Card>
  )
}