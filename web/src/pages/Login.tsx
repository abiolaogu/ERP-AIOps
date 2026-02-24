import React from "react";
import { Card, Form, Input, Button, Typography, Space, Divider } from "antd";
import { MailOutlined, LockOutlined } from "@ant-design/icons";
import { useLogin } from "@refinedev/core";

const { Title, Text } = Typography;

export const Login: React.FC = () => {
  const { mutate: login, isLoading } = useLogin();

  const handleSubmit = (values: { email: string; password: string }) => {
    login(values);
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%)",
        padding: 24,
      }}
    >
      <Card
        bordered={false}
        style={{
          width: 420,
          borderRadius: 16,
          boxShadow: "0 20px 60px rgba(0, 0, 0, 0.15)",
        }}
        bodyStyle={{ padding: 40 }}
      >
        <div style={{ textAlign: "center", marginBottom: 32 }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 14,
              background: "linear-gradient(135deg, #7c3aed, #6d28d9)",
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              marginBottom: 16,
            }}
          >
            <span
              style={{
                color: "#fff",
                fontWeight: 700,
                fontSize: 18,
              }}
            >
              AI
            </span>
          </div>
          <Title level={3} style={{ margin: 0 }}>
            Welcome Back
          </Title>
          <Text type="secondary">Sign in to ERP AIOps</Text>
        </div>

        <Form
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            email: "admin@erp.com",
            password: "password",
          }}
          requiredMark={false}
        >
          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: "Please enter your email" },
              { type: "email", message: "Please enter a valid email" },
            ]}
          >
            <Input
              prefix={<MailOutlined style={{ color: "#94a3b8" }} />}
              placeholder="your@email.com"
              size="large"
            />
          </Form.Item>

          <Form.Item
            label="Password"
            name="password"
            rules={[
              { required: true, message: "Please enter your password" },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: "#94a3b8" }} />}
              placeholder="Enter your password"
              size="large"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 12 }}>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              block
              loading={isLoading}
              style={{
                height: 48,
                fontWeight: 600,
                fontSize: 15,
                borderRadius: 10,
              }}
            >
              Sign In
            </Button>
          </Form.Item>
        </Form>

        <Divider plain>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Demo credentials pre-filled
          </Text>
        </Divider>

        <div style={{ textAlign: "center" }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            ERP AIOps &copy; 2026. All rights reserved.
          </Text>
        </div>
      </Card>
    </div>
  );
};
