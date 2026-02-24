import React, { useState } from "react";
import { Table, Input, Select, Space, Button, Typography, Card } from "antd";
import { SearchOutlined, EyeOutlined, ExperimentOutlined } from "@ant-design/icons";
import { useList } from "@refinedev/core";
import { Link, useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";
import { formatRelativeTime, formatDeviation } from "@/utils/formatters";
import { INCIDENT_SEVERITIES, ANOMALY_STATUSES } from "@/utils/constants";
import type { Anomaly } from "@/types/aiops.types";

const { Text } = Typography;

export const AnomalyList: React.FC = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState("");
  const [severityFilter, setSeverityFilter] = useState<string | undefined>();
  const [statusFilter, setStatusFilter] = useState<string | undefined>();
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading } = useList<Anomaly>({
    resource: "anomalies",
    pagination: { current: currentPage, pageSize },
    filters: [
      ...(severityFilter ? [{ field: "severity", operator: "eq" as const, value: severityFilter }] : []),
      ...(statusFilter ? [{ field: "status", operator: "eq" as const, value: statusFilter }] : []),
    ],
  });

  const anomalies = data?.data ?? [];
  const total = data?.total ?? 0;

  const filteredAnomalies = searchText
    ? anomalies.filter(
        (a) =>
          a.metric_name.toLowerCase().includes(searchText.toLowerCase()) ||
          a.service.toLowerCase().includes(searchText.toLowerCase()),
      )
    : anomalies;

  const columns = [
    {
      title: "Metric",
      dataIndex: "metric_name",
      key: "metric_name",
      render: (text: string, record: Anomaly) => (
        <Link to={`/anomalies/${record.id}`} style={{ fontWeight: 500, fontSize: 14 }}>
          {text}
        </Link>
      ),
    },
    {
      title: "Service",
      dataIndex: "service",
      key: "service",
      render: (text: string) => <Text style={{ fontSize: 13 }}>{text}</Text>,
    },
    {
      title: "Type",
      dataIndex: "anomaly_type",
      key: "anomaly_type",
      width: 120,
      render: (type: string) => (
        <Text style={{ fontSize: 13 }}>{type.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())}</Text>
      ),
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      width: 100,
      render: (severity: string) => <StatusBadge status={severity} />,
    },
    {
      title: "Deviation",
      dataIndex: "deviation_percent",
      key: "deviation_percent",
      width: 110,
      render: (val: number) => (
        <Text
          style={{
            fontWeight: 600,
            fontSize: 13,
            color: val && Math.abs(val) > 50 ? "#ef4444" : val && Math.abs(val) > 20 ? "#f59e0b" : "#3b82f6",
          }}
        >
          {val != null ? formatDeviation(val) : "-"}
        </Text>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: string) => <StatusBadge status={status} />,
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
    {
      title: "",
      key: "actions",
      width: 60,
      render: (_: unknown, record: Anomaly) => (
        <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/anomalies/${record.id}`)} />
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Anomalies"
        subtitle={`${total} total anomalies detected`}
        breadcrumbs={[
          { label: "Dashboard", path: "/", icon: <ExperimentOutlined /> },
          { label: "Anomalies" },
        ]}
      />

      <Card bordered={false}>
        <Space style={{ marginBottom: 16, width: "100%", flexWrap: "wrap" }} size={12}>
          <Input
            prefix={<SearchOutlined style={{ color: "#94a3b8" }} />}
            placeholder="Search by metric or service..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 280 }}
            allowClear
          />
          <Select
            placeholder="Severity"
            allowClear
            value={severityFilter}
            onChange={setSeverityFilter}
            style={{ width: 140 }}
            options={INCIDENT_SEVERITIES.map((s) => ({ label: s.label, value: s.value }))}
          />
          <Select
            placeholder="Status"
            allowClear
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 160 }}
            options={ANOMALY_STATUSES.map((s) => ({ label: s.label, value: s.value }))}
          />
        </Space>

        <Table
          dataSource={filteredAnomalies}
          columns={columns}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize,
            total: searchText ? filteredAnomalies.length : total,
            showSizeChanger: true,
            showTotal: (t) => `${t} anomalies`,
            onChange: (page) => setCurrentPage(page),
          }}
          scroll={{ x: 1000 }}
        />
      </Card>
    </div>
  );
};
