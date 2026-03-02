import { cn } from "@/lib/utils";

export function Sidebar({ className, children }: { className?: string; children: React.ReactNode }) {
  return <aside className={cn("border-r bg-[var(--card)]", className)}>{children}</aside>;
}

export function SidebarHeader({ children }: { children: React.ReactNode }) {
  return <div className="border-b px-4 py-3 text-sm font-semibold uppercase tracking-wide">{children}</div>;
}

export function SidebarContent({ children }: { children: React.ReactNode }) {
  return <div className="space-y-2 p-3">{children}</div>;
}
