import React from "react";
import { Typography, Space, Breadcrumb } from "antd";
import { Link } from "react-router-dom";
import type { BreadcrumbItem } from "@/types/common.types";

const { Title, Text } = Typography;

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  breadcrumbs?: BreadcrumbItem[];
  actions?: React.ReactNode;
  extra?: React.ReactNode;
}

export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  subtitle,
  breadcrumbs,
  actions,
  extra,
}) => {
  return (
    <div style={{ marginBottom: 24 }}>
      {breadcrumbs && breadcrumbs.length > 0 && (
        <Breadcrumb
          style={{ marginBottom: 12 }}
          items={breadcrumbs.map((item) => ({
            title: item.path ? (
              <Link to={item.path}>
                <Space size={4}>
                  {item.icon}
                  {item.label}
                </Space>
              </Link>
            ) : (
              <Space size={4}>
                {item.icon}
                {item.label}
              </Space>
            ),
          }))}
        />
      )}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          flexWrap: "wrap",
          gap: 16,
        }}
      >
        <div>
          <Title level={3} style={{ margin: 0, fontWeight: 700 }}>
            {title}
          </Title>
          {subtitle && (
            <Text type="secondary" style={{ fontSize: 14, marginTop: 4 }}>
              {subtitle}
            </Text>
          )}
        </div>
        {actions && <Space wrap>{actions}</Space>}
      </div>
      {extra && <div style={{ marginTop: 16 }}>{extra}</div>}
    </div>
  );
};
