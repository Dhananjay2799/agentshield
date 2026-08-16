"use client"

import {
  useEffect,
  useMemo,
  useState,
} from "react"

import Link from "next/link"

import {
  Activity,
  CheckCircle2,
  Clock3,
  Filter,
  RefreshCw,
  ShieldAlert,
  Wifi,
  WifiOff,
  XCircle,
} from "lucide-react"

import type {
  SecurityEvent,
} from "@/types/agentshield"

import { Badge } from "@/components/ui/badge"

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"

type LiveEventsClientProps = {
  initialEvents: SecurityEvent[]
}

type DecisionFilter =
  | "ALL"
  | "ALLOW"
  | "DENY"
  | "REQUIRE_APPROVAL"

type StreamStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"

type StreamSecurityEvent = {
  event_type: string
  agent_id?: string | null
  session_id?: string | null
  action: string
  resource: string
  decision: string
  risk_score: number
  metadata?: SecurityEvent["metadata"]
  occurred_at: string
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

function streamStatusClasses(
  status: StreamStatus
) {
  switch (status) {
    case "connected":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"

    case "connecting":
    case "reconnecting":
      return "border-amber-500/30 bg-amber-500/10 text-amber-300"

    case "disconnected":
      return "border-red-500/30 bg-red-500/10 text-red-300"

    default:
      return "border-zinc-700 bg-zinc-900 text-zinc-300"
  }
}

function streamStatusLabel(
  status: StreamStatus
) {
  switch (status) {
    case "connected":
      return "CONNECTED"

    case "connecting":
      return "CONNECTING"

    case "reconnecting":
      return "RECONNECTING"

    case "disconnected":
      return "DISCONNECTED"
  }
}

function convertStreamEvent(
  event: StreamSecurityEvent
): SecurityEvent {
  const syntheticID = [
    event.session_id ?? "control-plane",
    event.action,
    event.decision,
    event.occurred_at,
  ].join(":")

  return {
    id: syntheticID,
    agent_id: event.agent_id ?? null,
    session_id:
      event.session_id ?? null,
    event_type: event.event_type,
    action: event.action,
    resource: event.resource,
    decision: event.decision,
    risk_score: event.risk_score,
    metadata: event.metadata ?? {},
    created_at: event.occurred_at,
  }
}

function metadataText(
  metadata:
    | Record<string, unknown>
    | undefined,
  key: string
): string | null {
  const value =
    metadata?.[key]

  if (
    value === null ||
    value === undefined ||
    value === ""
  ) {
    return null
  }

  return String(value)
}

export default function LiveEventsClient({
  initialEvents,
}: LiveEventsClientProps) {
  const [events, setEvents] =
    useState<SecurityEvent[]>(
      initialEvents
    )

  const [filter, setFilter] =
    useState<DecisionFilter>("ALL")

  const [
    streamStatus,
    setStreamStatus,
  ] =
    useState<StreamStatus>(
      "connecting"
    )

  const [
    isRefreshing,
    setIsRefreshing,
  ] =
    useState(false)

  const [
    lastUpdated,
    setLastUpdated,
  ] =
    useState<Date | null>(null)

  async function loadEvents() {
    try {
      setIsRefreshing(true)

      const response =
        await fetch(
          "/api/events?limit=100",
          {
            cache: "no-store",
          }
        )

      if (!response.ok) {
        return
      }

      const payload:
        SecurityEvent[] =
        await response.json()

      setEvents(payload)

      setLastUpdated(
        new Date()
      )
    } finally {
      setIsRefreshing(false)
    }
  }

  useEffect(() => {
    let eventSource:
      EventSource | null = null

    let hasConnected = false

    function connect() {
      setStreamStatus(
        hasConnected
          ? "reconnecting"
          : "connecting"
      )

      eventSource =
        new EventSource(
          "/api/events/stream"
        )

      eventSource.onopen = () => {
        hasConnected = true

        setStreamStatus(
          "connected"
        )

        setLastUpdated(
          new Date()
        )
      }

      eventSource.addEventListener(
        "security_event",
        (message) => {
          try {
            const payload =
              JSON.parse(
                message.data
              ) as StreamSecurityEvent

            const event =
              convertStreamEvent(
                payload
              )

            setEvents(
              (current) => {
                const withoutDuplicate =
                  current.filter(
                    (existing) =>
                      !(
                        existing.session_id ===
                          event.session_id &&
                        existing.action ===
                          event.action &&
                        existing.decision ===
                          event.decision &&
                        existing.created_at ===
                          event.created_at
                      )
                  )

                return [
                  event,
                  ...withoutDuplicate,
                ].slice(
                  0,
                  100
                )
              }
            )

            setLastUpdated(
              new Date()
            )
          } catch {
            // Ignore malformed stream messages.
          }
        }
      )

      eventSource.onerror =
        () => {
          setStreamStatus(
            hasConnected
              ? "reconnecting"
              : "disconnected"
          )
        }
    }

    connect()

    return () => {
      eventSource?.close()
    }
  }, [])

  const filteredEvents =
    useMemo(() => {
      if (
        filter === "ALL"
      ) {
        return events
      }

      return events.filter(
        (event) =>
          event.decision ===
          filter
      )
    }, [
      events,
      filter,
    ])

  const allowedCount =
    events.filter(
      (event) =>
        event.decision ===
        "ALLOW"
    ).length

  const deniedCount =
    events.filter(
      (event) =>
        event.decision ===
        "DENY"
    ).length

  const approvalCount =
    events.filter(
      (event) =>
        event.decision ===
        "REQUIRE_APPROVAL"
    ).length

  const criticalCount =
    events.filter(
      (event) =>
        event.risk_score >= 80
    ).length

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          title="Allowed"
          value={
            allowedCount
          }
          description="Permitted actions"
          icon={
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          }
        />

        <MetricCard
          title="Denied"
          value={
            deniedCount
          }
          description="Blocked actions"
          icon={
            <XCircle className="h-4 w-4 text-red-400" />
          }
        />

        <MetricCard
          title="Approval Required"
          value={
            approvalCount
          }
          description="Human escalation events"
          icon={
            <ShieldAlert className="h-4 w-4 text-amber-400" />
          }
        />

        <MetricCard
          title="Critical Risk"
          value={
            criticalCount
          }
          description="Risk score 80 or greater"
          icon={
            <Activity className="h-4 w-4 text-red-400" />
          }
        />
      </section>

      <Card className="border-zinc-800 bg-zinc-900/70">
        <CardHeader>
          <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-center">
            <div>
              <CardTitle>
                Security Event Stream
              </CardTitle>

              <p className="mt-1 text-sm text-zinc-500">
                Server-pushed
                AgentShield
                security activity
                over SSE.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <Badge
                variant="outline"
                className={streamStatusClasses(
                  streamStatus
                )}
              >
                {streamStatus ===
                "connected" ? (
                  <Wifi className="mr-2 h-3.5 w-3.5" />
                ) : (
                  <WifiOff className="mr-2 h-3.5 w-3.5" />
                )}

                {streamStatusLabel(
                  streamStatus
                )}
              </Badge>

              <Filter className="ml-2 mr-1 h-4 w-4 text-zinc-500" />

              {(
                [
                  "ALL",
                  "ALLOW",
                  "DENY",
                  "REQUIRE_APPROVAL",
                ] as DecisionFilter[]
              ).map(
                (value) => (
                  <Button
                    key={
                      value
                    }
                    size="sm"
                    variant={
                      filter ===
                      value
                        ? "default"
                        : "outline"
                    }
                    onClick={() =>
                      setFilter(
                        value
                      )
                    }
                  >
                    {value ===
                    "REQUIRE_APPROVAL"
                      ? "APPROVAL"
                      : value}
                  </Button>
                )
              )}

              <Button
                size="sm"
                variant="outline"
                disabled={
                  isRefreshing
                }
                onClick={() =>
                  void loadEvents()
                }
              >
                <RefreshCw
                  className={`mr-2 h-4 w-4 ${
                    isRefreshing
                      ? "animate-spin"
                      : ""
                  }`}
                />

                Refresh History
              </Button>
            </div>
          </div>

          <div className="mt-2 flex items-center gap-2 text-xs text-zinc-600">
            <Clock3 className="h-3.5 w-3.5" />

            Last event/update{" "}
            {lastUpdated
              ? lastUpdated.toLocaleTimeString()
              : "waiting for stream..."}
          </div>
        </CardHeader>

        <CardContent>
          {filteredEvents.length ===
          0 ? (
            <div className="py-16 text-center text-sm text-zinc-500">
              No matching
              security events.
            </div>
          ) : (
            <div className="space-y-3">
              {filteredEvents.map(
                (event) => {
                  const policyEngine =
                    metadataText(
                      event.metadata,
                      "policy_engine"
                    )

                  const environment =
                    metadataText(
                      event.metadata,
                      "environment"
                    )

                  const riskReason =
                    metadataText(
                      event.metadata,
                      "risk_reason"
                    )

                  const policyReason =
                    metadataText(
                      event.metadata,
                      "policy_reason"
                    )

                  const requestReason =
                    metadataText(
                      event.metadata,
                      "request_reason"
                    )

                  const approvalID =
                    metadataText(
                      event.metadata,
                      "approval_id"
                    )

                  const grantID =
                    metadataText(
                      event.metadata,
                      "grant_id"
                    )

                  const policyName =
                    metadataText(
                      event.metadata,
                      "policy_name"
                    )

                  return (
                    <div
                      key={
                        event.id
                      }
                      className="rounded-xl border border-zinc-800 bg-zinc-950/60 p-4"
                    >
                      <div className="flex flex-col justify-between gap-4 xl:flex-row xl:items-start">
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-mono text-sm font-semibold text-zinc-200">
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
                              <span
                                className={riskClasses(
                                  event.risk_score
                                )}
                              >
                                Risk{" "}
                                {
                                  event.risk_score
                                }
                              </span>
                            </Badge>

                            <Badge
                              variant="outline"
                              className="border-zinc-700 bg-zinc-900 text-zinc-400"
                            >
                              {
                                event.event_type
                              }
                            </Badge>
                          </div>

                          <div className="mt-2 font-mono text-xs text-zinc-500">
                            {
                              event.resource
                            }
                          </div>

                          {policyName && (
                            <div className="mt-2 text-xs text-zinc-500">
                              Policy:{" "}
                              <span className="text-zinc-300">
                                {
                                  policyName
                                }
                              </span>
                            </div>
                          )}
                        </div>

                        <div className="whitespace-nowrap text-xs text-zinc-500">
                          {new Date(
                            event.created_at
                          ).toLocaleString()}
                        </div>
                      </div>

                      <div className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                        <EventField
                          label="Agent"
                          value={
                            event.agent_id ? (
                              <Link
                                href={`/agents/${event.agent_id}`}
                                className="font-mono text-xs text-zinc-300 transition hover:text-white"
                              >
                                {event.agent_id.slice(
                                  0,
                                  12
                                )}
                                ...
                              </Link>
                            ) : (
                              "Control plane"
                            )
                          }
                        />

                        <EventField
                          label="Session"
                          value={
                            event.agent_id &&
                            event.session_id ? (
                              <Link
                                href={`/agents/${event.agent_id}/sessions/${event.session_id}`}
                                className="font-mono text-xs text-zinc-300 transition hover:text-white"
                              >
                                {event.session_id.slice(
                                  0,
                                  12
                                )}
                                ...
                              </Link>
                            ) : (
                              "—"
                            )
                          }
                        />

                        <EventField
                          label="Policy Engine"
                          value={
                            policyEngine ??
                            "—"
                          }
                        />

                        <EventField
                          label="Environment"
                          value={
                            environment ??
                            "—"
                          }
                        />
                      </div>

                      <div className="mt-4 grid gap-4 border-t border-zinc-800 pt-4 md:grid-cols-2 xl:grid-cols-3">
                        <EventField
                          label="Risk Reason"
                          value={
                            riskReason ??
                            "—"
                          }
                        />

                        <EventField
                          label="Policy Reason"
                          value={
                            policyReason ??
                            "—"
                          }
                        />

                        <EventField
                          label="Request Reason"
                          value={
                            requestReason ??
                            "—"
                          }
                        />
                      </div>

                      {(approvalID ||
                        grantID) && (
                        <div className="mt-4 grid gap-4 border-t border-zinc-800 pt-4 md:grid-cols-2">
                          <EventField
                            label="Approval ID"
                            value={
                              approvalID ??
                              "—"
                            }
                          />

                          <EventField
                            label="Grant ID"
                            value={
                              grantID ??
                              "—"
                            }
                          />
                        </div>
                      )}
                    </div>
                  )
                }
              )}
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

function EventField({
  label,
  value,
}: {
  label: string
  value: React.ReactNode
}) {
  return (
    <div>
      <div className="text-xs uppercase tracking-wide text-zinc-600">
        {label}
      </div>

      <div className="mt-1 break-all text-sm text-zinc-400">
        {value}
      </div>
    </div>
  )
}