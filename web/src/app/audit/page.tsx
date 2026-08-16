import {
  Activity,
  Bot,
  ShieldCheck,
} from "lucide-react"

import {
  getRecentEvents,
} from "@/lib/api/events"

import {
  AuditExplorer,
} from "@/components/audit/AuditExplorer"

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export default async function AuditPage() {
  const events =
    await getRecentEvents(
      100
    )

  const policyEvents =
    events.filter(
      (event) =>
        event.event_type.startsWith(
          "policy."
        )
    ).length

  const agentEvents =
    events.length -
    policyEvents

  const deniedEvents =
    events.filter(
      (event) =>
        event.decision ===
        "DENY"
    ).length

  return (
    <div className="space-y-6 p-6">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">
          Audit Logs
        </h2>

        <p className="mt-2 text-sm text-zinc-500">
          Search and investigate
          autonomous-agent security
          decisions and policy
          control-plane lifecycle
          activity.
        </p>
      </div>

      <section className="grid gap-4 md:grid-cols-3">
        <MetricCard
          title="Agent Events"
          value={
            agentEvents
          }
          description="Runtime security activity"
          icon={
            <Bot className="h-4 w-4 text-zinc-400" />
          }
        />

        <MetricCard
          title="Policy Events"
          value={
            policyEvents
          }
          description="Control-plane lifecycle activity"
          icon={
            <ShieldCheck className="h-4 w-4 text-sky-400" />
          }
        />

        <MetricCard
          title="Denied"
          value={
            deniedEvents
          }
          description="Blocked runtime actions"
          icon={
            <Activity className="h-4 w-4 text-red-400" />
          }
        />
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Audit Explorer
          </CardTitle>

          <p className="text-sm text-zinc-500">
            Filter by source,
            decision, risk, or
            search across actions,
            resources, policies,
            agents, and sessions.
          </p>
        </CardHeader>

        <CardContent>
          <AuditExplorer
            events={events}
          />
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