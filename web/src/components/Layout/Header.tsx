import React from "react";
import {
  Layout,
  Input,
  Badge,
  Avatar,
  Dropdown,
  Space,
  Typography,
} from "antd";
import {
  SearchOutlined,
  BellOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from "@ant-design/icons";
import { useAuth } from "@/hooks/useAuth";
import { getInitials, getAvatarColor } from "@/utils/formatters";

const { Header: AntHeader } = Layout;
const { Text } = Typography;

interface HeaderProps {
  collapsed: boolean;
  onToggle: () => void;
}

export const Header: React.FC<HeaderProps> = ({ collapsed, onToggle }) => {
  const { user, logout } = useAuth();

  const userName = user?.name ?? "Admin User";
  const userRole = user?.role ?? "Administrator";

  const userMenuItems = [
    {
      key: "profile",
      icon: <UserOutlined />,
      label: "Profile",
    },
    {
      key: "settings",
      icon: <SettingOutlined />,
      label: "Settings",
    },
    {
      type: "divider" as const,
    },
    {
      key: "logout",
      icon: <LogoutOutlined />,
      label: "Sign Out",
      danger: true,
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    if (key === "logout") {
      logout();
    }
  };

  return (
    <AntHeader
      style={{
        background: "#fff",
        padding: "0 24px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        borderBottom: "1px solid #f0f0f0",
        height: 64,
        position: "sticky",
        top: 0,
        zIndex: 99,
        boxShadow: "0 1px 4px rgba(0, 0, 0, 0.04)",
      }}
    >
      <Space size={16} align="center">
        <div
          onClick={onToggle}
          style={{ cursor: "pointer", fontSize: 18, color: "#64748b" }}
        >
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </div>
        <Input
          prefix={<SearchOutlined style={{ color: "#94a3b8" }} />}
          placeholder="Search incidents, anomalies, rules..."
          style={{
            width: 320,
            borderRadius: 8,
            backgroundColor: "#f5f7fa",
            border: "1px solid #e2e8f0",
          }}
          allowClear
        />
      </Space>

      <Space size={20} align="center">
        <Badge count={5} size="small">
          <BellOutlined
            style={{ fontSize: 20, color: "#64748b", cursor: "pointer" }}
          />
        </Badge>

        <Dropdown
          menu={{ items: userMenuItems, onClick: handleMenuClick }}
          placement="bottomRight"
          trigger={["click"]}
        >
          <Space
            style={{ cursor: "pointer", marginLeft: 8 }}
            align="center"
          >
            <Avatar
              size={36}
              style={{
                backgroundColor: getAvatarColor(userName),
                fontWeight: 600,
                fontSize: 13,
              }}
            >
              {getInitials(userName)}
            </Avatar>
            <div style={{ lineHeight: 1.3 }}>
              <Text style={{ fontWeight: 600, fontSize: 13, display: "block" }}>
                {userName}
              </Text>
              <Text
                type="secondary"
                style={{ fontSize: 11, display: "block" }}
              >
                {userRole}
              </Text>
            </div>
          </Space>
        </Dropdown>
      </Space>
    </AntHeader>
  );
};
