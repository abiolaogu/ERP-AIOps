import React, { useState } from "react";
import { Layout } from "antd";
import { Outlet } from "react-router-dom";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";

const { Content } = Layout;

export const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <Layout
        style={{
          marginLeft: collapsed ? 80 : 260,
          transition: "margin-left 0.2s ease",
        }}
      >
        <Header collapsed={collapsed} onToggle={() => setCollapsed(!collapsed)} />
        <Content
          style={{
            padding: 24,
            minHeight: "calc(100vh - 64px)",
            background: "#f5f7fa",
          }}
        >
          <div className="page-content">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};
