import "./globals.css";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import { AppProviders } from "@/core/providers/app-providers";

export const metadata: Metadata = {
  title: "Sovereign ERP-AIOps Standard",
  description: "Next.js 16 + Shadcn + Workik service-layer baseline"
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
