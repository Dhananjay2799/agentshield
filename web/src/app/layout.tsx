import type { Metadata } from "next"
import { Geist, Geist_Mono } from "next/font/google"

import AppShell from "@/components/soc/AppShell"

import "./globals.css"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "AgentShield SOC",
  description:
    "Zero-trust security operations center for autonomous AI agents",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} min-h-screen bg-zinc-950 font-sans text-zinc-100 antialiased`}
      >
        <AppShell>{children}</AppShell>
      </body>
    </html>
  )
}