# Business Model -- Sovereign AIOps

**Confidential -- Series A Investment Memorandum**

---

## 1. Revenue Model: Per-Resource Pricing

### 1.1 Pricing Structure

Sovereign AIOps uses a **per-resource-per-month** pricing model. A "resource" is any monitored entity: microservice, database, cache, message queue, API gateway, or cloud resource.

| Tier | Price/Resource/Month | Included Capabilities |
|---|---|---|
| **Starter** | $2 | Event ingestion (100K events/sec), anomaly detection (Isolation Forest + LSTM), log anomaly detection, noise reduction (dedup + suppression), basic dashboards |
| **Professional** | $5 | All Starter + event correlation engine, topology auto-discovery, SLO management, change risk scoring, incident timeline, post-mortem generator |
| **Enterprise** | $8 | All Professional + autonomous remediation (runbook executor), capacity planning (ML forecasting), chaos engineering integration, custom ML model tuning, SSO/SAML, dedicated support |

### 1.2 Pricing Rationale

- **Per-resource aligns value with usage** -- Customers pay proportionally to infrastructure complexity
- **Low entry point** -- $2/resource lets small teams start immediately (200 resources = $400/month)
- **Natural expansion** -- As infrastructure grows, revenue grows without renegotiation
- **Competitive positioning** -- Datadog charges $15-33/host/month; we are significantly cheaper per unit while providing more intelligence
- **Land-and-expand** -- Start with Starter tier, upgrade to Professional/Enterprise as trust builds

### 1.3 Revenue Per Customer

| Segment | Avg Resources | Tier Mix | Monthly Revenue | Annual Revenue |
|---|---|---|---|---|
| SMB (50-200 employees) | 50 | 80% Starter, 20% Pro | $150 | $1,800 |
| Mid-Market (200-2K) | 300 | 30% Starter, 50% Pro, 20% Ent | $1,350 | $16,200 |
| Enterprise (2K+) | 2,000 | 10% Starter, 40% Pro, 50% Ent | $12,400 | $148,800 |

**Blended average ARR per customer:** $19,200 (Year 1 mix), scaling to $97,100 by Year 5 as enterprise mix increases.

---

## 2. Unit Economics

### 2.1 Customer Economics

| Metric | Year 1 | Year 3 (Target) | Year 5 (Target) |
|---|---|---|---|
| Average Contract Value (ACV) | $19,200 | $48,000 | $97,100 |
| Customer Acquisition Cost (CAC) | $38,000 | $35,000 | $32,000 |
| CAC Payback Period | 24 months | 9 months | 4 months |
| Gross Margin | 72% | 82% | 85% |
| Net Revenue Retention (NRR) | 115% | 135% | 145% |
| Logo Churn | 15% | 8% | 5% |
| LTV (5-year) | $72,000 | $288,000 | $582,000 |
| LTV:CAC Ratio | 1.9:1 | 8.2:1 | 18.2:1 |

### 2.2 Gross Margin Breakdown

| Cost Component | % of Revenue | Notes |
|---|---|---|
| Cloud infrastructure (compute, storage) | 15% | Kafka, PostgreSQL, ClickHouse, ML inference |
| ML training compute | 5% | Daily model retraining per tenant |
| Third-party APIs | 2% | PagerDuty, Slack APIs |
| Support engineering | 6% | Technical support team |
| **Total COGS** | **28%** | |
| **Gross Margin** | **72%** | Improves to 85% at scale (infrastructure amortization) |

### 2.3 Net Revenue Retention Drivers

