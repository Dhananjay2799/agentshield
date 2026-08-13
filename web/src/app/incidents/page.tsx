import Link from "next/link"
import {
  AlertTriangle,
  ArrowLeft,
  Clock,
  ShieldAlert,
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

export default async function IncidentsPage() {
  const incidents = await getIncidents()

  const open = incidents.filter(
    (incident) =>
      incident.status === "open" ||
      incident.status === "investigating"
  )

  const critical = open.filter(
    (incident) => incident.severity === "critical"
  )

  return (
    <main className="min-h-screen bg-zinc-950 p-6 text-zinc-100">
      <div className="mx-auto max-w-7xl space-y-6">

        <div>
          <Link
            href="/"
            className="mb-5 inline-flex items-center gap-2 text-sm text-zinc-400 transition hover:text-white"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to SOC Overview
          </Link>

          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
                Security Incidents
              </h1>

              <p className="mt-1 text-sm text-zinc-500">
                Investigate and manage detected autonomous-agent threats.
              </p>
            </div>

            <Badge
              variant="outline"
              className="border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
            >
              {incidents.length} Total
            </Badge>
          </div>
        </div>

        <section className="grid gap-4 md:grid-cols-3">
          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Total Incidents
              </CardTitle>

              <ShieldAlert className="h-4 w-4 text-zinc-500" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {incidents.length}
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Active
              </CardTitle>

              <Clock className="h-4 w-4 text-amber-400" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {open.length}
              </div>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/70">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm text-zinc-400">
                Critical Active
              </CardTitle>

              <AlertTriangle className="h-4 w-4 text-red-400" />
            </CardHeader>

            <CardContent>
              <div className="text-3xl font-semibold">
                {critical.length}
              </div>
            </CardContent>
          </Card>
        </section>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Incident Queue</CardTitle>
          </CardHeader>

          <CardContent>
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead>Severity</TableHead>
                  <TableHead>Incident</TableHead>
                  <TableHead>Agent</TableHead>
                  <TableHead>Events</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Seen</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>

              <TableBody>
                {incidents.map((incident) => (
                  <TableRow
                    key={incident.id}
                    className="border-zinc-800"
                  >
                    <TableCell>
                      <Badge className="border border-red-500/30 bg-red-500/10 text-red-300">
                        {incident.severity}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <div className="font-medium">
                        {incident.title}
                      </div>

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
                        className={statusClasses(incident.status)}
                      >
                        {incident.status}
                      </Badge>
                    </TableCell>

                    <TableCell className="text-sm text-zinc-500">
                      {new Date(
                        incident.last_seen_at
                      ).toLocaleString()}
                    </TableCell>

                    <TableCell className="text-right">
                      <Link
                        href={`/incidents/${incident.id}`}
                        className="text-sm font-medium text-zinc-300 hover:text-white"
                      >
                        Investigate →
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}

                {incidents.length === 0 && (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className="h-32 text-center text-zinc-500"
                    >
                      No incidents detected.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}