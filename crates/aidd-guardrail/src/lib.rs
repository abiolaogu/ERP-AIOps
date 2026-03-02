use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum GuardrailMode {
    Autonomous,
    Supervised,
    Protected,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct GuardrailAction {
    pub action: String,
    pub tenant_id: String,
    pub confidence: f64,
    pub blast_radius: u32,
    pub estimated_cost_usd: f64,
    pub cross_tenant: bool,
    pub privilege_escalation: bool,
    pub destructive: bool,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct GuardrailDecision {
    pub allowed: bool,
    pub mode: GuardrailMode,
    pub requires_approval: bool,
    pub reasons: Vec<String>,
}

#[derive(Debug, Clone, Copy)]
pub struct GuardrailPolicy {
    pub autonomous_min_confidence: f64,
    pub autonomous_max_blast_radius: u32,
    pub autonomous_max_cost_usd: f64,
    pub supervised_min_confidence: f64,
    pub supervised_max_blast_radius: u32,
    pub supervised_max_cost_usd: f64,
}

impl Default for GuardrailPolicy {
    fn default() -> Self {
        Self {
            autonomous_min_confidence: 0.84,
            autonomous_max_blast_radius: 400,
            autonomous_max_cost_usd: 5000.0,
            supervised_min_confidence: 0.72,
            supervised_max_blast_radius: 5000,
            supervised_max_cost_usd: 75000.0,
        }
    }
}

pub fn evaluate(policy: GuardrailPolicy, action: &GuardrailAction) -> GuardrailDecision {
    let mut reasons = Vec::new();

    if action.tenant_id.trim().is_empty() {
        return GuardrailDecision {
            allowed: false,
            mode: GuardrailMode::Protected,
            requires_approval: false,
            reasons: vec!["missing tenant context".to_string()],
        };
    }

    if action.cross_tenant {
        reasons.push("cross-tenant access denied".to_string());
    }
    if action.privilege_escalation {
        reasons.push("privilege escalation denied".to_string());
    }
    if action.destructive && action.confidence < 0.90 {
        reasons.push("destructive operation below confidence threshold".to_string());
    }
    if !reasons.is_empty() {
        return GuardrailDecision {
            allowed: false,
            mode: GuardrailMode::Protected,
            requires_approval: false,
            reasons,
        };
    }

    if action.confidence >= policy.autonomous_min_confidence
        && action.blast_radius <= policy.autonomous_max_blast_radius
        && action.estimated_cost_usd <= policy.autonomous_max_cost_usd
    {
        return GuardrailDecision {
            allowed: true,
            mode: GuardrailMode::Autonomous,
            requires_approval: false,
            reasons: vec!["autonomous execution allowed".to_string()],
        };
    }

    if action.confidence >= policy.supervised_min_confidence
        && action.blast_radius <= policy.supervised_max_blast_radius
        && action.estimated_cost_usd <= policy.supervised_max_cost_usd
    {
        return GuardrailDecision {
            allowed: true,
            mode: GuardrailMode::Supervised,
            requires_approval: true,
            reasons: vec!["approval required".to_string()],
        };
    }

    GuardrailDecision {
        allowed: false,
        mode: GuardrailMode::Protected,
        requires_approval: false,
        reasons: vec!["risk exceeds supervised guardrail".to_string()],
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn blocks_cross_tenant() {
        let policy = GuardrailPolicy::default();
        let action = GuardrailAction {
            action: "tenant.export".to_string(),
            tenant_id: "tenant-a".to_string(),
            confidence: 0.99,
            blast_radius: 1,
            estimated_cost_usd: 100.0,
            cross_tenant: true,
            privilege_escalation: false,
            destructive: false,
        };
        let decision = evaluate(policy, &action);
        assert!(!decision.allowed);
        assert_eq!(decision.mode, GuardrailMode::Protected);
    }

    #[test]
    fn requires_approval_for_medium_risk() {
        let policy = GuardrailPolicy::default();
        let action = GuardrailAction {
            action: "bulk.update".to_string(),
            tenant_id: "tenant-a".to_string(),
            confidence: 0.80,
            blast_radius: 1200,
            estimated_cost_usd: 12000.0,
            cross_tenant: false,
            privilege_escalation: false,
            destructive: false,
        };
        let decision = evaluate(policy, &action);
        assert!(decision.allowed);
        assert_eq!(decision.mode, GuardrailMode::Supervised);
        assert!(decision.requires_approval);
    }

    #[test]
    fn allows_autonomous_for_safe_actions() {
        let policy = GuardrailPolicy::default();
        let action = GuardrailAction {
            action: "refresh.dashboard".to_string(),
            tenant_id: "tenant-a".to_string(),
            confidence: 0.93,
            blast_radius: 40,
            estimated_cost_usd: 20.0,
            cross_tenant: false,
            privilege_escalation: false,
            destructive: false,
        };
        let decision = evaluate(policy, &action);
        assert!(decision.allowed);
        assert_eq!(decision.mode, GuardrailMode::Autonomous);
        assert!(!decision.requires_approval);
    }
}
