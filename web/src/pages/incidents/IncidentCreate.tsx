import React from "react";
import { Card, Form, Input, Select, Button, Space, Row, Col, Divider, message } from "antd";
import { ArrowLeftOutlined, SaveOutlined } from "@ant-design/icons";
import { useCreate } from "@refinedev/core";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";
import { INCIDENT_SEVERITIES } from "@/utils/constants";

export const IncidentCreate: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const { mutate: createIncident, isLoading } = useCreate();

  const handleSubmit = (values: Record<string, unknown>) => {
    createIncident(
      {
        resource: "incidents",
        values: {
          ...values,
          tenant_id: "default",
          status: "open",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      },
      {
        onSuccess: () => {
          message.success("Incident created successfully!");
          navigate("/incidents");
        },
        onError: () => {
          message.error("Failed to create incident");
        },
      },
    );
  };

  return (
    <div>
      <PageHeader
        title="Create Incident"
        breadcrumbs={[
          { label: "Dashboard", path: "/" },
          { label: "Incidents", path: "/incidents" },
          { label: "New Incident" },
        ]}
        actions={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/incidents")}>
            Back
          </Button>
        }
      />

      <Card bordered={false}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ severity: "medium" }}
          requiredMark="optional"
        >
          <Divider titlePlacement="start" plain>
            Incident Information
          </Divider>
          <Row gutter={16}>
            <Col xs={24} sm={16}>
              <Form.Item
                label="Title"
                name="title"
                rules={[{ required: true, message: "Title is required" }]}
              >
                <Input placeholder="Incident title" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item
                label="Severity"
                name="severity"
                rules={[{ required: true }]}
              >
                <Select
                  options={INCIDENT_SEVERITIES.map((s) => ({
                    label: s.label,
                    value: s.value,
                  }))}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24}>
              <Form.Item label="Description" name="description">
                <Input.TextArea rows={4} placeholder="Describe the incident..." />
              </Form.Item>
            </Col>
          </Row>

          <Divider titlePlacement="start" plain>
            Impact & Source
          </Divider>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item label="Source" name="source">
                <Select
                  placeholder="Select source"
                  options={[
                    { label: "Monitoring Alert", value: "monitoring" },
                    { label: "User Report", value: "user_report" },
                    { label: "Automated Detection", value: "automated" },
                    { label: "External", value: "external" },
                    { label: "Manual", value: "manual" },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item label="Affected Services" name="affected_services">
                <Select
                  mode="tags"
                  placeholder="Add affected services"
                  options={[
                    { label: "API Gateway", value: "api-gateway" },
                    { label: "Auth Service", value: "auth-service" },
                    { label: "Database", value: "database" },
                    { label: "Cache", value: "cache" },
                    { label: "Message Queue", value: "message-queue" },
                    { label: "Frontend", value: "frontend" },
                  ]}
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={isLoading}
              size="large"
            >
              Create Incident
            </Button>
            <Button size="large" onClick={() => navigate("/incidents")}>
              Cancel
            </Button>
          </Space>
        </Form>
      </Card>
    </div>
  );
};
