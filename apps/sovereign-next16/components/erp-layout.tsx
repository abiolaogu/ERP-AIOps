"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { LayoutDashboard, Menu, Package, Users } from "lucide-react";
import { Sidebar, SidebarContent, SidebarHeader } from "@/ui/sidebar";
import { Sheet, SheetContent, SheetTrigger } from "@/ui/sheet";
import { Button } from "@/ui/button";

const navItems = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/inventory", label: "Inventory", icon: Package },
  { href: "/tenants", label: "Tenants", icon: Users }
];

function NavItems() {
  return (
    <>
      {navItems.map((item) => (
        <Link key={item.href} href={item.href} className="block">
          <Button variant="ghost" className="w-full justify-start gap-2">
            <item.icon size={18} />
            {item.label}
          </Button>
        </Link>
      ))}
    </>
  );
}

export function ERPLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar className="hidden border-r border-[var(--border)] md:flex">
        <SidebarHeader className="p-4 text-base font-semibold">ERP-AIOps</SidebarHeader>
        <SidebarContent>
          <NavItems />
        </SidebarContent>
      </Sidebar>

      <div className="flex flex-1 flex-col">
        <header className="flex h-16 items-center gap-3 border-b border-[var(--border)] bg-white px-4">
          <Sheet>
            <SheetTrigger>
              <Button variant="outline" className="md:hidden" aria-label="Open navigation">
                <Menu size={18} />
              </Button>
            </SheetTrigger>
            <SheetContent side="left">
              <div className="mb-4 border-b border-[var(--border)] pb-3 text-base font-semibold">ERP-AIOps</div>
              <div className="space-y-1">
                <NavItems />
              </div>
            </SheetContent>
          </Sheet>
          <h1 className="text-sm font-semibold md:text-base">ERP Standard v2026</h1>
        </header>
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>
  );
}
