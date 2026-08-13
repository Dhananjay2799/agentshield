import Link from "next/link"

import {
  AlertTriangle,
  Clock3,
  KeyRound,
  ShieldCheck,
  UserCheck,
} from "lucide-react"

import ApprovalActions from "@/components/approvals/ApprovalActions"
import { getPendingApprovals } from "@/lib/api/approvals"

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

function riskClasses(score: number) {
  if (score >= 80) {
    return "border-red-500/30 bg-red-500/10 text-red-300"
  }

  if (score >= 50) {
    return "border-amber-500/30 bg-amber-500/10 text-amber-300"
  }

  return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
}

function getMinutesRemaining(expiresAt?: string) {
  if (!expiresAt) {
    return null
  }

  return Math.max(
    0,
    Math.floor(
      (new Date(expiresAt).getTime() - Date.now()) / 60000
    )
  )
}

export default async function ApprovalsPage() {
  const approvals = await getPendingApprovals()

  const highRiskApprovals = approvals.filter(
    (approval) => approval.risk_score >= 80
  )

  const elevatedRiskApprovals = approvals.filter(
    (approval) => approval.risk_score >= 50
  )

  const expiringSoon = approvals.filter((approval) => {
    const remaining = getMinutesRemaining(
      approval.expires_at
    )

    return (
      remaining !== null &&
      remaining <= 5
    )
  })

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">
            Approval Queue
          </h2>

          <p className="mt-1 text-sm text-zinc-500">
            Review autonomous-agent actions that require explicit human authorization.
          </p>
        </div>

        <Badge
          variant="outline"
          className="w-fit border-amber-500/30 bg-amber-500/10 text-amber-300"
        >
          {approvals.length} Pending
        </Badge>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          title="Pending Requests"
          value={approvals.length}
          description="Awaiting analyst decision"
          icon={
            <UserCheck className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Elevated Risk"
          value={elevatedRiskApprovals.length}
          description="Risk score 50 or greater"
          icon={
            <AlertTriangle className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Critical Risk"
          value={highRiskApprovals.length}
          description="Risk score 80 or greater"
          icon={
            <ShieldCheck className="h-4 w-4 text-red-400" />
          }
        />

        <MetricCard
          title="Expiring Soon"
          value={expiringSoon.length}
          description="Five minutes or less remaining"
          icon={
            <Clock3 className="h-4 w-4 text-zinc-500" />
          }
        />
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>
              Pending Human Approvals
            </CardTitle>

            <KeyRound className="h-4 w-4 text-zinc-500" />
          </div>
        </CardHeader>

        <CardContent>
          {approvals.length === 0 ? (
            <div className="flex min-h-56 flex-col items-center justify-center text-center">
              <ShieldCheck className="mb-4 h-9 w-9 text-emerald-500/60" />

              <div className="font-medium text-zinc-200">
                Approval queue is clear
              </div>

              <p className="mt-2 max-w-md text-sm text-zinc-500">
                There are currently no valid pending autonomous-agent approval requests.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="border-zinc-800 hover:bg-transparent">
                    <TableHead>
                      Agent
                    </TableHead>

                    <TableHead>
                      Requested Action
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
                      Requested
                    </TableHead>

                    <TableHead>
                      Expires
                    </TableHead>

                    <TableHead className="text-right">
                      Decision
                    </TableHead>
                  </TableRow>
                </TableHeader>

                <TableBody>
                  {approvals.map((approval) => {
                    const minutesRemaining =
                      getMinutesRemaining(
                        approval.expires_at
                      )

                    return (
                      <TableRow
                        key={approval.id}
                        className="border-zinc-800"
                      >
                        <TableCell>
                          <Link
                            href={`/agents/${approval.agent_id}`}
                            className="font-mono text-xs text-zinc-300 transition hover:text-white"
                          >
                            {approval.agent_id.slice(
                              0,
                              12
                            )}
                            ...
                          </Link>

                          <div className="mt-1">
                            <Link
                              href={`/agents/${approval.agent_id}/sessions/${approval.session_id}`}
                              className="text-xs text-zinc-500 transition hover:text-zinc-300"
                            >
                              View session →
                            </Link>
                          </div>
                        </TableCell>

                        <TableCell>
                          <div className="font-mono text-sm font-medium text-zinc-200">
                            {approval.action}
                          </div>

                          <div className="mt-1 font-mono text-xs text-zinc-600">
                            {approval.id.slice(
                              0,
                              12
                            )}
                            ...
                          </div>
                        </TableCell>

                        <TableCell className="font-mono text-xs text-zinc-400">
                          {approval.resource}
                        </TableCell>

                        <TableCell className="max-w-sm text-sm text-zinc-400">
                          {approval.reason}
                        </TableCell>

                        <TableCell>
                          <Badge
                            variant="outline"
                            className={riskClasses(
                              approval.risk_score
                            )}
                          >
                            Risk{" "}
                            {approval.risk_score}
                          </Badge>
                        </TableCell>

                        <TableCell className="whitespace-nowrap text-sm text-zinc-500">
                          {new Date(
                            approval.requested_at
                          ).toLocaleString()}
                        </TableCell>

                        <TableCell className="whitespace-nowrap">
                          {approval.expires_at ? (
                            <div>
                              <div className="text-sm text-zinc-500">
                                {new Date(
                                  approval.expires_at
                                ).toLocaleString()}
                              </div>

                              {minutesRemaining !==
                                null && (
                                <div
                                  className={
                                    minutesRemaining <=
                                    5
                                      ? "mt-1 text-xs text-red-400"
                                      : "mt-1 text-xs text-zinc-600"
                                  }
                                >
                                  {
                                    minutesRemaining
                                  }{" "}
                                  min remaining
                                </div>
                              )}
                            </div>
                          ) : (
                            <span className="text-zinc-600">
                              —
                            </span>
                          )}
                        </TableCell>

                        <TableCell>
                          <ApprovalActions
                            approvalId={
                              approval.id
                            }
                          />
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
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