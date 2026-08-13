import Link from "next/link"

import {
  Activity,
  ArrowLeft,
  CheckCircle2,
  Clock3,
  KeyRound,
  ShieldAlert,
  XCircle,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type AgentSession = {
  id: string
  agent_id: string
  task_id: string
  status: string
  started_at: string
  ended_at?: string
  expires_at?: string
}

type AuditMetadata = {
  grant_id?: string
  agent_type?: string
  approval_id?: string
  environment?: string
  risk_reason?: string
  policy_engine?: string
  policy_reason?: string
  request_reason?: string
  [key: string]: unknown
}

type AuditEvent = {
  id: string
  agent_id: string
  session_id: string
  event_type: string
  action: string
  resource: string
  decision: string
  risk_score: number
  metadata: AuditMetadata
  created_at: string
}

type ApprovalRequest = {
  id: string
  agent_id: string
  session_id: string
  action: string
  resource: string
  reason: string
  risk_score: number
  status: string
  requested_at: string
  approved_at?: string
  denied_at?: string
  expires_at?: string
}

async function getSession(
  sessionId: string
): Promise<AgentSession | null> {
  try {
    const response = await fetch(
      `http://localhost:8080/v1/sessions/${sessionId}`,
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

async function getSessionActions(
  sessionId: string
): Promise<AuditEvent[]> {
  try {
    const response = await fetch(
      `http://localhost:8080/v1/sessions/${sessionId}/actions`,
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

async function getSessionApprovals(
  sessionId: string
): Promise<ApprovalRequest[]> {
  try {
    const response = await fetch(
      `http://localhost:8080/v1/sessions/${sessionId}/approvals`,
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

function sessionStatus(session: AgentSession) {
  if (session.status !== "active") {
    return session.status
  }

  if (
    session.expires_at &&
    new Date(session.expires_at).getTime() < Date.now()
  ) {
    return "expired"
  }

  return "active"
}

function sessionStatusClasses(status: string) {
  switch (status) {
    case "active":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "revoked":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "expired":
      return "border-zinc-600 bg-zinc-800 text-zinc-300"

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
  switch (status) {
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

export default async function SessionDetailPage({
  params,
}: {
  params: Promise<{
    id: string
    sessionId: string
  }>
}) {
  const { id, sessionId } = await params

  const [session, actions, approvals] = await Promise.all([
    getSession(sessionId),
    getSessionActions(sessionId),
    getSessionApprovals(sessionId),
  ])

  if (!session) {
    return (
      <div className="p-6">
        <Link
          href={`/agents/${id}`}
          className="inline-flex items-center gap-2 text-sm text-zinc-400 hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to agent
        </Link>

        <Card className="mt-6 border-zinc-800 bg-zinc-900/70">
          <CardContent className="p-8">
            Session not found.
          </CardContent>
        </Card>
      </div>
    )
  }

  const status = sessionStatus(session)

  const allowedCount = actions.filter(
    (event) => event.decision === "ALLOW"
  ).length

  const deniedCount = actions.filter(
    (event) => event.decision === "DENY"
  ).length

  const approvalCount = actions.filter(
    (event) => event.decision === "REQUIRE_APPROVAL"
  ).length

  const highestRisk = actions.reduce(
    (highest, event) =>
      Math.max(highest, event.risk_score),
    0
  )

  return (
    <div className="space-y-6 p-6">
      <div>
        <Link
          href={`/agents/${id}`}
          className="mb-5 inline-flex items-center gap-2 text-sm text-zinc-400 transition hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to agent
        </Link>

        <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div>
            <div className="mb-3 flex flex-wrap gap-2">
              <Badge
                variant="outline"
                className={sessionStatusClasses(status)}
              >
                {status}
              </Badge>

              <Badge
                variant="outline"
                className="border-zinc-700 bg-zinc-900 text-zinc-300"
              >
                Session
              </Badge>
            </div>

            <h2 className="text-3xl font-semibold tracking-tight">
              {session.task_id}
            </h2>

            <p className="mt-2 text-sm text-zinc-500">
              Session-level security trace showing authorization decisions,
              policy reasons, risk scores, and human approval activity.
            </p>
          </div>

          <div className="font-mono text-xs text-zinc-500">
            {session.id}
          </div>
        </div>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <MetricCard
          title="Actions"
          value={actions.length}
          description="Authorization evaluations"
          icon={<Activity className="h-4 w-4 text-zinc-500" />}
        />

        <MetricCard
          title="Allowed"
          value={allowedCount}
          description="Permitted operations"
          icon={<CheckCircle2 className="h-4 w-4 text-emerald-400" />}
        />

        <MetricCard
          title="Denied"
          value={deniedCount}
          description="Blocked operations"
          icon={<XCircle className="h-4 w-4 text-red-400" />}
        />

        <MetricCard
          title="Approvals"
          value={approvalCount}
          description="Human escalation decisions"
          icon={<KeyRound className="h-4 w-4 text-amber-400" />}
        />

        <MetricCard
          title="Max Risk"
          value={highestRisk || "—"}
          description="Highest observed action risk"
          icon={<ShieldAlert className="h-4 w-4 text-red-400" />}
        />
      </section>

      <section className="grid gap-4 xl:grid-cols-2">
        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Session Details</CardTitle>
          </CardHeader>

          <CardContent className="space-y-4">
            <DetailRow
              label="Task ID"
              value={session.task_id}
            />

            <DetailRow
              label="Session ID"
              value={session.id}
              mono
            />

            <DetailRow
              label="Agent ID"
              value={session.agent_id}
              mono
            />

            <DetailRow
              label="Started"
              value={new Date(
                session.started_at
              ).toLocaleString()}
            />

            <DetailRow
              label="Expires"
              value={
                session.expires_at
                  ? new Date(
                      session.expires_at
                    ).toLocaleString()
                  : "—"
              }
            />

            <DetailRow
              label="Ended"
              value={
                session.ended_at
                  ? new Date(
                      session.ended_at
                    ).toLocaleString()
                  : "—"
              }
            />
          </CardContent>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>Security Summary</CardTitle>
          </CardHeader>

          <CardContent className="space-y-4">
            <SummaryRow
              label="Authorization decisions"
              value={actions.length}
            />

            <SummaryRow
              label="Allowed operations"
              value={allowedCount}
            />

            <SummaryRow
              label="Denied operations"
              value={deniedCount}
            />

            <SummaryRow
              label="Approval escalations"
              value={approvalCount}
            />

            <SummaryRow
              label="Approval requests"
              value={approvals.length}
            />
          </CardContent>
        </Card>
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Security Decision Timeline
          </CardTitle>
        </CardHeader>

        <CardContent>
          {actions.length === 0 ? (
            <div className="py-12 text-center text-sm text-zinc-500">
              No authorization activity found for this session.
            </div>
          ) : (
            <div className="relative space-y-0">
              {actions.map((event, index) => (
                <div
                  key={event.id}
                  className="relative grid gap-4 pb-8 pl-8 last:pb-0"
                >
                  {index < actions.length - 1 && (
                    <div className="absolute left-[7px] top-5 h-full w-px bg-zinc-800" />
                  )}

                  <div
                    className={`absolute left-0 top-1.5 h-3.5 w-3.5 rounded-full border ${
                      event.decision === "DENY"
                        ? "border-red-400 bg-red-500"
                        : event.decision === "REQUIRE_APPROVAL"
                          ? "border-amber-400 bg-amber-500"
                          : "border-emerald-400 bg-emerald-500"
                    }`}
                  />

                  <div className="rounded-xl border border-zinc-800 bg-zinc-950/60 p-4">
                    <div className="flex flex-col justify-between gap-3 lg:flex-row lg:items-start">
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-mono text-sm font-semibold text-zinc-200">
                            {event.action}
                          </span>

                          <Badge
                            variant="outline"
                            className={decisionClasses(
                              event.decision
                            )}
                          >
                            {event.decision}
                          </Badge>

                          <Badge
                            variant="outline"
                            className={`border-zinc-700 bg-zinc-900 ${riskClasses(
                              event.risk_score
                            )}`}
                          >
                            Risk {event.risk_score}
                          </Badge>
                        </div>

                        <div className="mt-2 font-mono text-xs text-zinc-500">
                          {event.resource}
                        </div>
                      </div>

                      <div className="flex items-center gap-2 text-xs text-zinc-500">
                        <Clock3 className="h-3.5 w-3.5" />

                        {new Date(
                          event.created_at
                        ).toLocaleString()}
                      </div>
                    </div>

                    <div className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                      <MetadataField
                        label="Risk Reason"
                        value={
                          event.metadata?.risk_reason
                        }
                      />

                      <MetadataField
                        label="Policy Reason"
                        value={
                          event.metadata?.policy_reason
                        }
                      />

                      <MetadataField
                        label="Request Reason"
                        value={
                          event.metadata?.request_reason
                        }
                      />

                      <MetadataField
                        label="Policy Engine"
                        value={
                          event.metadata?.policy_engine
                        }
                      />

                      <MetadataField
                        label="Approval ID"
                        value={
                          event.metadata?.approval_id
                        }
                      />

                      <MetadataField
                        label="Grant ID"
                        value={
                          event.metadata?.grant_id
                        }
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <CardTitle>
            Human Approval Activity
          </CardTitle>
        </CardHeader>

        <CardContent>
          {approvals.length === 0 ? (
            <div className="py-10 text-center text-sm text-zinc-500">
              No approval requests were created for this session.
            </div>
          ) : (
            <div className="space-y-3">
              {approvals.map((approval) => (
                <div
                  key={approval.id}
                  className="rounded-xl border border-zinc-800 bg-zinc-950/60 p-4"
                >
                  <div className="flex flex-col justify-between gap-3 lg:flex-row">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-mono text-sm font-medium text-zinc-200">
                          {approval.action}
                        </span>

                        <Badge
                          variant="outline"
                          className={approvalStatusClasses(
                            approval.status
                          )}
                        >
                          {approval.status}
                        </Badge>

                        <Badge
                          variant="outline"
                          className={`border-zinc-700 bg-zinc-900 ${riskClasses(
                            approval.risk_score
                          )}`}
                        >
                          Risk {approval.risk_score}
                        </Badge>
                      </div>

                      <div className="mt-2 font-mono text-xs text-zinc-500">
                        {approval.resource}
                      </div>

                      <p className="mt-3 text-sm text-zinc-400">
                        {approval.reason}
                      </p>
                    </div>

                    <div className="text-xs text-zinc-500">
                      Requested{" "}
                      {new Date(
                        approval.requested_at
                      ).toLocaleString()}
                    </div>
                  </div>

                  <div className="mt-4 grid gap-3 border-t border-zinc-800 pt-4 md:grid-cols-3">
                    <MetadataField
                      label="Approval ID"
                      value={approval.id}
                    />

                    <MetadataField
                      label="Approved At"
                      value={
                        approval.approved_at
                          ? new Date(
                              approval.approved_at
                            ).toLocaleString()
                          : undefined
                      }
                    />

                    <MetadataField
                      label="Expires At"
                      value={
                        approval.expires_at
                          ? new Date(
                              approval.expires_at
                            ).toLocaleString()
                          : undefined
                      }
                    />
                  </div>
                </div>
              ))}
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

function DetailRow({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="flex items-start justify-between gap-6 border-b border-zinc-800 pb-3 last:border-0">
      <span className="text-sm text-zinc-500">
        {label}
      </span>

      <span
        className={`break-all text-right text-sm text-zinc-200 ${
          mono ? "font-mono text-xs" : ""
        }`}
      >
        {value}
      </span>
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

      <div className="mt-1 break-all text-sm text-zinc-400">
        {String(value)}
      </div>
    </div>
  )
}