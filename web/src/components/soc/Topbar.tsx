import { Badge } from "@/components/ui/badge"

export default function Topbar() {
  return (
    <header className="flex h-16 items-center justify-between border-b border-zinc-800 bg-zinc-950 px-6">
      <div>
        <h1 className="text-sm font-semibold text-zinc-100 sm:text-base">
          Security Operations Center
        </h1>

        <p className="hidden text-xs text-zinc-500 sm:block">
          Autonomous agent activity and threat monitoring
        </p>
      </div>

      <Badge
        variant="outline"
        className="border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
      >
        <span className="mr-2 h-2 w-2 rounded-full bg-emerald-400" />
        LIVE
      </Badge>
    </header>
  )
}