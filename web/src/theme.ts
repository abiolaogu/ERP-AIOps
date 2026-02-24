import type { ThemeConfig } from "antd";

export const theme: ThemeConfig = {
  token: {
    colorPrimary: "#7c3aed",
    colorLink: "#7c3aed",
    colorSuccess: "#10b981",
    colorWarning: "#f59e0b",
    colorError: "#ef4444",
    colorInfo: "#3b82f6",
    borderRadius: 10,
    fontFamily:
      '"Plus Jakarta Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontSize: 14,
    colorBgContainer: "#ffffff",
    colorBgLayout: "#f5f7fa",
    controlHeight: 40,
  },
  components: {
    Layout: {
      siderBg: "#001529",
      headerBg: "#ffffff",
      bodyBg: "#f5f7fa",
    },
    Menu: {
      darkItemBg: "#001529",
      darkItemSelectedBg: "#7c3aed",
      darkItemHoverBg: "rgba(124, 58, 237, 0.6)",
      itemBorderRadius: 8,
      darkItemColor: "rgba(255, 255, 255, 0.75)",
      darkItemSelectedColor: "#ffffff",
    },
    Card: {
      borderRadiusLG: 10,
      paddingLG: 24,
    },
    Table: {
      borderRadiusLG: 10,
      headerBg: "#fafbfc",
      headerColor: "#64748b",
    },
    Button: {
      borderRadius: 8,
      controlHeight: 40,
      fontWeight: 500,
    },
    Input: {
      borderRadius: 8,
      controlHeight: 40,
    },
    Select: {
      borderRadius: 8,
      controlHeight: 40,
    },
    DatePicker: {
      borderRadius: 8,
      controlHeight: 40,
    },
    Tag: {
      borderRadiusSM: 6,
    },
    Statistic: {
      contentFontSize: 28,
    },
  },
};
