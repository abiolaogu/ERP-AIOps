import React from "react";
import { Row, Col, Card, Table, Typography, Space } from "antd";
import {
  AlertOutlined,
  ExperimentOutlined,
  SettingOutlined,
  ThunderboltOutlined,
  DollarOutlined,
  SafetyOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import { useList } from "@refinedev/core";
import { Link } from "react-router-dom";
import { KPICard } from "@/components/common/KPICard";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";
import { formatRelativeTime } from "@/utils/formatters";
import type { Incident, Anomaly, SecurityFinding } from "@/types/aiops.types";

const { Text } = Typography;

export const Dashboard: React.FC = () => {
  const { data: incidentsData, isLoading: incidentsLoading } = useList<Incident>({
    resource: "incidents",
    pagination: { current: 1, pageSize: 100 },
  });

  const { data: anomaliesData, isLoading: anomaliesLoading } = useList<Anomaly>({
    resource: "anomalies",
    pagination: { current: 1, pageSize: 100 },
  });

  const { data: securityData } = useList<SecurityFinding>({
    resource: "security_findings",
    pagination: { current: 1, pageSize: 100 },
  });

  const incidents = incidentsData?.data ?? [];
  const anomalies = anomaliesData?.data ?? [];
  const securityFindings = securityData?.data ?? [];

  // Calculate KPIs
  const openIncidents = incidents.filter((i) =>
    ["open", "acknowledged", "investigating"].includes(i.status),
  ).length;

  const activeAnomalies = anomalies.filter((a) =>
    ["active", "investigating"].includes(a.status),
  ).length;

  const rulesEnabled = 12; // placeholder
  const autoRemediations = 8; // placeholder
  const costSavings = "$24.5K"; // placeholder

  const openFindings = securityFindings.filter((f) =>
    ["open"].includes(f.status),
  ).length;

  // Recent incidents
  const recentIncidents = [...incidents]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5);

  const incidentColumns = [
    {
      title: "Title",
      dataIndex: "title",
      key: "title",
      render: (text: string, record: Incident) => (
        <Link to={`/incidents/${record.id}`} style={{ fontWeight: 500 }}>
          {text}
        </Link>
      ),
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      width: 100,
      render: (severity: string) => <StatusBadge status={severity} size="small" />,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: string) => <StatusBadge status={status} size="small" />,
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      width: 140,
      render: (date: string) => (
        <Text type="secondary" style={{ fontSize: 13 }}>
          {formatRelativeTime(date)}
        </Text>
      ),
    },
  ];

  // Recent anomalies
  const recentAnomalies = [...anomalies]
    .sort((a, b) => new Date(b.detected_at).getTime() - new Date(a.detected_at).getTime())
    .slice(0, 5);

  const anomalyColumns = [
    {
      title: "Metric",
      dataIndex: "metric_name",
      key: "metric_name",
      render: (text: string) => <Text style={{ fontWeight: 500 }}>{text}</Text>,
    },
    {
      title: "Service",
      dataIndex: "service",
      key: "service",
      render: (text: string) => <Text style={{ fontSize: 13 }}>{text}</Text>,
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      width: 100,
      render: (severity: string) => <StatusBadge status={severity} size="small" />,
    },
    {
      title: "Detected",
      dataIndex: "detected_at",
      key: "detected_at",
      width: 140,
      render: (date: string) => (
        <Text type="secondary" style={{ fontSize: 13 }}>
          {formatRelativeTime(date)}
        </Text>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Welcome back! Here's your AIOps operations overview."
      />

      {/* KPI Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Open Incidents"
            value={openIncidents}
            icon={<AlertOutlined />}
            color="#ef4444"
            trend={{ value: 12.5, isPositive: false }}
            loading={incidentsLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Active Anomalies"
            value={activeAnomalies}
            icon={<ExperimentOutlined />}
            color="#f59e0b"
            trend={{ value: 8.3, isPositive: false }}
            loading={anomaliesLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Rules Enabled"
            value={rulesEnabled}
            icon={<SettingOutlined />}
            color="#7c3aed"
            trend={{ value: 2, isPositive: true }}
          />
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Auto-Remediations"
            value={autoRemediations}
            icon={<ThunderboltOutlined />}
            color="#3b82f6"
            trend={{ value: 15, isPositive: true }}
          />
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Cost Savings"
            value={costSavings}
            icon={<DollarOutlined />}
            color="#10b981"
            trend={{ value: 22.1, isPositive: true }}
          />
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <KPICard
            title="Security Findings"
            value={openFindings}
            icon={<SafetyOutlined />}
            color="#f97316"
            trend={{ value: 5.2, isPositive: false }}
          />
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* Recent Incidents */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <AlertOutlined />
                <span>Recent Incidents</span>
              </Space>
            }
            extra={
              <Link to="/incidents">
                <Space>
                  <Text type="secondary" style={{ fontSize: 13 }}>View All</Text>
                  <ArrowRightOutlined style={{ fontSize: 12, color: "#94a3b8" }} />
                </Space>
              </Link>
            }
            bordered={false}
          >
            <Table
              dataSource={recentIncidents}
              columns={incidentColumns}
              rowKey="id"
              pagination={false}
              loading={incidentsLoading}
              size="middle"
              locale={{ emptyText: "No incidents recorded" }}
            />
          </Card>
        </Col>

        {/* Recent Anomalies */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <ExperimentOutlined />
                <span>Recent Anomalies</span>
              </Space>
            }
            extra={
              <Link to="/anomalies">
                <Space>
                  <Text type="secondary" style={{ fontSize: 13 }}>View All</Text>
                  <ArrowRightOutlined style={{ fontSize: 12, color: "#94a3b8" }} />
                </Space>
              </Link>
            }
            bordered={false}
          >
            <Table
              dataSource={recentAnomalies}
              columns={anomalyColumns}
              rowKey="id"
              pagination={false}
              loading={anomaliesLoading}
              size="middle"
              locale={{ emptyText: "No anomalies detected" }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};
