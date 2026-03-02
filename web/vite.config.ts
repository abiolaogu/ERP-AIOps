import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
  server: {
    port: 5179,
  },
  build: {
    chunkSizeWarningLimit: 5000,
    rollupOptions: {
      output: {
        manualChunks(id) {
          const normalizedId = id.split("\\").join("/");
          if (!normalizedId.includes("/node_modules/")) {
            return;
          }
          const packagePath = normalizedId.split("/node_modules/").pop();
          if (!packagePath) {
            return;
          }
          const [scopeOrName, maybeName] = packagePath.split("/");
          const packageName = scopeOrName?.startsWith("@")
            ? [scopeOrName, maybeName].filter(Boolean).join("/")
            : scopeOrName;

          if (
            packageName === "react" ||
            packageName === "react-dom" ||
            packageName === "react-router" ||
            packageName === "react-router-dom" ||
            packageName === "scheduler" ||
            packageName === "use-sync-external-store" ||
            packageName === "@refinedev/core" ||
            packageName === "@refinedev/react-router-v6"
          ) {
            return "vendor-core";
          }
          if (
            packageName === "antd" ||
            packageName?.startsWith("@ant-design/") ||
            packageName?.startsWith("@rc-component/") ||
            packageName?.startsWith("rc-")
          ) {
            return "vendor-antd";
          }
          return;
        },
      },
    },
  },
});
