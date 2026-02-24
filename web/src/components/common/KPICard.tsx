import React from "react";
import { Card, Statistic, Typography, Space } from "antd";
import { ArrowUpOutlined, ArrowDownOutlined } from "@ant-design/icons";
import type { KPIData } from "@/types/common.types";

const { Text } = Typography;

interface KPICardProps extends KPIData {
  loading?: boolean;
}

export const KPICard: React.FC<KPICardProps> = ({
  title,
  value,
  prefix,
  suffix,
  trend,
  icon,
  color = "#7c3aed",
  loading = false,
}) => {
  return (
    <Card className="kpi-card" loading={loading} bordered={false}>
      <Space
        direction="vertical"
        size={4}
        style={{ width: "100%" }}
      >
        <Space align="center" style={{ justifyContent: "space-between", width: "100%" }}>
          <Text
            type="secondary"
            style={{ fontSize: 13, fontWeight: 500, textTransform: "uppercase", letterSpacing: 0.5 }}
          >
            {title}
          </Text>
          {icon && (
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 10,
                backgroundColor: `${color}15`,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color,
                fontSize: 18,
              }}
            >
              {icon}
            </div>
          )}
        </Space>
        <Statistic
          value={typeof value === "string" ? undefined : value}
          formatter={typeof value === "string" ? () => value : undefined}
          prefix={prefix}
          suffix={suffix}
          valueStyle={{
            fontWeight: 700,
            fontSize: 28,
            color: "#1a1a2e",
          }}
        />
        {trend && (
          <Space size={4}>
            {trend.isPositive ? (
              <ArrowUpOutlined style={{ color: "#10b981", fontSize: 12 }} />
            ) : (
              <ArrowDownOutlined style={{ color: "#ef4444", fontSize: 12 }} />
            )}
            <Text
              style={{
                color: trend.isPositive ? "#10b981" : "#ef4444",
                fontSize: 13,
                fontWeight: 500,
              }}
            >
              {trend.value}%
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              vs last month
            </Text>
          </Space>
        )}
      </Space>
    </Card>
  );
};
