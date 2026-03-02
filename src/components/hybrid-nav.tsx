"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, LayoutDashboard, Shield, Settings } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Sidebar, SidebarContent, SidebarHeader } from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";

const items = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/login", label: "Access", icon: Shield },
  { href: "/dashboard?panel=settings", label: "Settings", icon: Settings },
];

function NavItems() {
  const pathname = usePathname();
  return (
    <>
      {items.map((item) => {
        const Icon = item.icon;
        const active = pathname === item.href;
        return (
          <Button key={item.href} asChild variant={active ? "default" : "ghost"} className="w-full justify-start">
            <Link href={item.href}>
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          </Button>
        );
      })}
    </>
  );
}

export function HybridNavLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className="flex min-h-screen">
      <Sidebar className={cn("hidden md:block transition-all", collapsed ? "w-16" : "w-64")}>
        <SidebarHeader>
          <button onClick={() => setCollapsed((value) => !value)} type="button" className="w-full text-left">
            {collapsed ? "ERP" : "Sovereign ERP"}
          </button>
        </SidebarHeader>
        <SidebarContent>
          <NavItems />
        </SidebarContent>
      </Sidebar>

      <div className="flex min-h-screen flex-1 flex-col">
        <header className="flex h-14 items-center justify-between border-b bg-[var(--card)] px-4 md:px-6">
          <div className="flex items-center gap-2 md:hidden">
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="outline" size="icon">
                  <Menu className="h-4 w-4" />
                </Button>
              </SheetTrigger>
              <SheetContent>
                <div className="mt-8 space-y-2">
                  <NavItems />
                </div>
              </SheetContent>
            </Sheet>
            <span className="text-sm font-semibold">Sovereign ERP</span>
          </div>
          <p className="hidden text-sm text-[var(--muted-foreground)] md:block">
            Next.js 16 + Shadcn + Workik feature slices
          </p>
        </header>
        <main className="flex-1 p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}
