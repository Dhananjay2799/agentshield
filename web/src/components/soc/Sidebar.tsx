"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Bot,
  LayoutDashboard,
  Radio,
  ScrollText,
  ShieldCheck,
  Siren,
  UserCheck,
} from "lucide-react"

const navItems = [
  {
    label: "Overview",
    href: "/",
    icon: LayoutDashboard,
  },
  {
    label: "Incidents",
    href: "/incidents",
    icon: Siren,
  },
  {
    label: "Agents",
    href: "/agents",
    icon: Bot,
  },
  {
    label: "Approvals",
    href: "/approvals",
    icon: UserCheck,
  },
  {
    label: "Live Events",
    href: "/events",
    icon: Radio,
  },
  {
    label: "Policies",
    href: "/policies",
    icon: ShieldCheck,
  },
  {
    label: "Audit Logs",
    href: "/audit",
    icon: ScrollText,
  },
]

export default function Sidebar() {
  const pathname = usePathname()

  function isActive(href: string) {
    if (href === "/") {
      return pathname === "/"
    }

    return pathname.startsWith(href)
  }

  return (
    <aside className="hidden min-h-screen border-r border-zinc-800 bg-zinc-950 lg:block">
      <div className="flex h-16 items-center border-b border-zinc-800 px-5">
        <Link href="/">
          <div className="text-lg font-semibold tracking-tight text-zinc-100">
            AgentShield
          </div>

          <div className="text-xs text-zinc-500">
            Security Control Plane
          </div>
        </Link>
      </div>

      <nav className="space-y-1 p-3">
        {navItems.map((item) => {
          const Icon = item.icon
          const active = isActive(item.href)

          return (
            <Link
              key={item.href}
              href={item.href}
              className={
                active
                  ? "flex items-center gap-3 rounded-lg bg-zinc-800 px-3 py-2 text-sm font-medium text-white"
                  : "flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-zinc-400 transition hover:bg-zinc-900 hover:text-white"
              }
            >
              <Icon className="h-4 w-4" />

              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>

      <div className="mx-3 mt-6 rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
        <div className="flex items-center gap-2 text-sm font-medium text-zinc-200">
          <span className="h-2 w-2 rounded-full bg-emerald-400" />
          System Operational
        </div>

        <div className="mt-4 space-y-3 text-xs">
          <div className="flex items-center justify-between">
            <span className="text-zinc-500">Gateway</span>
            <span className="text-emerald-400">Online</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-zinc-500">OPA</span>
            <span className="text-emerald-400">Online</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-zinc-500">Kafka</span>
            <span className="text-emerald-400">Online</span>
          </div>
        </div>
      </div>
    </aside>
  )
}