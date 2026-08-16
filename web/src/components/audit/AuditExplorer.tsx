"use client"

import {
  useEffect,
  useMemo,
  useState,
} from "react"

import Link from "next/link"

import {
  ChevronDown,
  ChevronUp,
  Filter,
  RefreshCw,
  Search,
  Wifi,
  WifiOff,
  X,
} from "lucide-react"

import type {
  SecurityEvent,
} from "@/types/agentshield"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"

type AuditExplorerProps = {
  events: SecurityEvent[]
}

type SourceFilter =
  | "ALL"
  | "AGENT"
  | "POLICY"

type DecisionFilter =
  | "ALL"
  | "ALLOW"
  | "DENY"
  | "REQUIRE_APPROVAL"
  | "SUCCESS"

type RiskFilter =
  | "ALL"
  | "ELEVATED"
  | "CRITICAL"

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
      return "border-zinc-700 bg-zinc-900 text-zinc-400"
  }
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
  }
}

function streamStatusLabel(
  status: StreamStatus
) {
  switch (status) {
    case "connected":
      return "LIVE"

    case "connecting":
      return "CONNECTING"

    case "reconnecting":
      return "RECONNECTING"

    case "disconnected":
      return "DISCONNECTED"
  }
}

function eventSource(
  event: SecurityEvent
): "AGENT" | "POLICY" {
  return event.event_type.startsWith(
    "policy."
  )
    ? "POLICY"
    : "AGENT"
}

function metadataText(
  event: SecurityEvent,
  key: string
): string | null {
  const value =
    event.metadata?.[key]

  if (
    value === undefined ||
    value === null ||
    value === ""
  ) {
    return null
  }

  return String(value)
}

function convertStreamEvent(
  event: StreamSecurityEvent
): SecurityEvent {
  const syntheticID = [
    event.session_id ??
      "control-plane",
    event.action,
    event.decision,
    event.occurred_at,
  ].join(":")

  return {
    id: syntheticID,

    agent_id:
      event.agent_id ?? null,

    session_id:
      event.session_id ?? null,

    event_type:
      event.event_type,

    action:
      event.action,

    resource:
      event.resource,

    decision:
      event.decision,

    risk_score:
      event.risk_score,

    metadata:
      event.metadata ?? {},

    created_at:
      event.occurred_at,
  }
}

