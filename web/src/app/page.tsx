import Link from "next/link"

import {
  Activity,
  AlertTriangle,
  Bot,
  FileClock,
  Gauge,
  ShieldCheck,
  Siren,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

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
}

type Agent = {
  id: string
  name: string
  agent_type: string
  owner: string
  framework: string
  model: string
  environment: string
  status: string
}

async function getIncidents(): Promise<Incident[]> {
  try {
    const response = await fetch(
      "http://localhost:8080/v1/incidents",
      {
        cache: "no-store",
      }
    )

    if (!response.ok) {
      return []
    }

    return response.json()
  } catch {
    return []
  }
}

async function getAgents(): Promise<Agent[]> {
  try {
    const response = await fetch(
      "http://localhost:8080/v1/agents",
      {
        cache: "no-store",
      }
    )

    if (!response.ok) {
      return []
    }

    return response.json()
  } catch {
    return []
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

function severityClasses(severity: string) {
  switch (severity) {
    case "critical":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "high":
      return "border-orange-500/30 bg-orange-500/10 text-orange-300"

    case "medium":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "low":
      return "border-blue-500/30 bg-blue-500/10 text-blue-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

export default async function Home() {
  const [incidents, agents] = await Promise.all([
    getIncidents(),
    getAgents(),
  ])

  const openIncidents = incidents.filter(
    (incident) =>
      incident.status === "open" ||
      incident.status === "investigating"
  )

  const criticalIncidents = openIncidents.filter(
    (incident) =>
      incident.severity === "critical"
  )

  const activeAgents = agents.filter(
    (agent) =>
      agent.status === "active"
  )

  const recentIncidents = incidents.slice(0, 8)

  return (
    <div className="space-y-6 p-6">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">
          Overview
        </h2>

        <p className="mt-1 text-sm text-zinc-500">
          Real-time security posture across registered AI agents.
        </p>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Active Agents
            </CardTitle>

            <Bot className="h-4 w-4 text-zinc-500" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {activeAgents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Registered autonomous agents
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Open Incidents
            </CardTitle>

            <AlertTriangle className="h-4 w-4 text-zinc-500" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {openIncidents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Awaiting analyst action
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Critical Incidents
            </CardTitle>

            <Siren className="h-4 w-4 text-red-400" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {criticalIncidents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Immediate attention required
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Policy Engine
            </CardTitle>

            <ShieldCheck className="h-4 w-4 text-zinc-500" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              OPA
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Rego enforcement active
            </p>
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-4 xl:grid-cols-[1.5fr_1fr]">
        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">
                Threat Activity
              </CardTitle>

              <Gauge className="h-4 w-4 text-zinc-500" />
            </div>
          </CardHeader>

          <CardContent>
            <div className="flex h-52 items-center justify-center rounded-lg border border-dashed border-zinc-800 bg-zinc-950/50">
              <div className="text-center">
                <Activity className="mx-auto mb-3 h-7 w-7 text-zinc-600" />

                <p className="text-sm text-zinc-500">
                  Risk activity chart will appear here
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle className="text-base">
              Security Pipeline
            </CardTitle>
          </CardHeader>

          <CardContent className="space-y-4">
            {[
              ["Gateway", "Online"],
              ["OPA / Rego", "Enforcing"],
              ["Kafka", "Streaming"],
              ["Detection", "Monitoring"],
              ["PostgreSQL", "Healthy"],
            ].map(([name, state]) => (
              <div
                key={name}
                className="flex items-center justify-between border-b border-zinc-800 pb-3 text-sm last:border-0"
              >
                <span className="text-zinc-400">
                  {name}
                </span>

                <span className="text-emerald-400">
                  {state}
                </span>
              </div>
            ))}
          </CardContent>
        </Card>
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">
              Recent Security Incidents
            </CardTitle>

            <div className="flex items-center gap-3">
              <Link
                href="/incidents"
                className="text-xs text-zinc-400 transition hover:text-white"
              >
                View all
              </Link>

              <FileClock className="h-4 w-4 text-zinc-500" />
            </div>
          </div>
        </CardHeader>

        <CardContent>
          {recentIncidents.length === 0 ? (
            <div className="flex min-h-32 items-center justify-center text-sm text-zinc-500">
              No security incidents found.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead>Severity</TableHead>
                  <TableHead>Incident</TableHead>
                  <TableHead>Agent</TableHead>
                  <TableHead>Events</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Seen</TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                {recentIncidents.map((incident) => (
                  <TableRow
                    key={incident.id}
                    className="border-zinc-800"
                  >
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={severityClasses(
                          incident.severity
                        )}
                      >
                        {incident.severity}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <Link
                        href={`/incidents/${incident.id}`}
                        className="font-medium text-zinc-200 transition hover:text-white"
                      >
                        {incident.title}
                      </Link>

                      <div className="font-mono text-xs text-zinc-500">
                        {incident.incident_type}
                      </div>
                    </TableCell>

                    <TableCell className="font-mono text-xs text-zinc-400">
                      {incident.agent_id.slice(0, 8)}...
                    </TableCell>

                    <TableCell>
                      {incident.event_count}
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="outline"
                        className={statusClasses(
                          incident.status
                        )}
                      >
                        {incident.status}
                      </Badge>
                    </TableCell>

                    <TableCell className="text-sm text-zinc-500">
                      {new Date(
                        incident.last_seen_at
                      ).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}