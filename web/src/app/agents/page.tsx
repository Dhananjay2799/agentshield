import Link from "next/link"

import {
  Bot,
  Boxes,
  BrainCircuit,
  CircleDot,
  Server,
  ShieldCheck,
  UserRound,
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
  switch (status.toLowerCase()) {
    case "active":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "suspended":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "inactive":
      return "border-zinc-600 bg-zinc-800 text-zinc-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

function environmentClasses(environment: string) {
  switch (environment.toLowerCase()) {
    case "production":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "staging":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "development":
      return "border-blue-500/30 bg-blue-500/10 text-blue-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

export default async function AgentsPage() {
  const agents = await getAgents()

  const activeAgents = agents.filter(
    (agent) => agent.status === "active"
  )

  const productionAgents = agents.filter(
    (agent) => agent.environment === "production"
  )

  const frameworks = new Set(
    agents
      .map((agent) => agent.framework)
      .filter(Boolean)
  )

  return (
    <div className="space-y-6 p-6">

      <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">
            Agent Inventory
          </h2>

          <p className="mt-1 text-sm text-zinc-500">
            Registered autonomous agents operating through the AgentShield
            security control plane.
          </p>
        </div>

        <Badge
          variant="outline"
          className="w-fit border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
        >
          {agents.length} Registered
        </Badge>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Registered Agents
            </CardTitle>

            <Bot className="h-4 w-4 text-zinc-500" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {agents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Total onboarded agents
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Active Agents
            </CardTitle>

            <CircleDot className="h-4 w-4 text-emerald-400" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {activeAgents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Currently enabled
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Production Agents
            </CardTitle>

            <Server className="h-4 w-4 text-red-400" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {productionAgents.length}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Production environment
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-zinc-300">
              Frameworks
            </CardTitle>

            <Boxes className="h-4 w-4 text-zinc-500" />
          </CardHeader>

          <CardContent>
            <div className="text-3xl font-semibold">
              {frameworks.size}
            </div>

            <p className="mt-1 text-xs text-zinc-500">
              Distinct agent frameworks
            </p>
          </CardContent>
        </Card>

      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">
              Registered Agents
            </CardTitle>

            <ShieldCheck className="h-4 w-4 text-zinc-500" />
          </div>
        </CardHeader>

        <CardContent>
          {agents.length === 0 ? (
            <div className="flex min-h-40 flex-col items-center justify-center text-center">
              <Bot className="mb-3 h-8 w-8 text-zinc-600" />

              <p className="text-sm font-medium text-zinc-300">
                No agents registered
              </p>

              <p className="mt-1 text-xs text-zinc-500">
                Registered AgentShield agents will appear here.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead>Agent</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Framework</TableHead>
                  <TableHead>Model</TableHead>
                  <TableHead>Environment</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>

              <TableBody>
                {agents.map((agent) => (
                  <TableRow
                    key={agent.id}
                    className="border-zinc-800"
                  >
                    <TableCell>
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-950">
                          <Bot className="h-4 w-4 text-zinc-400" />
                        </div>

                        <div>
                          <div className="font-medium text-zinc-200">
                            {agent.name || "Unnamed Agent"}
                          </div>

                          <div className="font-mono text-xs text-zinc-500">
                            {agent.id.slice(0, 12)}...
                          </div>

                          {agent.owner && (
                            <div className="mt-1 flex items-center gap-1 text-xs text-zinc-600">
                              <UserRound className="h-3 w-3" />
                              {agent.owner}
                            </div>
                          )}
                        </div>
                      </div>
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="outline"
                        className="border-zinc-700 bg-zinc-900 text-zinc-300"
                      >
                        {agent.agent_type || "unknown"}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <div className="flex items-center gap-2 text-sm text-zinc-300">
                        <Boxes className="h-3.5 w-3.5 text-zinc-500" />
                        {agent.framework || "—"}
                      </div>
                    </TableCell>

                    <TableCell>
                      <div className="flex items-center gap-2 text-sm text-zinc-300">
                        <BrainCircuit className="h-3.5 w-3.5 text-zinc-500" />
                        {agent.model || "—"}
                      </div>
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="outline"
                        className={environmentClasses(
                          agent.environment || ""
                        )}
                      >
                        {agent.environment || "unknown"}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="outline"
                        className={statusClasses(
                          agent.status || ""
                        )}
                      >
                        {agent.status || "unknown"}
                      </Badge>
                    </TableCell>

                    <TableCell className="text-right">
                      <Link
                        href={`/agents/${agent.id}`}
                        className="text-sm font-medium text-zinc-300 transition hover:text-white"
                      >
                        View →
                      </Link>
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