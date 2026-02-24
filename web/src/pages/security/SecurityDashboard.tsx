import React from "react";
import { useList } from "@refinedev/core";
import { Card, Row, Col, Table, Tag, Space, Typography, Progress } from "antd";
import { SafetyOutlined, BugOutlined, ExclamationCircleOutlined, CheckCircleOutlined } from "@ant-design/icons";
import { PageHeader } from "@/components/common/PageHeader";
import { KPICard } from "@/components/common/KPICard";
import { StatusBadge } from "@/components/common/StatusBadge";

const { Text } = Typography;

export const SecurityDashboard: React.FC = () => {
  const { data, isLoading } = useList({ resource: "security_findings" });
  const findings = data?.data || [];

  const critical = findings.filter((f: any) => f.severity === "critical").length;
  const high = findings.filter((f: any) => f.severity === "high").length;
  const resolved = findings.filter((f: any) => f.status === "resolved").length;
  const resolvedPct = findings.length > 0 ? Math.round((resolved / findings.length) * 100) : 100;

  const columns = [
    { title: "Finding", dataIndex: "title", key: "title", render: (text: string) => <strong>{text}</strong> },
    { title: "Severity", dataIndex: "severity", key: "severity", render: (sev: string) => {
      const colors: Record<string, string> = { critical: "red", high: "orange", medium: "gold", low: "blue", info: "default" };
      return <Tag color={colors[sev] || "default"}>{sev}</Tag>;
    }},
    { title: "Category", dataIndex: "category", key: "category", render: (cat: string) => <Tag>{cat}</Tag> },
    { title: "Component", dataIndex: "component", key: "component" },
    { title: "Status", dataIndex: "status", key: "status", render: (status: string) => <StatusBadge status={status} /> },
    { title: "Detected", dataIndex: "detected_at", key: "detected_at", render: (d: string) => d ? new Date(d).toLocaleDateString() : "-" },
  ];

  return (
    <div>
      <PageHeader title="Security Scanning" subtitle="View security findings and vulnerability assessments across ERP modules" />
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}><KPICard title="Total Findings" value={findings.length} icon={<SafetyOutlined />} /></Col>
        <Col span={6}><KPICard title="Critical" value={critical} icon={<ExclamationCircleOutlined />} color="#ff4d4f" /></Col>
        <Col span={6}><KPICard title="High" value={high} icon={<BugOutlined />} color="#fa8c16" /></Col>
        <Col span={6}><KPICard title="Resolution Rate" value={`${resolvedPct}%`} icon={<CheckCircleOutlined />} color="#52c41a" /></Col>
      </Row>
      <Card title="Security Findings">
        <Table columns={columns} dataSource={findings} loading={isLoading} rowKey="id" pagination={{ pageSize: 20 }} />
      </Card>
    </div>
  );
};
