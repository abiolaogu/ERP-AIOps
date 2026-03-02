import { HybridNavLayout } from "@/components/hybrid-nav";
import { ControlCenterPanel } from "@/features/control-center/ui/control-center-panel";
import { modulePolicy } from "@/lib/config/module-policy";

export default function DashboardPage() {
  return (
    <HybridNavLayout>
      <section className="space-y-5">
        <header className="space-y-1">
          <h1 className="text-2xl font-semibold">{modulePolicy.name} Control Center</h1>
          <p className="text-sm text-[var(--muted-foreground)]">
            Workik service to TanStack Query hook to Shadcn UI with realtime invalidation.
          </p>
        </header>
        <ControlCenterPanel />
      </section>
    </HybridNavLayout>
  );
}
