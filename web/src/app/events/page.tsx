import {
  Radio,
} from "lucide-react"

import LiveEventsClient from "@/components/events/LiveEventsClient"
import { getRecentEvents } from "@/lib/api/events"

import { Badge } from "@/components/ui/badge"

export default async function EventsPage() {
  const events =
    await getRecentEvents(100)

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div>
          <div className="flex items-center gap-3">
            <Radio className="h-6 w-6 text-emerald-400" />

            <h2 className="text-2xl font-semibold tracking-tight">
              Live Events
            </h2>
          </div>

          <p className="mt-2 text-sm text-zinc-500">
            Near-real-time autonomous-agent
            authorization and security activity.
          </p>
        </div>

        <Badge
          variant="outline"
          className="w-fit border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
        >
          <span className="mr-2 h-2 w-2 animate-pulse rounded-full bg-emerald-400" />
          STREAMING
        </Badge>
      </div>

      <LiveEventsClient
        initialEvents={events}
      />
    </div>
  )
}