import Link from "next/link"

import {
  Activity,
  ArrowDown,
  ArrowLeft,
  CheckCircle2,
  Clock3,
  KeyRound,
  ShieldAlert,
  ShieldCheck,
  XCircle,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"

import {
  getApprovalLineage,
} from "@/lib/api/approvals"

import type {
  ApprovalLineage,
} from "@/types/agentshield"

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

  policy_matched?: boolean
  policy_id?: string
  policy_name?: string
  policy_priority?: number
  policy_effect?: string
  policy_version?: number
  policy_source?: string

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

function sessionStatus(
  session: AgentSession
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

function sessionStatusClasses(
  status: string
) {
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

function decisionClasses(
  decision: string
) {
  switch (decision) {
    case "ALLOW":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "DENY":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    case "REQUIRE_APPROVAL":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "SUCCESS":
      return "border-sky-500/30 bg-sky-500/10 text-sky-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

function approvalStatusClasses(
  status: string
) {
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

function riskClasses(
  score: number
) {
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

function riskBadgeClasses(
  score: number
) {
  if (score >= 80) {
    return "border-red-500/30 bg-red-500/10 text-red-300"
  }

  if (score >= 50) {
    return "border-amber-500/30 bg-amber-500/10 text-amber-300"
  }

  if (score > 0) {
    return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
  }

  return "border-zinc-700 bg-zinc-900 text-zinc-400"
}

function riskLabel(
  score: number
) {
  if (score >= 80) {
    return "CRITICAL"
  }

  if (score >= 50) {
    return "ELEVATED"
  }

  if (score > 0) {
    return "LOW"
  }

  return "NONE"
}

function textMetadata(
  metadata: AuditMetadata,
  key: keyof AuditMetadata
): string | null {
  const value = metadata[key]

  if (
    value === undefined ||
    value === null ||
    value === ""
  ) {
    return null
  }

  return String(value)
}

function numberMetadata(
  metadata: AuditMetadata,
  key: keyof AuditMetadata
): number | null {
  const value = metadata[key]

  if (typeof value === "number") {
    return value
  }

  if (
    typeof value === "string" &&
    value.trim() !== ""
  ) {
    const parsed = Number(value)

    if (Number.isFinite(parsed)) {
      return parsed
    }
  }

  return null
}

function booleanMetadata(
  metadata: AuditMetadata,
  key: keyof AuditMetadata
): boolean {
  const value = metadata[key]

  if (typeof value === "boolean") {
    return value
  }

  if (typeof value === "string") {
    return value.toLowerCase() === "true"
  }

  return false
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

  const [
    session,
    actions,
    approvals,
  ] = await Promise.all([
    getSession(sessionId),
    getSessionActions(sessionId),
    getSessionApprovals(sessionId),
  ])

  const lineageResults =
  await Promise.all(
    approvals.map(
      async (approval) => {
        const lineage =
          await getApprovalLineage(
            approval.id
          )

        return lineage
      }
    )
  )

const lineages =
  lineageResults.filter(
    (
      lineage
    ): lineage is ApprovalLineage =>
      lineage !== null
  )

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
    (event) =>
      event.decision === "ALLOW"
  ).length

  const deniedCount = actions.filter(
    (event) =>
      event.decision === "DENY"
  ).length

  const approvalCount = actions.filter(
    (event) =>
      event.decision === "REQUIRE_APPROVAL"
  ).length

  const highestRisk = actions.reduce(
    (highest, event) =>
      Math.max(
        highest,
        event.risk_score
      ),
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
                className={sessionStatusClasses(
                  status
                )}
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

            <p className="mt-2 max-w-3xl text-sm text-zinc-500">
              Session-level security investigation showing
              risk assessment, matched policy evidence,
              authorization decisions, and human approval
              activity.
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
          icon={
            <Activity className="h-4 w-4 text-zinc-500" />
          }
        />

        <MetricCard
          title="Allowed"
          value={allowedCount}
          description="Permitted operations"
          icon={
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          }
        />

        <MetricCard
          title="Denied"
          value={deniedCount}
          description="Blocked operations"
          icon={
            <XCircle className="h-4 w-4 text-red-400" />
          }
        />

        <MetricCard
          title="Approvals"
          value={approvalCount}
          description="Human escalation decisions"
          icon={
            <KeyRound className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Max Risk"
          value={highestRisk || "—"}
          description="Highest observed action risk"
          icon={
            <ShieldAlert className="h-4 w-4 text-red-400" />
          }
        />
      </section>

      <section className="grid gap-4 xl:grid-cols-2">
        <Card className="border-zinc-800 bg-zinc-900/70">
          <CardHeader>
            <CardTitle>
              Session Details
            </CardTitle>
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
            <CardTitle>
              Security Summary
            </CardTitle>
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
            Explainable Decision Trace
          </CardTitle>

          <p className="text-sm text-zinc-500">
            Evidence showing how AgentShield evaluated each
            autonomous-agent request and reached the final
            authorization decision.
          </p>
        </CardHeader>

        <CardContent>
          {actions.length === 0 ? (
            <div className="py-12 text-center text-sm text-zinc-500">
              No authorization activity found for this session.
            </div>
          ) : (
            <div className="space-y-5">
              {actions.map(
                (event) => (
                  <DecisionTraceCard
                    key={event.id}
                    event={event}
                  />
                )
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-zinc-800 bg-zinc-900/70">
  	<CardHeader>
    	  <CardTitle>
      		Authorization Lineage
    	  </CardTitle>

      <p className="text-sm text-zinc-500">
      	Trace human approvals into temporary
      	authorization grants and their final
      	consumption state.
      </p>
      </CardHeader>

      <CardContent>
    	{lineages.length === 0 ? (
      	<div className="py-10 text-center text-sm text-zinc-500">
        	No authorization lineage exists for
        	this session.
      	</div>
    	) : (
      	<div className="space-y-5">
        	{lineages.map(
        	  (lineage) => (
        	    <AuthorizationLineageCard
        	      key={
        	        lineage.approval.id
       	       }
       	       lineage={
       	         lineage
       	       }
             />
           )
         )}
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
              {approvals.map(
                (approval) => (
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
                )
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function DecisionTraceCard({
  event,
}: {
  event: AuditEvent
}) {
  const metadata =
    event.metadata ?? {}

  const policyMatched =
    booleanMetadata(
      metadata,
      "policy_matched"
    )

  const policyID =
    textMetadata(
      metadata,
      "policy_id"
    )

  const policyName =
    textMetadata(
      metadata,
      "policy_name"
    )

  const policyPriority =
    numberMetadata(
      metadata,
      "policy_priority"
    )

  const policyEffect =
    textMetadata(
      metadata,
      "policy_effect"
    )

  const policyVersion =
    numberMetadata(
      metadata,
      "policy_version"
    )

  const policySource =
    textMetadata(
      metadata,
      "policy_source"
    )

  const policyEngine =
    textMetadata(
      metadata,
      "policy_engine"
    )

  const riskReason =
    textMetadata(
      metadata,
      "risk_reason"
    )

  const policyReason =
    textMetadata(
      metadata,
      "policy_reason"
    )

  const requestReason =
    textMetadata(
      metadata,
      "request_reason"
    )

  const approvalID =
    textMetadata(
      metadata,
      "approval_id"
    )

  const grantID =
    textMetadata(
      metadata,
      "grant_id"
    )

  const agentType =
    textMetadata(
      metadata,
      "agent_type"
    )

  const environment =
    textMetadata(
      metadata,
      "environment"
    )

  const approvalRequired =
    event.decision ===
      "REQUIRE_APPROVAL" ||
    Boolean(approvalID) ||
    Boolean(grantID)

  return (
    <div className="rounded-2xl border border-zinc-800 bg-zinc-950/60 p-5">
      <div className="flex flex-col justify-between gap-4 xl:flex-row xl:items-start">
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
              className={riskBadgeClasses(
                event.risk_score
              )}
            >
              Risk {event.risk_score} ·{" "}
              {riskLabel(
                event.risk_score
              )}
            </Badge>
          </div>

          <div className="mt-2 font-mono text-xs text-zinc-500">
            {event.resource}
          </div>
        </div>

        <div className="flex items-center gap-2 whitespace-nowrap text-xs text-zinc-500">
          <Clock3 className="h-3.5 w-3.5" />

          {new Date(
            event.created_at
          ).toLocaleString()}
        </div>
      </div>

      <div className="mt-5 grid gap-4 xl:grid-cols-[1fr_auto_1fr_auto_1fr_auto_1fr] xl:items-stretch">
        <TraceStage
          title="Request"
          badge="INPUT"
          badgeClass="border-zinc-700 bg-zinc-900 text-zinc-300"
        >
          <TraceRow
            label="Action"
            value={event.action}
            mono
          />

          <TraceRow
            label="Resource"
            value={event.resource}
            mono
          />

          <TraceRow
            label="Agent Type"
            value={
              agentType ?? "—"
            }
          />

          <TraceRow
            label="Environment"
            value={
              environment ?? "—"
            }
          />

          {requestReason && (
            <TraceRow
              label="Request Reason"
              value={requestReason}
            />
          )}
        </TraceStage>

        <TraceArrow />

        <TraceStage
          title="Risk"
          badge={String(
            event.risk_score
          )}
          badgeClass={riskBadgeClasses(
            event.risk_score
          )}
        >
          <TraceRow
            label="Score"
            value={`${event.risk_score} / 100`}
          />

          <TraceRow
            label="Level"
            value={riskLabel(
              event.risk_score
            )}
          />

          <TraceRow
            label="Reason"
            value={
              riskReason ??
              "No risk explanation recorded"
            }
          />
        </TraceStage>

        <TraceArrow />

        <TraceStage
          title="Policy"
          badge={
            policyMatched
              ? "MATCHED"
              : "NO MATCH"
          }
          badgeClass={
            policyMatched
              ? "border-sky-500/30 bg-sky-500/10 text-sky-300"
              : "border-zinc-700 bg-zinc-900 text-zinc-400"
          }
        >
          <TraceRow
            label="Engine"
            value={
              policyEngine ??
              "OPA"
            }
          />

          {policyMatched ? (
            <>
              <TraceRow
                label="Policy"
                value={
                  policyID &&
                  policyName ? (
                    <Link
                      href={`/policies/${policyID}`}
                      className="text-zinc-200 underline-offset-4 hover:text-white hover:underline"
                    >
                      {policyName}
                    </Link>
                  ) : (
                    policyName ??
                    "Managed policy"
                  )
                }
              />

              <TraceRow
                label="Priority"
                value={
                  policyPriority !==
                  null
                    ? String(
                        policyPriority
                      )
                    : "—"
                }
              />

              <TraceRow
                label="Effect"
                value={
                  policyEffect ??
                  "—"
                }
              />

              <TraceRow
                label="Version"
                value={
                  policyVersion !==
                  null
                    ? `v${policyVersion}`
                    : "—"
                }
              />

              <TraceRow
                label="Source"
                value={
                  policySource ??
                  "—"
                }
              />
            </>
          ) : (
            <TraceRow
              label="Reason"
              value={
                policyReason ??
                "No managed policy matched"
              }
            />
          )}
        </TraceStage>

        <TraceArrow />

        <TraceStage
          title="Authorization"
          badge={
            grantID
              ? "GRANT USED"
              : approvalRequired
                ? "HUMAN"
                : "DIRECT"
          }
          badgeClass={
            grantID
              ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
              : approvalRequired
                ? "border-amber-500/30 bg-amber-500/10 text-amber-300"
                : "border-zinc-700 bg-zinc-900 text-zinc-300"
          }
        >
          <TraceRow
            label="Approval Required"
            value={
              approvalRequired
                ? "Yes"
                : "No"
            }
          />

          <TraceRow
            label="Grant Used"
            value={
              grantID
                ? "Yes"
                : "No"
            }
          />

          {approvalID && (
            <TraceRow
              label="Approval ID"
              value={approvalID}
              mono
            />
          )}

          {grantID && (
            <TraceRow
              label="Grant ID"
              value={grantID}
              mono
            />
          )}
        </TraceStage>
      </div>

      <div className="mt-5 rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
        <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
          <div>
            <div className="text-xs uppercase tracking-wide text-zinc-600">
              Final Decision
            </div>

            <div className="mt-2 flex flex-wrap items-center gap-2">
              <Badge
                variant="outline"
                className={decisionClasses(
                  event.decision
                )}
              >
                {event.decision}
              </Badge>

              {policyMatched &&
                policyName && (
                  <span className="text-sm text-zinc-400">
                    Policy{" "}
                    <span className="text-zinc-200">
                      {policyName}
                    </span>
                  </span>
                )}
            </div>
          </div>

          <ShieldCheck
            className={`h-7 w-7 ${
              event.decision ===
              "ALLOW"
                ? "text-emerald-400"
                : event.decision ===
                    "DENY"
                  ? "text-red-400"
                  : "text-amber-400"
            }`}
          />
        </div>

        {policyReason && (
          <p className="mt-3 text-sm text-zinc-500">
            {policyReason}
          </p>
        )}
      </div>
    </div>
  )
}

function TraceStage({
  title,
  badge,
  badgeClass,
  children,
}: {
  title: string
  badge: string
  badgeClass: string
  children: React.ReactNode
}) {
  return (
    <div className="min-w-0 rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold text-zinc-200">
          {title}
        </div>

        <Badge
          variant="outline"
          className={badgeClass}
        >
          {badge}
        </Badge>
      </div>

      <div className="space-y-3">
        {children}
      </div>
    </div>
  )
}

function TraceArrow() {
  return (
    <>
      <div className="hidden items-center justify-center xl:flex">
        <ArrowDown className="h-5 w-5 -rotate-90 text-zinc-700" />
      </div>

      <div className="flex items-center justify-center xl:hidden">
        <ArrowDown className="h-5 w-5 text-zinc-700" />
      </div>
    </>
  )
}

function TraceRow({
  label,
  value,
  mono = false,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div>
      <div className="text-[11px] uppercase tracking-wide text-zinc-600">
        {label}
      </div>

      <div
        className={`mt-1 break-all text-sm text-zinc-400 ${
          mono
            ? "font-mono text-xs"
            : ""
        }`}
      >
        {value}
      </div>
    </div>
  )
}

function AuthorizationLineageCard({
  lineage,
}: {
  lineage: ApprovalLineage
}) {
  const approval =
    lineage.approval

  const grant =
    lineage.grant

  const state =
    lineage.lineage.state

  const finalDecision =
    lineage.lineage
      .final_decision

  return (
    <div className="rounded-2xl border border-zinc-800 bg-zinc-950/60 p-5">
      <div className="mb-5 flex flex-col justify-between gap-3 md:flex-row md:items-center">
        <div>
          <div className="text-sm font-semibold text-zinc-200">
            Authorization Chain
          </div>

          <div className="mt-1 font-mono text-xs text-zinc-500">
            {approval.action}
            {" → "}
            {approval.resource}
          </div>
        </div>

        <Badge
          variant="outline"
          className={lineageStateClasses(
            state
          )}
        >
          {formatLineageState(
            state
          )}
        </Badge>
      </div>

      <div className="grid gap-4 xl:grid-cols-[1fr_auto_1fr_auto_1fr] xl:items-stretch">
        <LineageStage
          title="Approval Request"
          status={
            approval.status
          }
          statusClass={approvalStatusClasses(
            approval.status
          )}
        >
          <LineageRow
            label="Approval ID"
            value={
              approval.id
            }
            mono
          />

          <LineageRow
            label="Risk"
            value={String(
              approval.risk_score
            )}
          />

          <LineageRow
            label="Requested"
            value={new Date(
              approval.requested_at
            ).toLocaleString()}
          />

          {approval.approved_at && (
            <LineageRow
              label="Approved"
              value={new Date(
                approval.approved_at
              ).toLocaleString()}
            />
          )}
        </LineageStage>

        <TraceArrow />

        <LineageStage
          title="Temporary Grant"
          status={
            grant?.status ??
            "not issued"
          }
          statusClass={
            grant
              ? grantStatusClasses(
                  grant.status
                )
              : "border-zinc-700 bg-zinc-900 text-zinc-400"
          }
        >
          {grant ? (
            <>
              <LineageRow
                label="Grant ID"
                value={
                  grant.id
                }
                mono
              />

              <LineageRow
                label="Issued"
                value={new Date(
                  grant.issued_at
                ).toLocaleString()}
              />

              <LineageRow
                label="Expires"
                value={new Date(
                  grant.expires_at
                ).toLocaleString()}
              />

              <LineageRow
                label="Used"
                value={
                  grant.used_at
                    ? new Date(
                        grant.used_at
                      ).toLocaleString()
                    : "—"
                }
              />
            </>
          ) : (
            <LineageRow
              label="State"
              value="No grant issued"
            />
          )}
        </LineageStage>

        <TraceArrow />

        <LineageStage
          title="Final Authorization"
          status={
            finalDecision ??
            "PENDING"
          }
          statusClass={
            finalDecision
              ? decisionClasses(
                  finalDecision
                )
              : "border-zinc-700 bg-zinc-900 text-zinc-400"
          }
        >
          <LineageRow
            label="Lineage State"
            value={formatLineageState(
              state
            )}
          />

          <LineageRow
            label="Decision"
            value={
              finalDecision ??
              "Awaiting completion"
            }
          />

          <LineageRow
            label="Grant Consumed"
            value={
              grant?.status ===
              "used"
                ? "Yes"
                : "No"
            }
          />
        </LineageStage>
      </div>
    </div>
  )
}

function LineageStage({
  title,
  status,
  statusClass,
  children,
}: {
  title: string
  status: string
  statusClass: string
  children: React.ReactNode
}) {
  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold text-zinc-200">
          {title}
        </div>

        <Badge
          variant="outline"
          className={
            statusClass
          }
        >
          {status}
        </Badge>
      </div>

      <div className="space-y-3">
        {children}
      </div>
    </div>
  )
}

function LineageRow({
  label,
  value,
  mono = false,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div>
      <div className="text-[11px] uppercase tracking-wide text-zinc-600">
        {label}
      </div>

      <div
        className={`mt-1 break-all text-sm text-zinc-400 ${
          mono
            ? "font-mono text-xs"
            : ""
        }`}
      >
        {value}
      </div>
    </div>
  )
}

function grantStatusClasses(
  status: string
) {
  switch (status) {
    case "active":
      return "border-sky-500/30 bg-sky-500/10 text-sky-300"

    case "used":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "expired":
      return "border-zinc-600 bg-zinc-800 text-zinc-300"

    case "revoked":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-400"
  }
}

function lineageStateClasses(
  state: string
) {
  switch (state) {
    case "consumed":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "grant_active":
      return "border-sky-500/30 bg-sky-500/10 text-sky-300"

    case "awaiting_human_approval":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "denied":
    case "grant_revoked":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-400"
  }
}

function formatLineageState(
  state: string
) {
  return state
    .replaceAll("_", " ")
    .toUpperCase()
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
          mono
            ? "font-mono text-xs"
            : ""
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