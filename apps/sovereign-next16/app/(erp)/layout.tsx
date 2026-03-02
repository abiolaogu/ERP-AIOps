import type { ReactNode } from "react";
import { ERPLayout } from "@/components/erp-layout";

export default function PlatformLayout({ children }: { children: ReactNode }) {
  return <ERPLayout>{children}</ERPLayout>;
}
