import type { ReactNode } from "react"

import Sidebar from "@/components/soc/Sidebar"
import Topbar from "@/components/soc/Topbar"

type AppShellProps = {
  children: ReactNode
}

export default function AppShell({
  children,
}: AppShellProps) {
  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100">
      <div className="grid min-h-screen lg:grid-cols-[250px_1fr]">
        <Sidebar />

        <div className="min-w-0">
          <Topbar />

          <main className="min-w-0">
            {children}
          </main>
        </div>
      </div>
    </div>
  )
}