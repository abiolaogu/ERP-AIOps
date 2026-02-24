import React from "react";
import { Tag } from "antd";

interface StatusConfig {
  label: string;
  color: string;
}

const STATUS_MAP: Record<string, StatusConfig> = {
  // Incident statuses
  open: { label: "Open", color: "#ef4444" },
  acknowledged: { label: "Acknowledged", color: "#f59e0b" },
  investigating: { label: "Investigating", color: "#3b82f6" },
  resolved: { label: "Resolved", color: "#10b981" },
  closed: { label: "Closed", color: "#64748b" },

  // Anomaly statuses
  active: { label: "Active", color: "#ef4444" },
  dismissed: { label: "Dismissed", color: "#94a3b8" },

  // Severity levels
  critical: { label: "Critical", color: "#ef4444" },
  high: { label: "High", color: "#f97316" },
  medium: { label: "Medium", color: "#f59e0b" },
  low: { label: "Low", color: "#3b82f6" },
  info: { label: "Info", color: "#94a3b8" },

  // Remediation statuses
  pending: { label: "Pending", color: "#f59e0b" },
  running: { label: "Running", color: "#3b82f6" },
  completed: { label: "Completed", color: "#10b981" },
  failed: { label: "Failed", color: "#ef4444" },
  cancelled: { label: "Cancelled", color: "#94a3b8" },

  // Topology statuses
  healthy: { label: "Healthy", color: "#10b981" },
  degraded: { label: "Degraded", color: "#f59e0b" },
  down: { label: "Down", color: "#ef4444" },
  unknown: { label: "Unknown", color: "#94a3b8" },

  // Boolean
  enabled: { label: "Enabled", color: "#10b981" },
  disabled: { label: "Disabled", color: "#94a3b8" },
};

interface StatusBadgeProps {
  status: string;
  label?: string;
  size?: "small" | "default";
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  label,
  size = "default",
}) => {
  const config = STATUS_MAP[status] ?? {
    label: status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()),
    color: "#94a3b8",
  };

  const displayLabel = label ?? config.label;

  return (
    <Tag
      color={config.color}
      style={{
        borderRadius: 6,
        fontWeight: 500,
        fontSize: size === "small" ? 11 : 12,
        padding: size === "small" ? "0 6px" : "2px 10px",
        border: "none",
      }}
    >
      {displayLabel}
    </Tag>
  );
};

interface SeverityBadgeProps {
  severity: string;
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  return <StatusBadge status={severity} />;
};
