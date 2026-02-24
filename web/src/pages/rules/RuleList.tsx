import React from "react";
import { useList } from "@refinedev/core";
import { Table, Tag, Switch, Space, Button, Input } from "antd";
import { PlusOutlined, SearchOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";

export const RuleList: React.FC = () => {
  const navigate = useNavigate();
  const { data, isLoading } = useList({ resource: "rules" });

  const columns = [
    { title: "Name", dataIndex: "name", key: "name", render: (text: string) => <strong>{text}</strong> },
    { title: "Type", dataIndex: "rule_type", key: "rule_type", render: (type: string) => <Tag color={type === "anomaly" ? "purple" : type === "threshold" ? "blue" : type === "correlation" ? "orange" : "green"}>{type}</Tag> },
    { title: "Severity", dataIndex: "severity", key: "severity", render: (sev: string) => <StatusBadge status={sev} /> },
    { title: "Enabled", dataIndex: "enabled", key: "enabled", render: (enabled: boolean) => <Switch checked={enabled} size="small" /> },
    { title: "Match Count", dataIndex: "match_count", key: "match_count", render: (count: number) => count || 0 },
    { title: "Last Evaluated", dataIndex: "last_evaluated_at", key: "last_evaluated_at", render: (d: string) => d ? new Date(d).toLocaleString() : "Never" },
    {
      title: "Actions", key: "actions",
      render: (_: any, record: any) => (
        <Space><Button size="small" onClick={() => navigate(`/rules/${record.id}`)}>View</Button></Space>
      ),
    },
  ];

  return (
    <div>
      <PageHeader title="AIOps Rules" subtitle="Manage anomaly detection, threshold, and correlation rules"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => navigate("/rules/create")}>Create Rule</Button>} />
      <Table columns={columns} dataSource={data?.data || []} loading={isLoading} rowKey="id" pagination={{ pageSize: 20 }} />
    </div>
  );
};
