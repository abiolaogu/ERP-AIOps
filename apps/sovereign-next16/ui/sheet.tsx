"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { type ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Sheet({ children }: { children: ReactNode }) {
  return <Dialog.Root>{children}</Dialog.Root>;
}

export function SheetTrigger({ children }: { children: ReactNode }) {
  return <Dialog.Trigger asChild>{children}</Dialog.Trigger>;
}

export function SheetContent({
  children,
  side = "left"
}: {
  children: ReactNode;
  side?: "left" | "right";
}) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="fixed inset-0 z-40 bg-black/30" />
      <Dialog.Content
        className={cn(
          "fixed top-0 z-50 h-full w-80 bg-white p-4 shadow-xl",
          side === "left" ? "left-0" : "right-0"
        )}
      >
        {children}
      </Dialog.Content>
    </Dialog.Portal>
  );
}
