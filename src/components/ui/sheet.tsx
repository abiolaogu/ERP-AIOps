"use client";

import * as React from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

export const Sheet = Dialog.Root;
export const SheetTrigger = Dialog.Trigger;

export function SheetContent({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="fixed inset-0 z-40 bg-black/40" />
      <Dialog.Content className={cn("fixed inset-y-0 left-0 z-50 w-[82vw] max-w-sm border-r bg-[var(--card)] p-4 shadow-lg", className)}>
        {children}
        <Dialog.Close className="absolute right-3 top-3 rounded p-1 hover:bg-[var(--muted)]" aria-label="Close">
          <X className="h-4 w-4" />
        </Dialog.Close>
      </Dialog.Content>
    </Dialog.Portal>
  );
}
