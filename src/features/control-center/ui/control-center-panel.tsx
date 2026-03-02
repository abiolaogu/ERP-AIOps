"use client";

import { motion } from "framer-motion";
import { Skeleton } from "@/components/ui/skeleton";
import { useControlCenter } from "@/features/control-center/hooks/use-control-center";

export function ControlCenterPanel() {
  const { data, isLoading } = useControlCenter();

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-3">
        <Skeleton className="h-32" />
        <Skeleton className="h-32" />
        <Skeleton className="h-32" />
      </div>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {data?.map((card, index) => (
        <motion.article
          key={card.id}
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.24, delay: index * 0.06 }}
          className="rounded-xl border bg-[var(--card)] p-4 shadow-sm"
        >
          <p className="text-xs uppercase tracking-wide text-[var(--muted-foreground)]">{card.title}</p>
          <p className="mt-2 text-2xl font-semibold">{card.value}</p>
          <p className="mt-1 text-sm text-[var(--muted-foreground)]">{card.delta}</p>
        </motion.article>
      ))}
    </div>
  );
}
