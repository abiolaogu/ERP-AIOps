import { Card, CardContent, CardHeader, CardTitle } from "@/ui/card";

export default function DashboardPage() {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader>
          <CardTitle>Throughput</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold">2.4M</p>
          <p className="text-sm text-[var(--muted)]">events/min via Redpanda mesh</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Tenants</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold">1,248</p>
          <p className="text-sm text-[var(--muted)]">isolated by policy and topic namespaces</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Replica Health</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold">99.999%</p>
          <p className="text-sm text-[var(--muted)]">Yugabyte cross-region availability target</p>
        </CardContent>
      </Card>
    </div>
  );
}
