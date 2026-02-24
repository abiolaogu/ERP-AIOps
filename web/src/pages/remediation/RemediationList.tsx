import React from "react";
import { useList } from "@refinedev/core";
import { Table, Tag, Space, Button, Progress, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";

const { Text } = Typography;

export const RemediationList: React.FC = () => {
  const navigate = useNavigate();
  const { data, isLoading } = useList({ resource: "remediation_actions" });

  const columns = [
    { title: "Action", dataIndex: "name", key: "name", render: (text: string) => <strong>{text}</strong> },
    { title: "Type", dataIndex: "action_type", key: "action_type", render: (type: string) => <Tag color={type === "auto" ? "green" : "blue"}>{type}</Tag> },
    { title: "Incident", dataIndex: "incident_id", key: "incident_id", render: (id: string) => id ? <Button type="link" size="small" onClick={() => navigate(`/incidents/${id}`)}>{id.slice(0, 8)}</Button> : "-" },
    { title: "Status", dataIndex: "status", key: "status", render: (status: string) => <StatusBadge status={status} /> },
    { title: "Progress", dataIndex: "progress", key: "progress", render: (p: number) => <Progress percent={p || 0} size="small" style={{ width: 120 }} /> },
    { title: "Duration", dataIndex: "duration_ms", key: "duration_ms", render: (ms: number) => ms ? `${(ms / 1000).toFixed(1)}s` : "-" },
    { title: "Executed At", dataIndex: "executed_at", key: "executed_at", render: (d: string) => d ? new Date(d).toLocaleString() : "-" },
  ];

  return (
    <div>
      <PageHeader title="Auto-Remediation" subtitle="View and manage automated remediation actions" />
      <Table columns={columns} dataSource={data?.data || []} loading={isLoading} rowKey="id" pagination={{ pageSize: 20 }} />
    </div>
  );
};
