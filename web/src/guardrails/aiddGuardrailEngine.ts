export type GuardrailMode = "autonomous" | "supervised" | "protected";

export type GuardrailAction = {
  action: string;
  tenantId: string;
  confidence: number;
  blastRadius: number;
  estimatedCostUsd: number;
  crossTenant?: boolean;
  privilegeEscalation?: boolean;
  destructive?: boolean;
};

export type GuardrailDecision = {
  allowed: boolean;
  mode: GuardrailMode;
  requiresApproval: boolean;
  reasons: string[];
};

export const defaultPolicy = {
  autonomous: { minConfidence: 0.84, maxBlastRadius: 400, maxEstimatedCostUsd: 5000 },
  supervised: { minConfidence: 0.72, maxBlastRadius: 5000, maxEstimatedCostUsd: 75000 },
};

export function evaluateGuardrail(action: GuardrailAction): GuardrailDecision {
  const reasons: string[] = [];

  if (!action.tenantId) {
    return { allowed: false, mode: "protected", requiresApproval: false, reasons: ["Missing tenant context"] };
  }

  if (action.crossTenant) reasons.push("Cross-tenant access denied");
  if (action.privilegeEscalation) reasons.push("Privilege escalation denied");
  if (action.destructive && action.confidence < 0.9) reasons.push("Destructive operation below confidence threshold");
  if (reasons.length > 0) {
    return { allowed: false, mode: "protected", requiresApproval: false, reasons };
  }

  if (
    action.confidence >= defaultPolicy.autonomous.minConfidence &&
    action.blastRadius <= defaultPolicy.autonomous.maxBlastRadius &&
    action.estimatedCostUsd <= defaultPolicy.autonomous.maxEstimatedCostUsd
  ) {
    return { allowed: true, mode: "autonomous", requiresApproval: false, reasons: ["Autonomous execution allowed"] };
  }

  if (
    action.confidence >= defaultPolicy.supervised.minConfidence &&
    action.blastRadius <= defaultPolicy.supervised.maxBlastRadius &&
    action.estimatedCostUsd <= defaultPolicy.supervised.maxEstimatedCostUsd
  ) {
    return { allowed: true, mode: "supervised", requiresApproval: true, reasons: ["Approval required"] };
  }

  return { allowed: false, mode: "protected", requiresApproval: false, reasons: ["Risk exceeds supervised guardrail"] };
}