- **Resource growth:** Customer infrastructure grows 20-30% annually
- **Tier upgrades:** Customers move from Starter -> Professional -> Enterprise
- **Feature adoption:** New capabilities (chaos engineering, capacity planning) drive upgrades
- **Target NRR: 135%+** by Year 3 (comparable to Datadog's 130%+ NRR)

---

## 3. Go-to-Market Strategy

### 3.1 Sales Motion

| Phase | Timeline | Strategy |
|---|---|---|
| **Phase 1: Product-Led Growth** | Months 1-12 | Free tier (10 resources), self-serve onboarding, developer community, content marketing |
| **Phase 2: Inside Sales** | Months 6-18 | SDR team qualifying PLG leads, demo-based sales for mid-market |
| **Phase 3: Enterprise Sales** | Months 12-24 | Named account AEs, solutions engineers, POC-based sales |

### 3.2 Channel Strategy

- **Direct sales:** Primary channel for Professional and Enterprise tiers
- **Cloud marketplaces:** AWS Marketplace, GCP Marketplace (instant procurement)
- **Technology partners:** Joint go-to-market with Kubernetes distributions (Red Hat OpenShift, Rancher)
- **MSP/MSSP partners:** Managed service providers offering AIOps as a service

### 3.3 Marketing Strategy

- **Content marketing:** SRE best practices, incident management guides, AIOps benchmarks
- **Developer relations:** Open-source contributions, conference talks (KubeCon, SREcon), blog posts
- **Community:** Slack community for SRE practitioners, monthly AMA with team
- **Demand generation:** Webinars, case studies, ROI calculator tool (public), benchmark reports

---

## 4. Expansion Strategy

### 4.1 Land and Expand

```
Month 1: Free tier (10 resources, basic anomaly detection)
  -> Customer sees value: 85% noise reduction on 10 services

Month 2-3: Starter tier ($2/resource, 50 resources = $100/mo)
  -> Expands to full production environment

Month 4-6: Professional tier upgrade ($5/resource, 200 resources = $1,000/mo)
  -> Wants correlation engine and SLO management

Month 9-12: Enterprise tier ($8/resource, 500 resources = $4,000/mo)
  -> Adopts autonomous remediation after building trust

Year 2: Full deployment (2,000 resources, $16,000/mo = $192K ARR)
  -> Deployed across all teams, all environments
```

### 4.2 Platform Extension Revenue (Future)

| Extension | Timeline | Revenue Model |
|---|---|---|
| AIOps Marketplace (community runbooks) | Year 2 | Revenue share on paid runbooks |
| Custom ML model training | Year 2 | Premium add-on ($1/resource/mo) |
| Compliance reporting (SOC2, DORA) | Year 3 | Add-on module ($2K/mo) |
| Multi-cloud cost optimization | Year 3 | % of cost savings (gain-share) |

---

## 5. Revenue Projections

### 5.1 Five-Year Revenue Model

| Metric | Year 1 | Year 2 | Year 3 | Year 4 | Year 5 |
|---|---|---|---|---|---|
| New Customers | 25 | 65 | 145 | 210 | 300 |
| Churned Customers | 0 | 4 | 10 | 16 | 20 |
| Total Customers | 25 | 86 | 221 | 415 | 695 |
| Resources Monitored | 50K | 250K | 750K | 1.5M | 3.0M |
| Avg Revenue/Resource/Mo | $2.00 | $2.60 | $3.20 | $3.80 | $4.20 |
| Monthly Recurring Revenue | $100K | $540K | $2.0M | $4.75M | $10.5M |
| **Annual Recurring Revenue** | **$1.2M** | **$5.8M** | **$18M** | **$38M** | **$68M** |
| YoY Growth | - | 383% | 210% | 111% | 79% |
| Gross Margin | 72% | 78% | 82% | 84% | 85% |

### 5.2 Revenue Bridge (Year 1 to Year 2)

```
Year 1 Ending ARR: $1.2M
  + New customer ARR: $3.1M (65 new customers)
  + Expansion ARR: $1.8M (existing customer growth + tier upgrades)
  - Churned ARR: -$0.3M (4 customers)
Year 2 Ending ARR: $5.8M
```

---

## 6. Key Assumptions

| Assumption | Value | Sensitivity |
|---|---|---|
| Average resources per new customer | 200 (Year 1) -> 400 (Year 5) | +/- 20% changes ARR by +/- 15% |
| Tier mix shift to Enterprise | 20% (Year 1) -> 50% (Year 5) | Critical driver of ARPU growth |
| Net Revenue Retention | 115% (Year 1) -> 145% (Year 5) | +/- 10% NRR = +/- $8M Year 5 ARR |
| Sales cycle (mid-market) | 45 days | Longer cycles delay revenue recognition |
| Infrastructure cost per resource | $0.30/mo at scale | GPU costs for ML training are the wildcard |

---

*This document is confidential and intended for potential investors only.*
