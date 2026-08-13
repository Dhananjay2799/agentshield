import Link from "next/link"

import {
  Activity,
  ArrowLeft,
  Bot,
  Boxes,
  BrainCircuit,
  CalendarClock,
  CircleDot,
  KeyRound,
  Server,
  ShieldAlert,
  UserRound,
  XCircle,
} from "lucide-react"

import {
  getAgent,
  getAgentActions,
  getAgentApprovals,
  getAgentSessionSecurity,
  getIncidents,
} from "@/lib/api/agents"

import type {
  ApprovalRequest,
  AuditEvent,
  Incident,
  SessionSecurity,
} from "@/types/agentshield"

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

import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"

function statusClasses(status: string) {
  switch (status.toLowerCase()) {
    case "active":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "revoked":
    case "suspended":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "expired":
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

function decisionClasses(decision: string) {
  switch (decision) {
    case "ALLOW":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "DENY":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "REQUIRE_APPROVAL":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

function approvalStatusClasses(status: string) {
  switch (status.toLowerCase()) {
    case "approved":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "denied":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "pending":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "expired":
      return "border-zinc-600 bg-zinc-800 text-zinc-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

function incidentStatusClasses(status: string) {
  switch (status.toLowerCase()) {
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

function getSessionStatus(
  session: SessionSecurity
) {
  if (session.status !== "active") {
    return session.status
  }

  if (
    session.expires_at &&
    new Date(session.expires_at).getTime() <
      Date.now()
  ) {
    return "expired"
  }

  return "active"
}

function riskClasses(score: number) {
  if (score >= 80) {
    return "text-red-400"
  }

  if (score >= 50) {
    return "text-amber-400"
  }

  if (score > 0) {
    return "text-emerald-400"
  }

  return "text-zinc-500"
}

export default async function AgentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  const [
    agent,
    sessions,
    actions,
    approvals,
    incidents,
  ] = await Promise.all([
    getAgent(id),
    getAgentSessionSecurity(id),
    getAgentActions(id),
    getAgentApprovals(id),
    getIncidents(),
  ])

  if (!agent) {
    return (
      <div className="p-6">
        <div className="mx-auto max-w-7xl">
          <Link
            href="/agents"
            className="inline-flex items-center gap-2 text-sm text-zinc-400 transition hover:text-white"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to agents
          </Link>

          <Card className="mt-6 border-zinc-800 bg-zinc-900/70">
            <CardContent className="p-8">
              <div className="text-lg font-medium">
                Agent not found.
              </div>

              <p className="mt-2 text-sm text-zinc-500">
                The requested agent could not be loaded.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  const agentIncidents =
    incidents.filter(
      (incident) =>
        incident.agent_id === agent.id
    )

  const activeIncidents =
    agentIncidents.filter(
      (incident) =>
        incident.status === "open" ||
        incident.status === "investigating"
    )

  const deniedActions =
    actions.filter(
      (event) =>
        event.decision === "DENY"
    )

  const approvalActions =
    actions.filter(
      (event) =>
        event.decision ===
        "REQUIRE_APPROVAL"
    )

  const allowedActions =
    actions.filter(
      (event) =>
        event.decision === "ALLOW"
    )

  const highestRisk =
    actions.reduce(
      (highest, event) =>
        Math.max(
          highest,
          event.risk_score
        ),
      0
    )

  const activeSessions =
    sessions.filter(
      (session) =>
        getSessionStatus(session) ===
        "active"
    )

  const pendingApprovals =
    approvals.filter(
      (approval) =>
        approval.status === "pending"
    )

  const totalSessionActions =
    sessions.reduce(
      (total, session) =>
        total + session.action_count,
      0
    )

  const totalSessionDenied =
    sessions.reduce(
      (total, session) =>
        total + session.denied_count,
      0
    )

  return (
    <div className="space-y-6 p-6">
      <div>
        <Link
          href="/agents"
          className="mb-5 inline-flex items-center gap-2 text-sm text-zinc-400 transition hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to agents
        </Link>

        <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div>
            <div className="mb-3 flex flex-wrap gap-2">
              <Badge
                variant="outline"
                className={statusClasses(
                  agent.status
                )}
              >
                {agent.status}
              </Badge>

              <Badge
                variant="outline"
                className={environmentClasses(
                  agent.environment
                )}
              >
                {agent.environment}
              </Badge>

              <Badge
                variant="outline"
                className="border-zinc-700 bg-zinc-900 text-zinc-300"
              >
                {agent.agent_type}
              </Badge>
            </div>

            <h2 className="text-3xl font-semibold tracking-tight">
              {agent.name}
            </h2>

            <p className="mt-2 text-sm text-zinc-500">
              Security posture, runtime activity,
              sessions, approvals, and policy
              decisions.
            </p>
          </div>

          <div className="font-mono text-xs text-zinc-500">
            {agent.id}
          </div>
        </div>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <MetricCard
          title="Risk Score"
          value={highestRisk || "—"}
          description="Highest observed action risk"
          icon={
            <Activity className="h-4 w-4 text-red-400" />
          }
        />

        <MetricCard
          title="Denied Actions"
          value={deniedActions.length}
          description="Blocked authorization attempts"
          icon={
            <XCircle className="h-4 w-4 text-red-400" />
          }
        />

        <MetricCard
          title="Approvals Required"
          value={approvalActions.length}
          description="Human escalation decisions"
          icon={
            <KeyRound className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Open Incidents"
          value={activeIncidents.length}
          description="Active security incidents"
          icon={
            <ShieldAlert className="h-4 w-4 text-zinc-500" />
          }
        />

        <MetricCard
          title="Active Sessions"
          value={activeSessions.length}
          description="Currently valid sessions"
          icon={
            <CircleDot className="h-4 w-4 text-emerald-400" />
          }
        />
      </section>

      <Tabs
        defaultValue="overview"
        className="space-y-5"
      >
        <TabsList className="border border-zinc-800 bg-zinc-900">
          <TabsTrigger value="overview">
            Overview
          </TabsTrigger>

          <TabsTrigger value="sessions">
            Sessions ({sessions.length})
          </TabsTrigger>

          <TabsTrigger value="actions">
            Action History ({actions.length})
          </TabsTrigger>

          <TabsTrigger value="approvals">
            Approvals ({approvals.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <OverviewTab
            agent={agent}
            incidents={agentIncidents}
            actions={actions}
            pendingApprovals={
              pendingApprovals.length
            }
            allowedActions={
              allowedActions.length
            }
            totalSessionActions={
              totalSessionActions
            }
            totalSessionDenied={
              totalSessionDenied
            }
          />
        </TabsContent>

        <TabsContent value="sessions">
          <SessionsTab
            sessions={sessions}
          />
        </TabsContent>

        <TabsContent value="actions">
          <ActionsTab
            actions={actions}
          />
        </TabsContent>

        <TabsContent value="approvals">
          <ApprovalsTab
            approvals={approvals}
          />
        </TabsContent>
      </Tabs>
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
  value: string | number
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

function OverviewTab({
  agent,
  incidents,
  actions,
  pendingApprovals,
  allowedActions,
  totalSessionActions,
  totalSessionDenied,
}: {
  agent: {
    name: string
    owner: string
    agent_type: string
    framework: string
    model: string
    environment: string
    status: string
    created_at: string
  }
  incidents: Incident[]
  actions: AuditEvent[]
  pendingApprovals: number
  allowedActions: number
  totalSessionActions: number
  totalSessionDenied: number
}) {
  return (
    <div className="grid gap-4 xl:grid-cols-2">
      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Agent Identity
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-5">
          <ProfileRow
            icon={
              <Bot className="h-4 w-4" />
            }
            label="Agent Name"
            value={agent.name}
          />

          <ProfileRow
            icon={
              <UserRound className="h-4 w-4" />
            }
            label="Owner"
            value={
              agent.owner || "—"
            }
          />

          <ProfileRow
            icon={
              <Boxes className="h-4 w-4" />
            }
            label="Agent Type"
            value={
              agent.agent_type || "—"
            }
          />

          <ProfileRow
            icon={
              <Boxes className="h-4 w-4" />
            }
            label="Framework"
            value={
              agent.framework || "—"
            }
          />

          <ProfileRow
            icon={
              <BrainCircuit className="h-4 w-4" />
            }
            label="Model"
            value={
              agent.model || "—"
            }
          />

          <ProfileRow
            icon={
              <Server className="h-4 w-4" />
            }
            label="Environment"
            value={
              agent.environment || "—"
            }
          />

          <ProfileRow
            icon={
              <CalendarClock className="h-4 w-4" />
            }
            label="Registered"
            value={new Date(
              agent.created_at
            ).toLocaleString()}
          />
        </CardContent>
      </Card>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Security Summary
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-4">
          <SummaryRow
            label="Total authorization decisions"
            value={actions.length}
          />

          <SummaryRow
            label="Allowed actions"
            value={allowedActions}
          />

          <SummaryRow
            label="Session action count"
            value={totalSessionActions}
          />

          <SummaryRow
            label="Session denied count"
            value={totalSessionDenied}
          />

          <SummaryRow
            label="Associated incidents"
            value={incidents.length}
          />

          <SummaryRow
            label="Pending approvals"
            value={pendingApprovals}
          />
        </CardContent>
      </Card>

      <Card className="border-zinc-800 bg-zinc-900/70 xl:col-span-2">
        <CardHeader>
          <CardTitle>
            Associated Incidents
          </CardTitle>
        </CardHeader>

        <CardContent>
          <IncidentTable
            incidents={incidents}
          />
        </CardContent>
      </Card>
    </div>
  )
}

function SessionsTab({
  sessions,
}: {
  sessions: SessionSecurity[]
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900/70">
      <CardHeader>
        <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
          <div>
            <CardTitle>
              Agent Sessions
            </CardTitle>

            <p className="mt-1 text-sm text-zinc-500">
              Execution sessions and their
              associated security activity.
            </p>
          </div>

          <Badge
            variant="outline"
            className="w-fit border-zinc-700 bg-zinc-900 text-zinc-300"
          >
            {sessions.length} Sessions
          </Badge>
        </div>
      </CardHeader>

      <CardContent>
        {sessions.length === 0 ? (
          <div className="flex min-h-32 items-center justify-center text-sm text-zinc-500">
            No sessions found.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead>
                    Task
                  </TableHead>

                  <TableHead>
                    Status
                  </TableHead>

                  <TableHead>
                    Actions
                  </TableHead>

                  <TableHead>
                    Allowed
                  </TableHead>

                  <TableHead>
                    Denied
                  </TableHead>

                  <TableHead>
                    Approvals
                  </TableHead>

                  <TableHead>
                    Max Risk
                  </TableHead>

                  <TableHead>
                    Last Activity
                  </TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                {sessions.map(
                  (session) => {
                    const status =
                      getSessionStatus(
                        session
                      )

                    return (
                      <TableRow
                        key={session.id}
                        className="border-zinc-800"
                      >
                        <TableCell>
                          <Link
                            href={`/agents/${session.agent_id}/sessions/${session.id}`}
                            className="font-medium text-zinc-200 transition hover:text-white"
                          >
                            {session.task_id}
                          </Link>

                          <div className="mt-1 font-mono text-xs text-zinc-500">
                            {session.id.slice(0, 12)}...
                          </div>
                        </TableCell>

                        <TableCell>
                          <Badge
                            variant="outline"
                            className={statusClasses(
                              status
                            )}
                          >
                            {status}
                          </Badge>
                        </TableCell>

                        <TableCell className="font-medium">
                          {
                            session.action_count
                          }
                        </TableCell>

                        <TableCell>
                          <span className="text-emerald-400">
                            {
                              session.allowed_count
                            }
                          </span>
                        </TableCell>

                        <TableCell>
                          <span
                            className={
                              session.denied_count >
                              0
                                ? "font-medium text-red-400"
                                : "text-zinc-500"
                            }
                          >
                            {
                              session.denied_count
                            }
                          </span>
                        </TableCell>

                        <TableCell>
                          <span
                            className={
                              session.approval_count >
                              0
                                ? "font-medium text-amber-400"
                                : "text-zinc-500"
                            }
                          >
                            {
                              session.approval_count
                            }
                          </span>
                        </TableCell>

                        <TableCell>
                          <span
                            className={`font-semibold ${riskClasses(
                              session.highest_risk_score
                            )}`}
                          >
                            {
                              session.highest_risk_score
                            }
                          </span>
                        </TableCell>

                        <TableCell className="whitespace-nowrap text-sm text-zinc-500">
                          {session.last_action_at
                            ? new Date(
                                session.last_action_at
                              ).toLocaleString()
                            : "No activity"}
                        </TableCell>
                      </TableRow>
                    )
                  }
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ActionsTab({
  actions,
}: {
  actions: AuditEvent[]
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900/70">
      <CardHeader>
        <CardTitle>
          Authorization Decision History
        </CardTitle>
      </CardHeader>

      <CardContent>
        <div className="space-y-3">
          {actions.map(
            (event) => (
              <div
                key={event.id}
                className="rounded-xl border border-zinc-800 bg-zinc-950/50 p-4"
              >
                <div className="flex flex-col justify-between gap-3 lg:flex-row lg:items-start">
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono text-sm font-medium text-zinc-200">
                        {
                          event.action
                        }
                      </span>

                      <Badge
                        variant="outline"
                        className={decisionClasses(
                          event.decision
                        )}
                      >
                        {
                          event.decision
                        }
                      </Badge>

                      <Badge
                        variant="outline"
                        className="border-zinc-700 bg-zinc-900 text-zinc-300"
                      >
                        Risk{" "}
                        {
                          event.risk_score
                        }
                      </Badge>
                    </div>

                    <div className="mt-2 font-mono text-xs text-zinc-500">
                      {
                        event.resource
                      }
                    </div>
                  </div>

                  <div className="text-xs text-zinc-500">
                    {new Date(
                      event.created_at
                    ).toLocaleString()}
                  </div>
                </div>

                <div className="mt-4 grid gap-3 lg:grid-cols-3">
                  <MetadataField
                    label="Risk Reason"
                    value={
                      event.metadata
                        ?.risk_reason
                    }
                  />

                  <MetadataField
                    label="Policy Reason"
                    value={
                      event.metadata
                        ?.policy_reason
                    }
                  />

                  <MetadataField
                    label="Request Reason"
                    value={
                      event.metadata
                        ?.request_reason
                    }
                  />
                </div>

                {(event.metadata
                  ?.approval_id ||
                  event.metadata
                    ?.grant_id) && (
                  <div className="mt-4 border-t border-zinc-800 pt-3">
                    <div className="grid gap-3 lg:grid-cols-2">
                      <MetadataField
                        label="Approval ID"
                        value={
                          event.metadata
                            ?.approval_id
                        }
                      />

                      <MetadataField
                        label="Grant ID"
                        value={
                          event.metadata
                            ?.grant_id
                        }
                      />
                    </div>
                  </div>
                )}
              </div>
            )
          )}

          {actions.length === 0 && (
            <div className="py-12 text-center text-sm text-zinc-500">
              No authorization decisions
              found.
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function ApprovalsTab({
  approvals,
}: {
  approvals: ApprovalRequest[]
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900/70">
      <CardHeader>
        <CardTitle>
          Human Approval History
        </CardTitle>
      </CardHeader>

      <CardContent>
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="border-zinc-800 hover:bg-transparent">
                <TableHead>
                  Action
                </TableHead>

                <TableHead>
                  Resource
                </TableHead>

                <TableHead>
                  Reason
                </TableHead>

                <TableHead>
                  Risk
                </TableHead>

                <TableHead>
                  Status
                </TableHead>

                <TableHead>
                  Requested
                </TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {approvals.map(
                (approval) => (
                  <TableRow
                    key={
                      approval.id
                    }
                    className="border-zinc-800"
                  >
                    <TableCell className="font-mono text-sm">
                      {
                        approval.action
                      }
                    </TableCell>

                    <TableCell className="font-mono text-xs text-zinc-500">
                      {
                        approval.resource
                      }
                    </TableCell>

                    <TableCell className="max-w-sm text-sm text-zinc-400">
                      {
                        approval.reason
                      }
                    </TableCell>

                    <TableCell>
                      {
                        approval.risk_score
                      }
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="outline"
                        className={approvalStatusClasses(
                          approval.status
                        )}
                      >
                        {
                          approval.status
                        }
                      </Badge>
                    </TableCell>

                    <TableCell className="whitespace-nowrap text-sm text-zinc-500">
                      {new Date(
                        approval.requested_at
                      ).toLocaleString()}
                    </TableCell>
                  </TableRow>
                )
              )}

              {approvals.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="h-32 text-center text-zinc-500"
                  >
                    No approval requests
                    found.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}

function IncidentTable({
  incidents,
}: {
  incidents: Incident[]
}) {
  if (incidents.length === 0) {
    return (
      <div className="py-10 text-center text-sm text-zinc-500">
        No incidents associated with this
        agent.
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow className="border-zinc-800 hover:bg-transparent">
            <TableHead>
              Incident
            </TableHead>

            <TableHead>
              Severity
            </TableHead>

            <TableHead>
              Events
            </TableHead>

            <TableHead>
              Status
            </TableHead>

            <TableHead>
              Last Seen
            </TableHead>

            <TableHead />
          </TableRow>
        </TableHeader>

        <TableBody>
          {incidents.map(
            (incident) => (
              <TableRow
                key={
                  incident.id
                }
                className="border-zinc-800"
              >
                <TableCell>
                  <div className="font-medium text-zinc-200">
                    {
                      incident.title
                    }
                  </div>

                  <div className="font-mono text-xs text-zinc-500">
                    {
                      incident.incident_type
                    }
                  </div>
                </TableCell>

                <TableCell>
                  <Badge
                    variant="outline"
                    className="border-red-500/30 bg-red-500/10 text-red-300"
                  >
                    {
                      incident.severity
                    }
                  </Badge>
                </TableCell>

                <TableCell>
                  {
                    incident.event_count
                  }
                </TableCell>

                <TableCell>
                  <Badge
                    variant="outline"
                    className={incidentStatusClasses(
                      incident.status
                    )}
                  >
                    {
                      incident.status
                    }
                  </Badge>
                </TableCell>

                <TableCell className="whitespace-nowrap text-sm text-zinc-500">
                  {new Date(
                    incident.last_seen_at
                  ).toLocaleString()}
                </TableCell>

                <TableCell className="text-right">
                  <Link
                    href={`/incidents/${incident.id}`}
                    className="text-sm font-medium text-zinc-300 transition hover:text-white"
                  >
                    Inspect →
                  </Link>
                </TableCell>
              </TableRow>
            )
          )}
        </TableBody>
      </Table>
    </div>
  )
}

function ProfileRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: string
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 text-zinc-500">
        {icon}
      </div>

      <div>
        <div className="text-xs uppercase tracking-wide text-zinc-500">
          {label}
        </div>

        <div className="mt-1 break-all text-sm text-zinc-200">
          {value}
        </div>
      </div>
    </div>
  )
}

function SummaryRow({
  label,
  value,
}: {
  label: string
  value: number
}) {
  return (
    <div className="flex items-center justify-between border-b border-zinc-800 pb-3 text-sm last:border-0">
      <span className="text-zinc-400">
        {label}
      </span>

      <span className="font-medium text-zinc-200">
        {value}
      </span>
    </div>
  )
}

function MetadataField({
  label,
  value,
}: {
  label: string
  value?: unknown
}) {
  if (
    value === undefined ||
    value === null ||
    value === ""
  ) {
    return null
  }

  return (
    <div>
      <div className="text-xs uppercase tracking-wide text-zinc-600">
        {label}
      </div>

      <div className="mt-1 text-sm text-zinc-400">
        {String(value)}
      </div>
    </div>
  )
}