export function AuditExplorer({
  events: initialEvents,
}: AuditExplorerProps) {
  const [
    events,
    setEvents,
  ] =
    useState<SecurityEvent[]>(
      initialEvents
    )

  const [
    search,
    setSearch,
  ] =
    useState("")

  const [
    sourceFilter,
    setSourceFilter,
  ] =
    useState<SourceFilter>(
      "ALL"
    )

  const [
    decisionFilter,
    setDecisionFilter,
  ] =
    useState<DecisionFilter>(
      "ALL"
    )

  const [
    riskFilter,
    setRiskFilter,
  ] =
    useState<RiskFilter>(
      "ALL"
    )

  const [
    expandedEvents,
    setExpandedEvents,
  ] =
    useState<Set<string>>(
      new Set()
    )

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
    useState<Date | null>(
      null
    )

  async function loadEvents() {
    try {
      setIsRefreshing(true)

      const response =
        await fetch(
          "/api/events?limit=100",
          {
            cache:
              "no-store",
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
      EventSource | null =
      null

    let hasConnected =
      false

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

      eventSource.onopen =
        () => {
          hasConnected =
            true

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

            const newEvent =
              convertStreamEvent(
                payload
              )

            setEvents(
              (current) => {
                const deduplicated =
                  current.filter(
                    (
                      existing
                    ) =>
                      !(
                        existing.session_id ===
                          newEvent.session_id &&
                        existing.action ===
                          newEvent.action &&
                        existing.decision ===
                          newEvent.decision &&
                        existing.created_at ===
                          newEvent.created_at
                      )
                  )

                return [
                  newEvent,
                  ...deduplicated,
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
            // Ignore malformed SSE messages.
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
      const normalizedSearch =
        search
          .trim()
          .toLowerCase()

      return events.filter(
        (event) => {
          const source =
            eventSource(
              event
            )

          if (
            sourceFilter !==
              "ALL" &&
            source !==
              sourceFilter
          ) {
            return false
          }

          if (
            decisionFilter !==
              "ALL" &&
            event.decision !==
              decisionFilter
          ) {
            return false
          }

          if (
            riskFilter ===
              "ELEVATED" &&
            event.risk_score <
              50
          ) {
            return false
          }

          if (
            riskFilter ===
              "CRITICAL" &&
            event.risk_score <
              80
          ) {
            return false
          }

          if (
            !normalizedSearch
          ) {
            return true
          }

          const policyName =
            metadataText(
              event,
              "policy_name"
            )

          const searchable =
            [
              event.action,
              event.resource,
              event.event_type,
              event.decision,
              event.agent_id,
              event.session_id,
              policyName,
            ]
              .filter(Boolean)
              .join(" ")
              .toLowerCase()

          return searchable.includes(
            normalizedSearch
          )
        }
      )
    }, [
      events,
      search,
      sourceFilter,
      decisionFilter,
      riskFilter,
    ])

  const hasFilters =
    search.trim() !== "" ||
    sourceFilter !==
      "ALL" ||
    decisionFilter !==
      "ALL" ||
    riskFilter !==
      "ALL"

  function clearFilters() {
    setSearch("")
    setSourceFilter(
      "ALL"
    )
    setDecisionFilter(
      "ALL"
    )
    setRiskFilter(
      "ALL"
    )
  }

  function toggleMetadata(
    eventID: string
  ) {
    setExpandedEvents(
      (current) => {
        const next =
          new Set(
            current
          )

        if (
          next.has(
            eventID
          )
        ) {
          next.delete(
            eventID
          )
        } else {
          next.add(
            eventID
          )
        }

        return next
      }
    )
  }

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-zinc-800 bg-zinc-950/40 p-4">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col justify-between gap-3 lg:flex-row lg:items-center">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-600" />

              <input
                value={
                  search
                }
                onChange={(
                  event
                ) =>
                  setSearch(
                    event
                      .target
                      .value
                  )
                }
                placeholder="Search action, resource, policy, agent, session..."
                className="w-full rounded-lg border border-zinc-800 bg-zinc-950 py-2.5 pl-10 pr-3 text-sm text-zinc-200 outline-none placeholder:text-zinc-700 focus:border-zinc-600"
              />
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

          <div className="flex flex-wrap items-center gap-2">
            <Filter className="mr-1 h-4 w-4 text-zinc-600" />

            {(
              [
                "ALL",
                "AGENT",
                "POLICY",
              ] as SourceFilter[]
            ).map(
              (value) => (
                <Button
                  key={
                    value
                  }
                  size="sm"
                  variant={
                    sourceFilter ===
                    value
                      ? "default"
                      : "outline"
                  }
                  onClick={() =>
                    setSourceFilter(
                      value
                    )
                  }
                >
                  {value}
                </Button>
              )
            )}

            <div className="mx-1 h-5 w-px bg-zinc-800" />

            {(
              [
                "ALL",
                "ALLOW",
                "DENY",
                "REQUIRE_APPROVAL",
                "SUCCESS",
              ] as DecisionFilter[]
            ).map(
              (value) => (
                <Button
                  key={
                    value
                  }
                  size="sm"
                  variant={
                    decisionFilter ===
                    value
                      ? "default"
                      : "outline"
                  }
                  onClick={() =>
                    setDecisionFilter(
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

            <div className="mx-1 h-5 w-px bg-zinc-800" />

            {(
              [
                "ALL",
                "ELEVATED",
                "CRITICAL",
              ] as RiskFilter[]
            ).map(
              (value) => (
                <Button
                  key={
                    value
                  }
                  size="sm"
                  variant={
                    riskFilter ===
                    value
                      ? "default"
                      : "outline"
                  }
                  onClick={() =>
                    setRiskFilter(
                      value
                    )
                  }
                >
                  {value}
                </Button>
              )
            )}

            {hasFilters && (
              <Button
                size="sm"
                variant="ghost"
                onClick={
                  clearFilters
                }
              >
                <X className="mr-2 h-4 w-4" />

                Clear
              </Button>
            )}
          </div>

          <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-zinc-600">
            <div>
              Showing{" "}
              <span className="text-zinc-300">
                {
                  filteredEvents.length
                }
              </span>{" "}
              of{" "}
              <span className="text-zinc-300">
                {
                  events.length
                }
              </span>{" "}
              audit events
            </div>

            <div>
              Last update:{" "}
              <span className="text-zinc-400">
                {lastUpdated
                  ? lastUpdated.toLocaleTimeString()
                  : "waiting for stream..."}
              </span>
            </div>
          </div>
        </div>
      </div>

      {filteredEvents.length ===
      0 ? (
        <div className="rounded-xl border border-zinc-800 bg-zinc-950/40 py-16 text-center text-sm text-zinc-500">
          No audit events match
          the selected filters.
        </div>
      ) : (
        <div className="space-y-3">
          {filteredEvents.map(
            (event) => {
              const source =
                eventSource(
                  event
                )

              const policyName =
                metadataText(
                  event,
                  "policy_name"
                )

              const policyID =
                metadataText(
                  event,
                  "policy_id"
                )

              const isExpanded =
                expandedEvents.has(
                  event.id
                )

              const hasMetadata =
                event.metadata &&
                Object.keys(
                  event.metadata
                ).length > 0

              return (
                <div
                  key={
                    event.id
                  }
                  className="rounded-xl border border-zinc-800 bg-zinc-950/50 p-4"
                >
                  <div className="flex flex-col justify-between gap-4 xl:flex-row xl:items-start">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge
                          variant="outline"
                          className="border-zinc-700 bg-zinc-900 text-zinc-400"
                        >
                          {source}
                        </Badge>

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

                        <span className="font-mono text-sm font-semibold text-zinc-200">
                          {
                            event.action
                          }
                        </span>
                      </div>

                      <div className="mt-2 font-mono text-xs text-zinc-500">
                        {
                          event.resource
                        }
                      </div>

                      {policyName &&
                        policyID && (
                          <div className="mt-2 text-sm text-zinc-400">
                            Policy:{" "}
                            <Link
                              href={`/policies/${policyID}`}
                              className="text-zinc-200 underline-offset-4 transition hover:text-white hover:underline"
                            >
                              {
                                policyName
                              }
                            </Link>
                          </div>
                        )}
                    </div>

                    <div className="whitespace-nowrap text-xs text-zinc-500">
                      {new Date(
                        event.created_at
                      ).toLocaleString()}
                    </div>
                  </div>

                  <div className="mt-4 grid gap-4 border-t border-zinc-800 pt-4 md:grid-cols-2 xl:grid-cols-4">
                    <AuditField
                      label="Event Type"
                      value={
                        event.event_type
                      }
                    />

                    <AuditField
                      label="Agent"
                      value={
                        event.agent_id ? (
                          <Link
                            href={`/agents/${event.agent_id}`}
                            className="text-zinc-300 underline-offset-4 hover:text-white hover:underline"
                          >
                            {
                              event.agent_id
                            }
                          </Link>
                        ) : (
                          "Control plane"
                        )
                      }
                    />

                    <AuditField
                      label="Session"
                      value={
                        event.agent_id &&
                        event.session_id ? (
                          <Link
                            href={`/agents/${event.agent_id}/sessions/${event.session_id}`}
                            className="text-zinc-300 underline-offset-4 hover:text-white hover:underline"
                          >
                            {
                              event.session_id
                            }
                          </Link>
                        ) : (
                          "—"
                        )
                      }
                    />

                    <AuditField
                      label="Risk Score"
                      value={String(
                        event.risk_score
                      )}
                    />
                  </div>

                  {hasMetadata && (
                    <div className="mt-4 border-t border-zinc-800 pt-4">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          toggleMetadata(
                            event.id
                          )
                        }
                      >
                        {isExpanded ? (
                          <ChevronUp className="mr-2 h-4 w-4" />
                        ) : (
                          <ChevronDown className="mr-2 h-4 w-4" />
                        )}

                        {isExpanded
                          ? "Hide metadata"
                          : "Show metadata"}
                      </Button>

                      {isExpanded && (
                        <pre className="mt-3 overflow-x-auto whitespace-pre-wrap break-all rounded-lg border border-zinc-800 bg-zinc-950/70 p-4 font-mono text-xs leading-6 text-zinc-500">
                          {JSON.stringify(
                            event.metadata,
                            null,
                            2
                          )}
                        </pre>
                      )}
                    </div>
                  )}
                </div>
              )
            }
          )}
        </div>
      )}
    </div>
  )
}

function AuditField({
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

      <div className="mt-1 break-all font-mono text-xs text-zinc-400">
        {value}
      </div>
    </div>
  )
}