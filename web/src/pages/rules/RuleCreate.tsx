import React from "react";
import { useCreate } from "@refinedev/core";
import { Form, Input, Select, InputNumber, Switch, Button, Card, Space } from "antd";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/PageHeader";

export const RuleCreate: React.FC = () => {
  const navigate = useNavigate();
  const { mutate, isLoading } = useCreate();
  const [form] = Form.useForm();

  const onFinish = (values: any) => {
    mutate({ resource: "rules", values }, { onSuccess: () => navigate("/rules") });
  };

  return (
    <div>
      <PageHeader title="Create Rule" subtitle="Define a new AIOps detection rule" />
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ enabled: true, rule_type: "threshold" }}>
          <Form.Item name="name" label="Rule Name" rules={[{ required: true }]}><Input placeholder="e.g. high-cpu-usage" /></Form.Item>
          <Form.Item name="rule_type" label="Rule Type" rules={[{ required: true }]}>
            <Select options={[{ label: "Threshold", value: "threshold" }, { label: "Anomaly", value: "anomaly" }, { label: "Correlation", value: "correlation" }, { label: "Remediation", value: "remediation" }]} />
          </Form.Item>
          <Form.Item name="metric" label="Metric"><Input placeholder="e.g. cpu_usage_percent" /></Form.Item>
          <Form.Item name="operator" label="Operator">
            <Select options={[{ label: ">", value: "gt" }, { label: "<", value: "lt" }, { label: "=", value: "eq" }, { label: ">=", value: "gte" }, { label: "<=", value: "lte" }]} />
          </Form.Item>
          <Form.Item name="threshold_value" label="Threshold Value"><InputNumber style={{ width: "100%" }} /></Form.Item>
          <Form.Item name="duration" label="Duration"><Input placeholder="e.g. 5m" /></Form.Item>
          <Form.Item name="severity" label="Severity" rules={[{ required: true }]}>
            <Select options={[{ label: "Critical", value: "critical" }, { label: "High", value: "high" }, { label: "Medium", value: "medium" }, { label: "Low", value: "low" }]} />
          </Form.Item>
          <Form.Item name="action_type" label="Action Type">
            <Select options={[{ label: "Alert", value: "alert" }, { label: "Remediate", value: "remediate" }, { label: "Escalate", value: "escalate" }, { label: "Suppress", value: "suppress" }]} />
          </Form.Item>
          <Form.Item name="enabled" label="Enabled" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item name="description" label="Description"><Input.TextArea rows={3} /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isLoading}>Create Rule</Button>
            <Button onClick={() => navigate("/rules")}>Cancel</Button>
          </Space>
        </Form>
      </Card>
    </div>
  );
};
