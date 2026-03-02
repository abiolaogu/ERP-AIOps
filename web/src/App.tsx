import React, { Suspense, lazy } from "react";
import { Refine } from "@refinedev/core";
import { ConfigProvider, Spin } from "antd";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import routerBindings from "@refinedev/react-router-v6";
import { dataProvider } from "@/providers/hasuraDataProvider";
import { liveProvider } from "@/providers/hasuraLiveProvider";
import { authProvider } from "@/authProvider";
import { theme } from "@/theme";
import { MainLayout } from "@/components/Layout/MainLayout";

const lazyNamed = (importFn: () => Promise<any>, name: string) =>
  lazy(() => importFn().then((m) => ({ default: m[name] ?? m.default })));

// Lazy-loaded pages
const Dashboard = lazyNamed(() => import("@/pages/Dashboard"), "Dashboard");
const Login = lazyNamed(() => import("@/pages/Login"), "Login");
const IncidentList = lazyNamed(() => import("@/pages/incidents/IncidentList"), "IncidentList");
const IncidentShow = lazyNamed(() => import("@/pages/incidents/IncidentShow"), "IncidentShow");
const IncidentCreate = lazyNamed(() => import("@/pages/incidents/IncidentCreate"), "IncidentCreate");
const AnomalyList = lazyNamed(() => import("@/pages/anomalies/AnomalyList"), "AnomalyList");
const AnomalyShow = lazyNamed(() => import("@/pages/anomalies/AnomalyShow"), "AnomalyShow");
const RuleList = lazyNamed(() => import("@/pages/rules/RuleList"), "RuleList");
const RuleCreate = lazyNamed(() => import("@/pages/rules/RuleCreate"), "RuleCreate");
const TopologyView = lazyNamed(() => import("@/pages/topology/TopologyView"), "TopologyView");
const RemediationList = lazyNamed(() => import("@/pages/remediation/RemediationList"), "RemediationList");
const CostDashboard = lazyNamed(() => import("@/pages/cost/CostDashboard"), "CostDashboard");
const SecurityDashboard = lazyNamed(() => import("@/pages/security/SecurityDashboard"), "SecurityDashboard");

const PageLoader: React.FC = () => (
  <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "60vh" }}>
    <Spin size="large" tip="Loading..." />
  </div>
);

const App: React.FC = () => {
  return (
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <ConfigProvider theme={theme}>
        <Refine
          routerProvider={routerBindings}
          dataProvider={dataProvider}
          liveProvider={liveProvider}
          authProvider={authProvider}
          resources={[
            {
              name: "dashboard",
              list: "/",
              meta: { label: "Dashboard", icon: "dashboard" },
            },
            {
              name: "incidents",
              list: "/incidents",
              show: "/incidents/:id",
              create: "/incidents/create",
              meta: { label: "Incidents" },
            },
            {
              name: "anomalies",
              list: "/anomalies",
              show: "/anomalies/:id",
              meta: { label: "Anomalies" },
            },
            {
              name: "rules",
              list: "/rules",
              create: "/rules/create",
              meta: { label: "Rules" },
            },
            {
              name: "topology_nodes",
              list: "/topology",
              meta: { label: "Topology" },
            },
            {
              name: "remediation_actions",
              list: "/remediation",
              meta: { label: "Remediation" },
            },
            {
              name: "cost_reports",
              list: "/cost",
              meta: { label: "Cost Optimization" },
            },
            {
              name: "security_findings",
              list: "/security",
              meta: { label: "Security" },
            },
          ]}
          options={{
            syncWithLocation: true,
            warnWhenUnsavedChanges: true,
            liveMode: "auto",
          }}
        >
          <Suspense fallback={<PageLoader />}>
            <Routes>
              {/* Public routes */}
              <Route path="/login" element={<Login />} />

              {/* Authenticated routes */}
              <Route element={<MainLayout />}>
                <Route index element={<Dashboard />} />
                <Route path="/incidents" element={<IncidentList />} />
                <Route path="/incidents/:id" element={<IncidentShow />} />
                <Route path="/incidents/create" element={<IncidentCreate />} />
                <Route path="/anomalies" element={<AnomalyList />} />
                <Route path="/anomalies/:id" element={<AnomalyShow />} />
                <Route path="/rules" element={<RuleList />} />
                <Route path="/rules/create" element={<RuleCreate />} />
                <Route path="/topology" element={<TopologyView />} />
                <Route path="/remediation" element={<RemediationList />} />
                <Route path="/cost" element={<CostDashboard />} />
                <Route path="/security" element={<SecurityDashboard />} />
              </Route>

              {/* Catch-all redirect */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </Refine>
      </ConfigProvider>
    </BrowserRouter>
  );
};

export default App;
