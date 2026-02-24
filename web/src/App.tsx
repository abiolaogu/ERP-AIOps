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

// Lazy-loaded pages
const Dashboard = lazy(() => import("@/pages/Dashboard").then(m => ({ default: m.default ?? m.Dashboard })));
const Login = lazy(() => import("@/pages/Login").then(m => ({ default: m.default ?? m.Login })));

const IncidentList = lazy(() => import("@/pages/incidents/IncidentList").then(m => ({ default: m.default ?? m.IncidentList })));
const IncidentShow = lazy(() => import("@/pages/incidents/IncidentShow").then(m => ({ default: m.default ?? m.IncidentShow })));
const IncidentCreate = lazy(() => import("@/pages/incidents/IncidentCreate").then(m => ({ default: m.default ?? m.IncidentCreate })));

const AnomalyList = lazy(() => import("@/pages/anomalies/AnomalyList").then(m => ({ default: m.default ?? m.AnomalyList })));
const AnomalyShow = lazy(() => import("@/pages/anomalies/AnomalyShow").then(m => ({ default: m.default ?? m.AnomalyShow })));

const RuleList = lazy(() => import("@/pages/rules/RuleList").then(m => ({ default: m.default ?? m.RuleList })));
const RuleCreate = lazy(() => import("@/pages/rules/RuleCreate").then(m => ({ default: m.default ?? m.RuleCreate })));

const TopologyView = lazy(() => import("@/pages/topology/TopologyView").then(m => ({ default: m.default ?? m.TopologyView })));

const RemediationList = lazy(() => import("@/pages/remediation/RemediationList").then(m => ({ default: m.default ?? m.RemediationList })));

const CostDashboard = lazy(() => import("@/pages/cost/CostDashboard").then(m => ({ default: m.default ?? m.CostDashboard })));

const SecurityDashboard = lazy(() => import("@/pages/security/SecurityDashboard").then(m => ({ default: m.default ?? m.SecurityDashboard })));

const PageLoader: React.FC = () => (
  <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "60vh" }}>
    <Spin size="large" tip="Loading..." />
  </div>
);

const App: React.FC = () => {
  return (
    <BrowserRouter>
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
