import React, { useState } from "react";
import { Table, Input, Select, Space, Button, Typography, Card } from "antd";
import { PlusOutlined, SearchOutlined, EyeOutlined, AlertOutlined } from "@ant-design/icons";
import { useList } from "@refinedev/core";
import { Link, useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";
import { formatRelativeTime } from "@/utils/formatters";
import { INCIDENT_SEVERITIES, INCIDENT_STATUSES } from "@/utils/constants";
import type { Incident } from "@/types/aiops.types";

const { Text } = Typography;

export const IncidentList: React.FC = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState("");
  const [severityFilter, setSeverityFilter] = useState<string | undefined>();
  const [statusFilter, setStatusFilter] = useState<string | undefined>();
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading } = useList<Incident>({
    resource: "incidents",
    pagination: { current: currentPage, pageSize },
    filters: [
      ...(severityFilter ? [{ field: "severity", operator: "eq" as const, value: severityFilter }] : []),
      ...(statusFilter ? [{ field: "status", operator: "eq" as const, value: statusFilter }] : []),
    ],
  });

  const incidents = data?.data ?? [];
  const total = data?.total ?? 0;

  const filteredIncidents = searchText
    ? incidents.filter(
        (i) =>
          i.title.toLowerCase().includes(searchText.toLowerCase()) ||
          (i.source ?? "").toLowerCase().includes(searchText.toLowerCase()),
      )
    : incidents;

  const columns = [
    {
      title: "Title",
      dataIndex: "title",
      key: "title",
      render: (text: string, record: Incident) => (
        <Link to={`/incidents/${record.id}`} style={{ fontWeight: 500, fontSize: 14 }}>
          {text}
        </Link>
      ),
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      width: 110,
      render: (severity: string) => <StatusBadge status={severity} />,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status: string) => <StatusBadge status={status} />,
    },
    {
      title: "Source",
      dataIndex: "source",
      key: "source",
      render: (source: string) => <Text style={{ fontSize: 13 }}>{source || "-"}</Text>,
    },
    {
      title: "Affected Services",
      dataIndex: "affected_services",
      key: "affected_services",
      render: (services: string[]) => (
        <Text style={{ fontSize: 13 }}>{services?.length ?? 0} service(s)</Text>
      ),
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
    {
      title: "Actions",
      key: "actions",
      width: 80,
      render: (_: unknown, record: Incident) => (
        <Button
          type="text"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/incidents/${record.id}`)}
        />
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Incidents"
        subtitle={`${total} total incidents`}
        breadcrumbs={[
          { label: "Dashboard", path: "/", icon: <AlertOutlined /> },
          { label: "Incidents" },
        ]}
        actions={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate("/incidents/new")}
          >
            Create Incident
          </Button>
        }
      />

      <Card bordered={false}>
        <Space style={{ marginBottom: 16, width: "100%", flexWrap: "wrap" }} size={12}>
          <Input
            prefix={<SearchOutlined style={{ color: "#94a3b8" }} />}
            placeholder="Search incidents..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 260 }}
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
            options={INCIDENT_STATUSES.map((s) => ({ label: s.label, value: s.value }))}
          />
        </Space>

        <Table
          dataSource={filteredIncidents}
          columns={columns}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize,
            total: searchText ? filteredIncidents.length : total,
            showSizeChanger: true,
            showTotal: (t) => `${t} incidents`,
            onChange: (page) => setCurrentPage(page),
          }}
          scroll={{ x: 900 }}
        />
      </Card>
    </div>
  );
};
