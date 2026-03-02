import { type HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export function Sidebar({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <aside className={cn("w-72 bg-white", className)} {...props} />;
}

export function SidebarHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("border-b border-[var(--border)]", className)} {...props} />;
}

export function SidebarContent({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <nav className={cn("space-y-1 p-3", className)} {...props} />;
}
