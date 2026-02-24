import React from "react";
import {
  Card, Descriptions, Tabs, Tag, Space, Typography, Timeline, Row, Col, Button, Statistic, Empty,
} from "antd";
import {
  ArrowLeftOutlined, AlertOutlined, ClockCircleOutlined, CheckCircleOutlined,
  ExclamationCircleOutlined, ThunderboltOutlined,
} from "@ant-design/icons";
import { useOne } from "@refinedev/core";
import { useParams, useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";
import { formatDate, formatDateTime, formatRelativeTime } from "@/utils/formatters";
import type { Incident } from "@/types/aiops.types";

const { Text, Title, Paragraph } = Typography;

export const IncidentShow: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: incidentData, isLoading } = useOne<Incident>({
    resource: "incidents",
    id: id!,
  });

  const incident = incidentData?.data;

  if (!incident && !isLoading) {
    return <Empty description="Incident not found" />;
  }

  const timelineItems = [
    {
      dot: <ExclamationCircleOutlined style={{ color: "#ef4444" }} />,
      children: (
        <div>
          <Text style={{ fontWeight: 500 }}>Incident Created</Text>
          <div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {incident?.created_at ? formatDateTime(incident.created_at) : "-"}
            </Text>
          </div>
        </div>
      ),
    },
    ...(incident?.acknowledged_at
      ? [
          {
            dot: <ClockCircleOutlined style={{ color: "#f59e0b" }} />,
            children: (
              <div>
                <Text style={{ fontWeight: 500 }}>Acknowledged</Text>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    by {incident.acknowledged_by ?? "System"} - {formatDateTime(incident.acknowledged_at)}
                  </Text>
                </div>
              </div>
            ),
          },
        ]
      : []),
    ...(incident?.resolved_at
      ? [
          {
            dot: <CheckCircleOutlined style={{ color: "#10b981" }} />,
            children: (
              <div>
                <Text style={{ fontWeight: 500 }}>Resolved</Text>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    by {incident.resolved_by ?? "System"} - {formatDateTime(incident.resolved_at)}
                  </Text>
                </div>
              </div>
            ),
          },
        ]
      : []),
  ];

  const tabItems = [
    {
      key: "overview",
      label: "Overview",
      children: (
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={16}>
            <Card title="Incident Details" size="small" bordered={false}>
              <Descriptions column={{ xs: 1, sm: 2 }} size="small">
                <Descriptions.Item label="Title">{incident?.title}</Descriptions.Item>
                <Descriptions.Item label="Severity">
                  <StatusBadge status={incident?.severity ?? "medium"} />
                </Descriptions.Item>
                <Descriptions.Item label="Status">
                  <StatusBadge status={incident?.status ?? "open"} />
                </Descriptions.Item>
                <Descriptions.Item label="Source">{incident?.source ?? "-"}</Descriptions.Item>
                <Descriptions.Item label="Created">
                  {incident?.created_at ? formatDate(incident.created_at) : "-"}
                </Descriptions.Item>
                <Descriptions.Item label="Last Updated">
                  {incident?.updated_at ? formatRelativeTime(incident.updated_at) : "-"}
                </Descriptions.Item>
                <Descriptions.Item label="Affected Services" span={2}>
                  <Space wrap>
                    {incident?.affected_services?.map((svc) => (
                      <Tag key={svc} color="purple">{svc}</Tag>
                    )) ?? <Text type="secondary">None</Text>}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="Description" span={2}>
                  <Paragraph style={{ margin: 0 }}>
                    {incident?.description ?? "No description provided."}
                  </Paragraph>
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          <Col xs={24} lg={8}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>
              <Card size="small" bordered={false}>
                <Statistic
                  title="Affected Services"
                  value={incident?.affected_services?.length ?? 0}
                  valueStyle={{ color: "#7c3aed" }}
                />
              </Card>
              <Card size="small" bordered={false}>
                <Statistic
                  title="Time Open"
                  value={incident?.created_at ? formatRelativeTime(incident.created_at) : "-"}
                  valueStyle={{ color: "#f59e0b", fontSize: 18 }}
                />
              </Card>
            </Space>
          </Col>
        </Row>
      ),
    },
    {
      key: "root_cause",
      label: "Root Cause Analysis",
      children: (
        <Card bordered={false}>
          {incident?.root_cause ? (
            <div>
              <Title level={5}>Identified Root Cause</Title>
              <Paragraph>{incident.root_cause}</Paragraph>
            </div>
          ) : (
            <Empty description="Root cause analysis not yet available. Click 'Analyze' to trigger AI analysis." />
          )}
        </Card>
      ),
    },
    {
      key: "timeline",
      label: "Timeline",
      children: (
        <Card bordered={false}>
          <Timeline items={timelineItems} />
        </Card>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title=""
        breadcrumbs={[
          { label: "Dashboard", path: "/" },
          { label: "Incidents", path: "/incidents" },
          { label: incident?.title ?? "Loading..." },
        ]}
        actions={
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/incidents")}>
              Back
            </Button>
            <Button type="primary" icon={<ThunderboltOutlined />}>
              Trigger Remediation
            </Button>
          </Space>
        }
      />

      {/* Incident Header Card */}
      <Card bordered={false} loading={isLoading} style={{ marginBottom: 16 }}>
        <Space size={16} align="start">
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: 12,
              background: "#ef444415",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <AlertOutlined style={{ fontSize: 24, color: "#ef4444" }} />
          </div>
          <div>
            <Title level={4} style={{ margin: 0 }}>
              {incident?.title}
            </Title>
            <Space style={{ marginTop: 8 }}>
              <StatusBadge status={incident?.severity ?? "medium"} />
              <StatusBadge status={incident?.status ?? "open"} />
              <Text type="secondary" style={{ fontSize: 12 }}>
                Created {incident?.created_at ? formatRelativeTime(incident.created_at) : ""}
              </Text>
            </Space>
          </div>
        </Space>
      </Card>

      <Card bordered={false}>
        <Tabs items={tabItems} defaultActiveKey="overview" />
      </Card>
    </div>
  );
};
