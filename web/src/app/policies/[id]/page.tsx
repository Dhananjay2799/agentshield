import Link from "next/link"
import { notFound } from "next/navigation"
import { PolicyLifecycleActions } from "@/components/policies/PolicyLifecycleActions"
import {
  ArrowLeft,
  Bot,
  Database,
  FileCode2,
  Gauge,
  Globe2,
  ShieldCheck,
} from "lucide-react"

import {
  getPolicy,
  type Policy,
} from "@/lib/api/policies"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function statusClasses(status: Policy["status"]) {
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

function effectClasses(effect: Policy["effect"]) {
  switch (effect) {
    case "ALLOW":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
    case "REQUIRE_APPROVAL":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"
    case "DENY":
      return "border-red-500/30 bg-red-500/10 text-red-300"
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(new Date(value))
}

export default async function PolicyDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  const policy = await getPolicy(id)

  if (!policy) {
    notFound()
  }

  return (
    <div className="space-y-6 p-6">
      <div>
        <Link
          href="/policies"
          className="inline-flex items-center gap-2 text-sm text-zinc-500 transition hover:text-zinc-200"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Policies
        </Link>

        <div className="mt-5 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-2xl font-semibold tracking-tight">
                {policy.name}
              </h2>

              <Badge
                variant="outline"
                className={statusClasses(policy.status)}
              >
                {policy.status.toUpperCase()}
              </Badge>

              <Badge
                variant="outline"
                className={effectClasses(policy.effect)}
              >
                {policy.effect}
              </Badge>
            </div>

            <p className="mt-3 max-w-3xl text-sm leading-6 text-zinc-500">
              {policy.description}
            </p>
          </div>

          <div className="text-right text-xs text-zinc-600">
            <div>Policy ID</div>
            <div className="mt-1 font-mono text-zinc-400">
              {policy.id}
            </div>
          </div>
        </div>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <InfoCard
          title="Effect"
          value={policy.effect}
          icon={<ShieldCheck className="h-4 w-4" />}
        />

        <InfoCard
          title="Priority"
          value={String(policy.priority)}
          icon={<Gauge className="h-4 w-4" />}
        />

        <InfoCard
          title="Environment"
          value={policy.environment || "Any"}
          icon={<Globe2 className="h-4 w-4" />}
        />

        <InfoCard
          title="Version"
          value={`v${policy.version}`}
          icon={<FileCode2 className="h-4 w-4" />}
        />
      </section>

      <div className="grid gap-6 xl:grid-cols-[1.4fr_1fr]">
        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Policy Match Conditions</CardTitle>
          </CardHeader>

          <CardContent className="space-y-5">
            <Condition
              label="Agent Type"
              value={policy.agent_type || "Any agent"}
              icon={<Bot className="h-4 w-4" />}
            />

            <Condition
              label="Action"
              value={policy.action}
              secondary={`Matcher: ${policy.action_match}`}
              icon={<FileCode2 className="h-4 w-4" />}
            />

            <Condition
              label="Resource"
              value={policy.resource}
              secondary={`Matcher: ${policy.resource_match}`}
              icon={<Database className="h-4 w-4" />}
            />

            <Condition
              label="Environment"
              value={policy.environment || "Any"}
              icon={<Globe2 className="h-4 w-4" />}
            />
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Policy Metadata</CardTitle>
          </CardHeader>

          <CardContent className="space-y-4 text-sm">
            <MetadataRow
              label="Status"
              value={policy.status}
            />

            <MetadataRow
              label="Source"
              value={policy.source}
            />

            <MetadataRow
              label="Created By"
              value={policy.created_by}
            />

            <MetadataRow
              label="Version"
              value={`v${policy.version}`}
            />

            <MetadataRow
              label="Created"
              value={formatDate(policy.created_at)}
            />

            <MetadataRow
              label="Last Updated"
              value={formatDate(policy.updated_at)}
            />
          </CardContent>
        </Card>
      </div>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>Lifecycle Controls</CardTitle>
        </CardHeader>

        <CardContent>
  		<p className="mb-5 text-sm text-zinc-500">
    		 Validate policy configuration, synchronize active
    		 policies with OPA, or remove policies from runtime
    		 enforcement.
  		</p>

  		<PolicyLifecycleActions
    		 policyId={policy.id}
    		 status={policy.status}
  		/>
	</CardContent>
      </Card>
    </div>
  )
}

function InfoCard({
  title,
  value,
  icon,
}: {
  title: string
  value: string
  icon: React.ReactNode
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900/70">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-zinc-400">
          {title}
        </CardTitle>

        <div className="text-zinc-600">
          {icon}
        </div>
      </CardHeader>

      <CardContent>
        <div className="break-words font-mono text-sm text-zinc-200">
          {value}
        </div>
      </CardContent>
    </Card>
  )
}

function Condition({
  label,
  value,
  secondary,
  icon,
}: {
  label: string
  value: string
  secondary?: string
  icon: React.ReactNode
}) {
  return (
    <div className="flex gap-3 rounded-lg border border-zinc-800 bg-zinc-950/40 p-4">
      <div className="mt-0.5 text-zinc-600">
        {icon}
      </div>

      <div>
        <div className="text-xs uppercase tracking-wide text-zinc-600">
          {label}
        </div>

        <div className="mt-1 font-mono text-sm text-zinc-200">
          {value}
        </div>

        {secondary && (
          <div className="mt-1 text-xs text-zinc-600">
            {secondary}
          </div>
        )}
      </div>
    </div>
  )
}

function MetadataRow({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-zinc-800 pb-3 last:border-0">
      <span className="text-zinc-600">
        {label}
      </span>

      <span className="max-w-[65%] break-words text-right text-zinc-300">
        {value}
      </span>
    </div>
  )
}