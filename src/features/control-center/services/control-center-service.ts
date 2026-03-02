import type { AppSession } from "@/lib/auth/session";
import { requestCosmo } from "@/lib/api/cosmo-client";

export interface ControlCenterCard {
  id: string;
  title: string;
  value: string;
  delta: string;
}

interface ControlCenterQuery {
  control_center_cards?: ControlCenterCard[];
}

export class ControlCenterService {
  static async getCards(session: AppSession): Promise<ControlCenterCard[]> {
    const query = `
      query ControlCenterCards($tenantId: String!) {
        control_center_cards(where: { tenant_id: { _eq: $tenantId } }) {
          id
          title
          value
          delta
        }
      }
    `;

    try {
      const data = await requestCosmo<ControlCenterQuery, { tenantId: string }>(
        query,
        { tenantId: session.tenantId },
        session,
      );
      if (Array.isArray(data.control_center_cards) && data.control_center_cards.length > 0) {
        return data.control_center_cards;
      }
    } catch {
      // Fall through to deterministic local placeholders when shared APIs are unavailable.
    }

    return [
      { id: "rev", title: "Revenue Throughput", value: "$2.4M", delta: "+9.2%" },
      { id: "lat", title: "P95 Request Latency", value: "112ms", delta: "-14ms" },
      { id: "nps", title: "Tenant Satisfaction", value: "74", delta: "+3" },
    ];
  }
}
