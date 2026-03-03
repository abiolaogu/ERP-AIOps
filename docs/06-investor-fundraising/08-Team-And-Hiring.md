# Team & Hiring Plan -- Sovereign AIOps

**Confidential -- Series A Investment Memorandum**

---

## 1. Founding Team

### Leadership

| Role | Background | Why This Person |
|---|---|---|
| **CEO/Co-Founder** | 15 years infrastructure, ex-VP Engineering at scale-up (50M+ users), built platform serving 100M+ requests/day | Understands enterprise buyer, has scaled engineering teams from 10 to 200 |
| **CTO/Co-Founder** | ML/AI specialist, ex-Staff Engineer at FAANG, published research in anomaly detection for distributed systems, 3 patents | Deep technical expertise in the exact ML domain we need |
| **VP Engineering** | 12 years SRE leadership, ex-Google SRE, co-authored SRE best practices, managed 500+ service fleet | Lived the pain -- built AIOps internally, now productizing |
| **Head of ML** | PhD in time-series analysis, ex-research scientist at leading AI lab, contributed to open-source ML libraries | Academic rigor + production ML experience, rare combination |

### Current Team (10 people)

| Department | Count | Key Hires |
|---|---|---|
| Engineering (Backend/Platform) | 5 | Go microservices, Kafka, PostgreSQL, Kubernetes operators |
| ML Engineering | 2 | Anomaly detection, model serving, feature engineering |
| Product | 1 | Ex-PagerDuty PM, deep domain expertise |
| Design | 1 | Ex-Datadog, operations dashboard specialist |
| Operations/Admin | 1 | Finance, HR, office management |

---

## 2. Hiring Plan (Post-Series A)

### 2.1 Year 1 Hires (10 -> 27)

| Role | Count | Priority | Comp Range (Total) |
|---|---|---|---|
| Senior Backend Engineer (Go) | 3 | Q1 | $200-250K |
| ML Engineer | 2 | Q1 | $220-280K |
| Frontend Engineer (React) | 2 | Q1-Q2 | $180-220K |
| Platform/Infra Engineer | 2 | Q2 | $200-250K |
| Account Executive (Mid-Market) | 2 | Q1 | $150-200K OTE |
| Account Executive (Enterprise) | 1 | Q2 | $250-350K OTE |
| SDR | 2 | Q1 | $80-100K OTE |
| Solutions Engineer | 1 | Q2 | $180-220K |
| Developer Relations | 1 | Q2 | $170-200K |
| Content Marketing | 1 | Q2 | $130-160K |

### 2.2 Year 2 Hires (27 -> 47)

| Department | Additional Hires | Focus |
|---|---|---|
| Engineering | 6 | Remediation engine, chaos engineering, scale |
| ML/AI | 2 | Capacity forecasting, NLP for runbooks |
| Sales | 5 | Enterprise expansion, international |
| Marketing | 2 | Demand gen, product marketing |
| Customer Success | 2 | Onboarding, health management |
| G&A | 1 | People operations |

---

## 3. Engineering Culture

### 3.1 Technical Values

- **Dogfooding:** We run Sovereign AIOps on our own infrastructure. Every engineer does on-call.
- **Ownership:** Service owners are responsible end-to-end (build, deploy, operate, support)
- **Data-driven:** Every feature has measurable impact metrics defined before development
- **Open source first:** Contribute to upstream projects, open-source non-core components
- **Security by design:** Threat modeling in design phase, not as an afterthought

### 3.2 Key Technical Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Primary language | Go | Performance, concurrency model, Kubernetes ecosystem |
| ML framework | ONNX Runtime (Go) + Python (training) | Production inference in Go, training flexibility in Python |
| Database | PostgreSQL + TimescaleDB + pgvector | Single database engine, extensions for time-series and vectors |
| Streaming | Apache Kafka | Industry standard, proven at scale, rich ecosystem |
| Frontend | React + Vite + Refine.dev + Ant Design | Rapid development, enterprise UI components |
| API | Hasura GraphQL | Auto-generated from schema, real-time subscriptions |

---

## 4. Advisory Board

| Advisor | Background | Contribution |
|---|---|---|
| **[SRE Leader]** | Ex-VP SRE at major tech company, authored industry SRE book | Product direction, customer introductions |
| **[Enterprise Sales]** | Ex-CRO at $500M ARR dev tools company | GTM strategy, enterprise sales playbook |
| **[ML Research]** | Professor at top CS department, anomaly detection research | ML model architecture, research partnership |
| **[Security/Compliance]** | Ex-CISO at Fortune 500 | SOC2/FedRAMP guidance, enterprise security requirements |

---

## 5. Diversity & Culture Commitments

- Target 40% underrepresented groups in engineering by Year 2
- Fully remote-first with quarterly team gatherings
- Transparent compensation bands published internally
- Unlimited PTO with minimum 3-week mandatory usage
- Mental health support (on-call burnout prevention is our mission internally too)

---

*This document is confidential and intended for potential investors only.*
