import React from "react";
import { Layout, Menu, Typography } from "antd";
import {
  DashboardOutlined,
  AlertOutlined,
  ExperimentOutlined,
  SettingOutlined,
  ApartmentOutlined,
  ThunderboltOutlined,
  DollarOutlined,
  SafetyOutlined,
} from "@ant-design/icons";
import { useNavigate, useLocation } from "react-router-dom";

const { Sider } = Layout;
const { Title } = Typography;

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

const menuItems = [
  {
    key: "/",
    icon: <DashboardOutlined />,
    label: "Dashboard",
  },
  {
    key: "/incidents",
    icon: <AlertOutlined />,
    label: "Incidents",
  },
  {
    key: "/anomalies",
    icon: <ExperimentOutlined />,
    label: "Anomalies",
  },
  {
    key: "/rules",
    icon: <SettingOutlined />,
    label: "Rules",
  },
  {
    key: "/topology",
    icon: <ApartmentOutlined />,
    label: "Topology",
  },
  {
    key: "/remediation",
    icon: <ThunderboltOutlined />,
    label: "Remediation",
  },
  {
    key: "/cost",
    icon: <DollarOutlined />,
    label: "Cost",
  },
  {
    key: "/security",
    icon: <SafetyOutlined />,
    label: "Security",
  },
];

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, onCollapse }) => {
  const navigate = useNavigate();
  const location = useLocation();

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path === "/") return "/";
    if (path.startsWith("/incidents")) return "/incidents";
    if (path.startsWith("/anomalies")) return "/anomalies";
    if (path.startsWith("/rules")) return "/rules";
    if (path.startsWith("/topology")) return "/topology";
    if (path.startsWith("/remediation")) return "/remediation";
    if (path.startsWith("/cost")) return "/cost";
    if (path.startsWith("/security")) return "/security";
    return path;
  };

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={onCollapse}
      width={260}
      style={{
        overflow: "auto",
        height: "100vh",
        position: "fixed",
        left: 0,
        top: 0,
        bottom: 0,
        zIndex: 100,
      }}
      theme="dark"
    >
      <div
        style={{
          height: 64,
          display: "flex",
          alignItems: "center",
          justifyContent: collapsed ? "center" : "flex-start",
          padding: collapsed ? "0" : "0 24px",
          borderBottom: "1px solid rgba(255,255,255,0.08)",
        }}
      >
        <div
          style={{
            width: 32,
            height: 32,
            borderRadius: 8,
            background: "linear-gradient(135deg, #7c3aed, #6d28d9)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          <span style={{ color: "#fff", fontWeight: 700, fontSize: 13 }}>
            AI
          </span>
        </div>
        {!collapsed && (
          <Title
            level={5}
            style={{
              color: "#fff",
              margin: "0 0 0 12px",
              whiteSpace: "nowrap",
              fontWeight: 600,
            }}
          >
            ERP AIOps
          </Title>
        )}
      </div>

      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[getSelectedKey()]}
        items={menuItems}
        onClick={({ key }) => {
          navigate(key);
        }}
        style={{ borderRight: 0, marginTop: 8 }}
      />
    </Sider>
  );
};
