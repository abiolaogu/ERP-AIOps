import React, { useState } from "react";
import { useList } from "@refinedev/core";
import { Card, Row, Col, Tag, Badge, Select, Empty, Descriptions, List, Typography } from "antd";
import { ApartmentOutlined, ApiOutlined, CheckCircleOutlined, WarningOutlined, CloseCircleOutlined } from "@ant-design/icons";
import { PageHeader } from "@/components/common/PageHeader";

const { Text } = Typography;

interface TopologyNode {
  id: string;
  name: string;
  type: string;
  status: string;
  dependencies: string[];
  metrics: { cpu: number; memory: number; requests_per_sec: number };
}

export const TopologyView: React.FC = () => {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const { data, isLoading } = useList<TopologyNode>({ resource: "topology_nodes" });

  const nodes = data?.data || [];
  const statusIcon = (status: string) => {
    if (status === "healthy") return <CheckCircleOutlined style={{ color: "#52c41a" }} />;
    if (status === "degraded") return <WarningOutlined style={{ color: "#faad14" }} />;
    return <CloseCircleOutlined style={{ color: "#ff4d4f" }} />;
  };

  const selected = nodes.find(n => n.id === selectedNode);

  return (
    <div>
      <PageHeader title="Service Topology" subtitle="Visualize service dependencies and health across the ERP ecosystem" />
      <Row gutter={[16, 16]}>
        <Col span={16}>
          <Card title={<><ApartmentOutlined /> Service Map</>} loading={isLoading}>
            {nodes.length === 0 ? <Empty description="No topology data available" /> : (
              <Row gutter={[12, 12]}>
                {nodes.map(node => (
                  <Col span={6} key={node.id}>
                    <Card size="small" hoverable onClick={() => setSelectedNode(node.id)}
                      style={{ borderColor: selectedNode === node.id ? "#7c3aed" : undefined, cursor: "pointer" }}>
                      <div style={{ textAlign: "center" }}>
                        {statusIcon(node.status)}
                        <div><Text strong>{node.name}</Text></div>
                        <Tag>{node.type}</Tag>
                        <div><Text type="secondary">{node.dependencies?.length || 0} deps</Text></div>
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </Card>
        </Col>
        <Col span={8}>
          <Card title="Node Details">
            {selected ? (
              <Descriptions column={1} size="small">
                <Descriptions.Item label="Name">{selected.name}</Descriptions.Item>
                <Descriptions.Item label="Type"><Tag>{selected.type}</Tag></Descriptions.Item>
                <Descriptions.Item label="Status"><Badge status={selected.status === "healthy" ? "success" : selected.status === "degraded" ? "warning" : "error"} text={selected.status} /></Descriptions.Item>
                <Descriptions.Item label="CPU">{selected.metrics?.cpu}%</Descriptions.Item>
                <Descriptions.Item label="Memory">{selected.metrics?.memory}%</Descriptions.Item>
                <Descriptions.Item label="Req/s">{selected.metrics?.requests_per_sec}</Descriptions.Item>
                <Descriptions.Item label="Dependencies">
                  {selected.dependencies?.map(d => <Tag key={d}>{d}</Tag>) || "None"}
                </Descriptions.Item>
              </Descriptions>
            ) : <Empty description="Select a node to view details" />}
          </Card>
        </Col>
      </Row>
    </div>
  );
};
