import React from "react";
import { Card, Descriptions, Space, Typography, Row, Col, Button, Statistic, Empty } from "antd";
import { ArrowLeftOutlined, ExperimentOutlined } from "@ant-design/icons";
import { useOne } from "@refinedev/core";
import { useParams, useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { StatusBadge } from "@/components/common/StatusBadge";
import { formatDate, formatRelativeTime, formatDeviation } from "@/utils/formatters";
import type { Anomaly } from "@/types/aiops.types";

const { Text, Title, Paragraph } = Typography;

export const AnomalyShow: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: anomalyData, isLoading } = useOne<Anomaly>({
    resource: "anomalies",
    id: id!,
  });

  const anomaly = anomalyData?.data;

  if (!anomaly && !isLoading) {
    return <Empty description="Anomaly not found" />;
  }

  return (
    <div>
      <PageHeader
        title=""
        breadcrumbs={[
          { label: "Dashboard", path: "/" },
          { label: "Anomalies", path: "/anomalies" },
          { label: anomaly?.metric_name ?? "Loading..." },
        ]}
        actions={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/anomalies")}>
            Back
          </Button>
        }
      />

      {/* Anomaly Header Card */}
      <Card bordered={false} loading={isLoading} style={{ marginBottom: 16 }}>
        <Space size={16} align="start">
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: 12,
              background: "#f59e0b15",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <ExperimentOutlined style={{ fontSize: 24, color: "#f59e0b" }} />
          </div>
          <div>
            <Title level={4} style={{ margin: 0 }}>
              {anomaly?.metric_name}
            </Title>
            <Space style={{ marginTop: 8 }}>
              <StatusBadge status={anomaly?.severity ?? "medium"} />
              <StatusBadge status={anomaly?.status ?? "active"} />
              <Text type="secondary" style={{ fontSize: 12 }}>
                on {anomaly?.service} - detected {anomaly?.detected_at ? formatRelativeTime(anomaly.detected_at) : ""}
              </Text>
            </Space>
          </div>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="Anomaly Details" bordered={false}>
            <Descriptions column={{ xs: 1, sm: 2 }} size="small">
              <Descriptions.Item label="Metric Name">{anomaly?.metric_name}</Descriptions.Item>
              <Descriptions.Item label="Service">{anomaly?.service}</Descriptions.Item>
              <Descriptions.Item label="Module">{anomaly?.module ?? "-"}</Descriptions.Item>
              <Descriptions.Item label="Anomaly Type">
                {anomaly?.anomaly_type?.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())}
              </Descriptions.Item>
              <Descriptions.Item label="Severity">
                <StatusBadge status={anomaly?.severity ?? "medium"} />
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <StatusBadge status={anomaly?.status ?? "active"} />
              </Descriptions.Item>
              <Descriptions.Item label="Detected">
                {anomaly?.detected_at ? formatDate(anomaly.detected_at) : "-"}
              </Descriptions.Item>
              <Descriptions.Item label="Resolved">
                {anomaly?.resolved_at ? formatDate(anomaly.resolved_at) : "Not yet resolved"}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <Card size="small" bordered={false}>
              <Statistic
                title="Expected Value"
                value={anomaly?.expected_value ?? 0}
                precision={2}
                valueStyle={{ color: "#10b981" }}
              />
            </Card>
            <Card size="small" bordered={false}>
              <Statistic
                title="Actual Value"
                value={anomaly?.actual_value ?? 0}
                precision={2}
                valueStyle={{ color: "#ef4444" }}
              />
            </Card>
            <Card size="small" bordered={false}>
              <Statistic
                title="Deviation"
                value={anomaly?.deviation_percent ? formatDeviation(anomaly.deviation_percent) : "N/A"}
                valueStyle={{
                  color: anomaly?.deviation_percent && Math.abs(anomaly.deviation_percent) > 50
                    ? "#ef4444"
                    : "#f59e0b",
                  fontSize: 24,
                }}
              />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
};
