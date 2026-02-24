import React from "react";
import { useList } from "@refinedev/core";
import { Card, Row, Col, Table, Tag, Statistic, Progress, Typography } from "antd";
import { DollarOutlined, RiseOutlined, FallOutlined, ThunderboltOutlined } from "@ant-design/icons";
import { PageHeader } from "@/components/common/PageHeader";
import { KPICard } from "@/components/common/KPICard";

const { Text } = Typography;

export const CostDashboard: React.FC = () => {
  const { data, isLoading } = useList({ resource: "cost_reports" });
  const reports = data?.data || [];

  const totalCost = reports.reduce((sum: number, r: any) => sum + (r.total_cost || 0), 0);
  const optimizableCost = reports.reduce((sum: number, r: any) => sum + (r.optimizable_cost || 0), 0);
  const savingsPercent = totalCost > 0 ? ((optimizableCost / totalCost) * 100).toFixed(1) : "0";

  const columns = [
    { title: "Service", dataIndex: "service_name", key: "service_name", render: (text: string) => <strong>{text}</strong> },
    { title: "Category", dataIndex: "category", key: "category", render: (cat: string) => <Tag>{cat}</Tag> },
    { title: "Monthly Cost", dataIndex: "total_cost", key: "total_cost", render: (cost: number) => <Text>${cost?.toFixed(2)}</Text> },
    { title: "Optimizable", dataIndex: "optimizable_cost", key: "optimizable_cost", render: (cost: number) => <Text type={cost > 0 ? "warning" : "secondary"}>${cost?.toFixed(2)}</Text> },
    { title: "Efficiency", dataIndex: "efficiency", key: "efficiency", render: (eff: number) => <Progress percent={eff || 0} size="small" style={{ width: 100 }} strokeColor={eff > 80 ? "#52c41a" : eff > 50 ? "#faad14" : "#ff4d4f"} /> },
    { title: "Recommendation", dataIndex: "recommendation", key: "recommendation", render: (rec: string) => rec || "-" },
  ];

  return (
    <div>
      <PageHeader title="Cost Optimization" subtitle="Analyze and optimize infrastructure costs across the ERP ecosystem" />
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}><KPICard title="Total Monthly Cost" value={`$${totalCost.toFixed(2)}`} icon={<DollarOutlined />} /></Col>
        <Col span={6}><KPICard title="Potential Savings" value={`$${optimizableCost.toFixed(2)}`} icon={<FallOutlined />} color="#52c41a" /></Col>
        <Col span={6}><KPICard title="Savings Potential" value={`${savingsPercent}%`} icon={<ThunderboltOutlined />} color="#faad14" /></Col>
        <Col span={6}><KPICard title="Services Analyzed" value={reports.length} icon={<RiseOutlined />} /></Col>
      </Row>
      <Card title="Cost Breakdown by Service">
        <Table columns={columns} dataSource={reports} loading={isLoading} rowKey="id" pagination={{ pageSize: 20 }} />
      </Card>
    </div>
  );
};
