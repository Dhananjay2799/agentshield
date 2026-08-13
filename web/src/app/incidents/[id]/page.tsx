import Link from "next/link"
import {
  Activity,
  ArrowLeft,
  Bot,
  CalendarClock,
  Database,
  Fingerprint,
  ShieldAlert,
  TimerReset,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import IncidentActions from "@/components/incidents/IncidentActions"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type IncidentMetadata = {
  action?: string
  resource?: string
  risk_score?: number
  event_count_window?: number
  [key: string]: unknown
}

type Incident = {
  id: string
  agent_id: string
  session_id?: string
  incident_type: string
  severity: string
  title: string
  description?: string
  status: string
  event_count: number
  first_seen_at: string
  last_seen_at: string
  created_at: string
  resolved_at?: string
  metadata?: IncidentMetadata
}

async function getIncident(id: string): Promise<Incident | null> {
  try {
    const response = await fetch(
      `http://localhost:8080/v1/incidents/${id}`,
      {
        cache: "no-store",
      }
    )

    if (!response.ok) {
      return null
    }

    return response.json()
  } catch {
    return null
  }
}

function statusClasses(status: string) {
  switch (status) {
    case "open":
      return "border-red-500/30 bg-red-500/10 text-red-300"
    case "investigating":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"
    case "resolved":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
    case "dismissed":
      return "border-zinc-600 bg-zinc-800 text-zinc-300"
    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

export default async function IncidentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const incident = await getIncident(id)

  if (!incident) {
    return (
      <main className="min-h-screen bg-zinc-950 p-6 text-zinc-100">
        <div className="mx-auto max-w-5xl">
          <Link
            href="/incidents"
            className="inline-flex items-center gap-2 text-sm text-zinc-400 hover:text-white"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to incidents
          </Link>

          <div className="mt-10 rounded-xl border border-zinc-800 bg-zinc-900 p-8">
            Incident not found.
          </div>
        </div>
      </main>
    )
  }

  const metadata = incident.metadata ?? {}

  return (
    <main className="min-h-screen bg-zinc-950 p-6 text-zinc-100">
      <div className="mx-auto max-w-7xl space-y-6">

        <div>
          <Link
            href="/incidents"
            className="mb-5 inline-flex items-center gap-2 text-sm text-zinc-400 hover:text-white"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to incidents
          </Link>

          <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
            <div>
              <div className="mb-3 flex items-center gap-2">
                <Badge className="border border-red-500/30 bg-red-500/10 text-red-300">
                  {incident.severity}
                </Badge>

                <Badge
                  variant="outline"
                  className={statusClasses(incident.status)}
                >
                  {incident.status}
                </Badge>
              </div>

              <h1 className="text-3xl font-semibold tracking-tight">
                {incident.title}
              </h1>

              <p className="mt-2 max-w-3xl text-sm text-zinc-400">
                {incident.description}
              </p>
            </div>

            <div className="font-mono text-xs text-zinc-500">
              {incident.id}
            </div>
          </div>
        </div>

        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Risk Score
              </CardTitle>
              <Activity className="h-4 w-4 text-red-400" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {metadata.risk_score ?? "—"}
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Event Count
              </CardTitle>
              <ShieldAlert className="h-4 w-4 text-zinc-500" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {incident.event_count}
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Window Count
              </CardTitle>
              <TimerReset className="h-4 w-4 text-zinc-500" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {metadata.event_count_window ?? "—"}
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Incident Type
              </CardTitle>
              <Fingerprint className="h-4 w-4 text-zinc-500" />
            </CardHeader>

            <CardContent>
              <div className="break-all font-mono text-sm">
                {incident.incident_type}
              </div>
            </CardContent>
          </Card>
        </section>

        <section className="grid gap-4 xl:grid-cols-2">
          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader>
              <CardTitle>Threat Evidence</CardTitle>
            </CardHeader>

            <CardContent className="space-y-5">
              <div className="flex items-start gap-3">
                <Activity className="mt-0.5 h-4 w-4 text-zinc-500" />
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500">
                    Action
                  </div>
                  <div className="mt-1 font-mono text-sm">
                    {metadata.action ?? "Unknown"}
                  </div>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Database className="mt-0.5 h-4 w-4 text-zinc-500" />
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500">
                    Resource
                  </div>
                  <div className="mt-1 font-mono text-sm">
                    {metadata.resource ?? "Unknown"}
                  </div>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Bot className="mt-0.5 h-4 w-4 text-zinc-500" />
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500">
                    Agent ID
                  </div>
                  <div className="mt-1 break-all font-mono text-sm">
                    {incident.agent_id}
                  </div>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Fingerprint className="mt-0.5 h-4 w-4 text-zinc-500" />
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500">
                    Session ID
                  </div>
                  <div className="mt-1 break-all font-mono text-sm">
                    {incident.session_id ?? "Unavailable"}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader>
              <CardTitle>Incident Timeline</CardTitle>
            </CardHeader>

            <CardContent className="space-y-5">
              <TimelineItem
                label="Created"
                value={incident.created_at}
              />

              <TimelineItem
                label="First Seen"
                value={incident.first_seen_at}
              />

              <TimelineItem
                label="Last Seen"
                value={incident.last_seen_at}
              />

              {incident.resolved_at && (
                <TimelineItem
                  label={
                    incident.status === "dismissed"
                      ? "Dismissed"
                      : "Resolved"
                  }
                  value={incident.resolved_at}
                />
              )}
            </CardContent>
          </Card>
        </section>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Analyst Response</CardTitle>
          </CardHeader>

          <CardContent>
            <IncidentActions
              incidentId={incident.id}
              status={incident.status}
            />
          </CardContent>
        </Card>
      </div>
    </main>
  )
}

function TimelineItem({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="flex gap-3">
      <CalendarClock className="mt-0.5 h-4 w-4 text-zinc-500" />

      <div>
        <div className="text-xs uppercase tracking-wide text-zinc-500">
          {label}
        </div>

        <div className="mt-1 text-sm text-zinc-300">
          {new Date(value).toLocaleString()}
        </div>
      </div>
    </div>
  )
}