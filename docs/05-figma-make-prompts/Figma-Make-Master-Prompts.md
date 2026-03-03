# Figma & Make.com Master Prompts -- Sovereign AIOps Platform

**Module:** ERP-AIOps | **Port:** 5179 | **Version:** 2.0 | **Date:** 2026-03-03

---

## Inputs to Attach Before Prompting

1. **BRD.md** -- Business Requirements Document
2. **PRD.md** -- Product Requirements Document with all 38 functional requirements
3. **HLD.md** -- High-Level Design with component architecture
4. **Data-Architecture.md** -- Full ERD and data dictionary
5. **Brand Kit** -- Sovereign brand colors, typography, iconography

---

## Screen 1: Operations Command Center (Home Dashboard)

### Figma Prompt

```
Design a full-screen operations command center for an AIOps platform called "Sovereign AIOps."
This is the primary view that SREs and Incident Commanders use during their shift.

Layout (3-column):
LEFT COLUMN (25% width):
- Active Incidents panel: list of open incidents sorted by severity (fatal=red, critical=orange, warning=yellow)
- Each incident card shows: severity badge, title, affected service, MTTR timer (counting up), assigned engineer avatar, correlation badge (number of correlated events)
- Filter pills: All | Fatal | Critical | Warning | Acknowledged | Unacknowledged
- "New Incident" count badge at top

CENTER COLUMN (50% width):
- Service Map: interactive topology visualization showing all discovered services as nodes
- Node shapes: circle=microservice, cylinder=database, diamond=cache, hexagon=queue, pentagon=gateway
- Node colors: green=healthy, yellow=degraded, red=unhealthy, gray=unknown
- Edge lines: solid=sync, dashed=async; thickness=traffic volume; color=health
- Animated pulse on nodes with active anomalies
- Click node to see: service name, health score, active anomalies, SLO status, recent changes
- Blast radius overlay: when incident selected, affected services highlight with red halo
- Zoom/pan controls, fullscreen toggle

RIGHT COLUMN (25% width):
- Live Alert Stream: real-time feed of incoming events (after noise reduction)
- Each alert shows: timestamp, severity dot, service name, message (truncated), source icon
- Auto-scroll with pause-on-hover
- Noise reduction stats bar: "472 alerts reduced to 12 incidents today (97.5% noise reduction)"

TOP BAR:
- Sovereign AIOps logo
- Environment selector: Production | Staging | Development
- Time range: Last 1h | 6h | 24h | 7d | Custom
- Global search (cmd+k)
- Quick stats: MTTR Today: 4.2 min | Open Incidents: 3 | SLO Compliance: 99.7% | Noise Reduced: 97.5%

BOTTOM BAR:
- On-call banner: "On-call: Jane Smith (Tier 1) | Next handoff: 6h 23m"
- System health: "Ingestion: 42K events/sec | Detection Latency: 1.2s | Correlation: OK"

Color scheme: Dark theme (charcoal background #1a1a2e, card backgrounds #16213e, accent blue #0f3460, alert red #e94560, success green #0db39e)
Typography: Inter for body, JetBrains Mono for metrics/timestamps
```

---

## Screen 2: Anomaly Investigation Workspace

### Figma Prompt

```
Design an anomaly investigation workspace for SREs to deep-dive into a detected anomaly.

HEADER:
- Breadcrumb: Command Center > payment-api > Anomaly #A-2847
- Anomaly title: "Latency anomaly detected on payment-api"
- Confidence score badge: "0.92 confidence" with color (green >0.8, yellow 0.5-0.8, red <0.5)
- Detection model: "Isolation Forest + LSTM" tag
- Time detected: "2 minutes ago" with absolute timestamp on hover
- Status: Investigating | Acknowledged | Resolved | False Positive (dropdown)

LEFT PANEL (60%):
- Time series chart (large, interactive):
  - Actual metric line (blue)
  - Predicted baseline (gray dashed)
  - Anomaly region shaded in red
  - Deployment markers (vertical green lines with deploy info on hover)
  - Maintenance windows (shaded gray background)
  - Zoom/brush selection for time range
  - Toggleable overlays: predicted, threshold, baseline

- Below chart: Feature Attribution panel
  - Horizontal bar chart of SHAP values
  - "z_score: +0.42, rate_of_change: +0.31, deployment_window: +0.18"
  - Label: "Top contributing features to this anomaly classification"

RIGHT PANEL (40%):
- Anomaly Details card:
  - Metric: latency_p99_ms
  - Value: 1,250ms (baseline: 180ms, +594%)
  - Service: payment-api
  - Environment: production
  - Pod: payment-api-7b9c4-x8k2j

- Correlated Events card:
  - List of events that arrived in the same time window
  - Each with: source icon, severity, service, message, timestamp
  - "3 events correlated via topological analysis"

- Similar Past Anomalies card:
  - List of historically similar anomalies with resolution
  - "2026-02-28: Similar latency spike, resolved by scaling replicas (auto)"
  - "2026-02-15: Similar pattern, root cause was database connection pool"

- Suggested Actions card:
  - "Run runbook: restart-unhealthy-pods (85% match)" with Execute button
  - "Scale payment-api replicas (72% match)" with Execute button
  - "Investigate upstream: db-primary (root cause probability: 0.78)"

Dark theme, consistent with command center.
```

---

## Screen 3: Runbook Executor with Live Step Tracking

### Figma Prompt

```
Design a runbook execution interface showing live step-by-step progress.

HEADER:
- Runbook name: "Restart Unhealthy Pods"
- Trigger: "Incident #INC-1847 - payment-api CrashLoopBackOff"
- Mode badge: "Suggest & Approve" (blue) or "Autonomous" (green) or "Observe Only" (gray)
- Overall status: large progress indicator

LEFT PANEL (40%):
- Step Pipeline (vertical stepper):
  - Step 1: "Verify Issue" - checkmark (completed, 2.1s)
  - Step 2: "Count Unhealthy" - checkmark (completed, 0.8s)
  - Step 3: "Check Guardrails" - checkmark (completed, 0.3s)
  - Step 4: "Delete Unhealthy Pods" - spinner (running, 12s elapsed)
  - Step 5: "Wait for Healthy" - gray (pending)
  - Step 6: "Verify Health" - gray (pending)
  - Each step shows: name, status icon, duration, expand arrow
  - Expanded step shows: action type, params (formatted YAML), output

- Guardrails Status card:
  - Max affected pods: 2/3 (green checkmark)
  - Max affected %: 16%/25% (green checkmark)
  - Healthy replicas: 4/2 minimum (green checkmark)
  - Cooldown: No recent execution (green checkmark)
  - Blackout window: Not in blackout (green checkmark)

RIGHT PANEL (60%):
- Live Output terminal (dark, monospace):
  - Scrolling output from current step
  - Syntax highlighted (JSON outputs, timestamps in gray, errors in red)
  - "[10:15:32] Identifying unhealthy pods in namespace production..."
  - "[10:15:33] Found 2 pods in CrashLoopBackOff: payment-api-7b9c4, payment-api-8d5e1"
  - "[10:15:34] Guardrails check: PASSED (2 pods <= 3 max, 16% <= 25% max)"
  - "[10:15:35] Deleting pod payment-api-7b9c4..."

- Below terminal:
  - Action buttons: "Abort Execution" (red), "Pause" (yellow), "Approve Next Step" (blue, if in approve mode)
  - Rollback info: "Rollback available: Scale deployment to 12 replicas"

- Audit trail card:
  - Compact timeline of all logged actions with timestamps
  - "Immutable audit log - 6 entries"
```

---

## Screen 4: SLO Dashboard

### Figma Prompt

```
Design an SLO management dashboard showing error budget burn rates and compliance.

TOP ROW - Summary Cards (4 across):
- "SLO Compliance": 99.7% with trend arrow (up, green) and sparkline
- "Budget Remaining (avg)": 68.3% with donut chart
- "Services At Risk": 2 with red background
- "Burn Rate Alerts": 1 active PAGE, 2 active TICKETs

MAIN CONTENT:
Table/Grid of SLOs (one row per SLO definition):
Columns:
- Service name + SLO name
- SLI Type icon (availability, latency, error rate)
- Target (99.9%, 99.95%, etc.)
- Current SLI value (green if meeting, red if not)
- Error Budget bar: horizontal progress bar showing consumed vs remaining
  - Green zone: <50% consumed
  - Yellow zone: 50-80% consumed
  - Red zone: >80% consumed
  - Tooltip: "12.4 minutes remaining of 43.2 minute budget"
- Burn Rate: current multiplier with directional arrow
  - "2.1x" (green), "6.8x" (yellow, TICKET), "15.2x" (red, PAGE)
- Trend sparkline (7-day SLI values)
- Status: "Healthy" | "At Risk" | "Breaching" | "Exhausted"

DETAIL PANEL (when row clicked, slides in from right):
- Large burn rate chart: multi-window view
  - 1-hour window line
  - 6-hour window line
  - 3-day window line
  - Threshold lines: 14.4x (PAGE), 6x (TICKET), 1x (REVIEW)
- Error budget consumption timeline (30-day view)
  - Area chart: budget consumed over time
  - Incident markers: dots on the chart where budget burned fastest
  - Projected exhaustion date line
- SLI measurement chart: raw SLI values over time
- Recent incidents affecting this SLO
- "Edit SLO" and "Mute Alerts" buttons

Dark theme. Error budget visualizations use gradient from green to red.
```

---

## Screen 5: Topology Visualizer

### Figma Prompt

```
Design an interactive service topology graph visualization.

MAIN AREA (full screen, interactive canvas):
- Force-directed graph layout with physics-based positioning
- Nodes represent services:
  - Size proportional to traffic volume
  - Shape by type: circle=microservice, cylinder=database, diamond=cache, hexagon=queue
  - Color by health: green=healthy (pulsing gently), yellow=degraded, red=unhealthy (pulsing urgently)
  - Label below each node: service name
  - Badge on node: anomaly count (if any)
- Edges represent dependencies:
  - Solid lines for synchronous (HTTP/gRPC)
  - Dashed lines for asynchronous (Kafka/SQS)
  - Arrow direction shows call direction
  - Edge thickness proportional to requests/second
  - Edge color: green=healthy, yellow=elevated latency, red=high error rate
  - Animated dots flowing along edges showing traffic direction

TOOLBAR (top):
- Filter by: environment, namespace, service type, health status
- Layout: Force | Hierarchical | Circular | Grid
- Overlay: None | Latency Heatmap | Error Rate | Traffic Volume
- Search: find and highlight specific service
- Group by: namespace | team | tier
- "Show blast radius for:" dropdown (select incident)

HOVER PANEL (floating, appears on node hover):
- Service: payment-api
- Type: Microservice | Tier: Critical
- Health: Healthy (99.8% availability)
- Active Anomalies: 0
- SLO Status: 99.92% (target: 99.9%) - green
- Latency p99: 45ms
- Error rate: 0.02%
- Last deployment: 3 hours ago
- Dependencies: 4 upstream, 2 downstream

SIDE PANEL (collapsible, on node click):
- Detailed service card with metrics over time
- Dependency list with health indicators
- Recent incidents for this service
- Active runbook executions
- Quick actions: "View anomalies", "View SLO", "Open in Datadog"

Dark theme with high-contrast nodes against dark background.
```

---

## Screen 6: Capacity Planning Forecast Charts

### Figma Prompt

```
Design a capacity planning dashboard with ML-powered forecast charts.

TOP ROW - Alerts:
- Capacity warning cards: "payment-api CPU predicted to exceed 85% in 18 hours"
- Each card: resource icon, service name, predicted breach time, confidence badge, "View Details" link
- Color: yellow for 24-72h warnings, red for <24h warnings

MAIN GRID (2x2 chart layout, selectable per service):
Each chart:
- Title: "{Service} - {Resource} Utilization"
- X-axis: time (72-hour view default, toggleable: 72h | 7d | 30d | 90d)
- Y-axis: utilization percentage (0-100%)
- Lines:
  - Actual (solid blue): historical utilization
  - Predicted (dashed blue): ML forecast
  - Confidence interval (shaded light blue): upper/lower bounds
  - Threshold line (horizontal red dashed): 85% capacity limit
- Markers:
  - Deployment events (vertical green lines)
  - Scaling events (vertical purple lines)
  - Predicted breach point (red diamond with timestamp)
- Legend below chart

RECOMMENDATIONS PANEL (below charts):
- Table of scaling recommendations:
  - Service | Resource | Current | Recommended | Savings/Risk | Action
  - "payment-api | CPU | 4 replicas | 6 replicas | Prevent breach in 18h | Apply"
  - "user-service | Memory | r5.xlarge | r5.large | Save $120/mo | Review"
  - "cache-service | Connections | 1000 max | 500 max | Overprovisioned | Apply"

MODEL ACCURACY SIDEBAR (collapsible):
- Last 7 predictions vs. actuals
- MAE and MAPE scores
- Model version and last training date
- "Retrain Model" button (admin only)
```

---

## Screen 7: Incident War Room

### Figma Prompt

```
Design an incident war room view for active major incident coordination.

HEADER (red banner for severity):
- "INCIDENT #INC-1847 - CRITICAL" with pulsing dot
- Title: "Payment processing degradation - multiple services affected"
- Duration timer: "Active for 23 minutes"
- Assigned IC: "Jane Smith" with avatar
- Action buttons: "Acknowledge" | "Escalate" | "Resolve" | "Add Update"

LEFT PANEL (50%):
- Incident Timeline (vertical, real-time updating):
  - Each entry: timestamp, type icon (anomaly/alert/action/communication/deploy), description, actor
  - "10:02:14 - Anomaly detected on payment-api (latency spike, 0.92 confidence)"
  - "10:02:16 - Correlated with 4 downstream service alerts"
  - "10:02:18 - Incident created, assigned to Jane Smith (on-call Tier 1)"
  - "10:02:45 - Runbook 'restart-unhealthy-pods' suggested (82% match)"
  - "10:03:12 - Jane acknowledged incident"
  - "10:05:00 - Runbook execution started (approved by Jane)"
  - "10:07:30 - Health check failed, rollback initiated"
  - "10:08:15 - Escalated to Tier 2 (Mike Johnson)"
  - Entries auto-scroll, with anchor option

RIGHT PANEL (50%):
- Blast Radius Map (mini topology): affected services highlighted, unaffected grayed
- Correlated Events table: all events grouped into this incident
- Active Remediation tracker (if running): mini step progress
- Communication panel: integration with Slack thread, ability to post updates
- Responders list: who is on this incident, their role, join time

BOTTOM:
- Quick update input: "Type incident update..." with severity selector
- Keyboard shortcuts: Ctrl+A=Acknowledge, Ctrl+E=Escalate, Ctrl+R=Resolve
```

---

## Screen 8: Post-Mortem Builder

### Figma Prompt

```
Design a post-mortem document builder that auto-populates from incident data.

HEADER:
- Title: "Post-Mortem: Payment Processing Degradation - INC-1847"
- Status: Draft | In Review | Published (workflow steps)
- Author: auto-filled from IC
- Reviewers: multi-select user picker
- Incident link: clickable reference to original incident

MAIN CONTENT (rich text editor, left 65%):
Sections (auto-populated, editable):

1. Executive Summary (auto-generated, editable)
   - "On March 3, 2026, payment-api experienced a 23-minute degradation..."
   - Severity, duration, customer impact auto-calculated

2. Timeline (auto-generated from incident timeline)
   - Formatted timeline with timestamps and events
   - "Edit" button to add/remove entries
   - Drag-and-drop reordering

3. Root Cause Analysis (partially auto-generated)
   - Root cause service identified by correlation engine
   - Contributing factors listed from anomaly data
   - Editable free-text for human analysis

4. Impact Assessment (auto-calculated)
   - Users affected: estimated from traffic data
   - Revenue impact: calculated from SLO burn
   - SLO budget consumed: shown with before/after

5. Action Items (editable table)
   - Auto-suggested from incident data
   - Columns: Action | Priority | Owner | Due Date | Status
   - "Add Action Item" button
   - Each action links to task tracker

6. Lessons Learned (free text)

SIDEBAR (right 35%):
- Incident Reference: link to incident, timeline, topology
- Attached Evidence: screenshots, logs, charts (drag to add)
- Review Status: who has reviewed, comments
- Publishing controls: "Submit for Review" | "Publish" | "Export PDF"
- Template selector: choose post-mortem format

Dark theme with clean document editing experience. Auto-saved indicator.
```

---

## Make.com Automation Prompts

### Automation 1: Incident-to-Slack Notification Pipeline
```
Create a Make.com scenario that:
1. Trigger: Webhook receives new incident from Sovereign AIOps API
2. Filter: Only process incidents with severity "critical" or "fatal"
3. Slack: Post to #incidents channel with formatted message:
   - Incident ID, severity emoji, title, affected services, MTTR timer link
   - Thread: add correlated events as thread replies
4. PagerDuty: Create PagerDuty incident and link back
5. Google Sheets: Log incident metadata for monthly reporting
```

### Automation 2: SLO Budget Alert Pipeline
```
Create a Make.com scenario that:
1. Trigger: Webhook receives SLO burn rate alert
2. If burn rate > 14.4x: Send Slack DM to on-call + PagerDuty page
3. If burn rate > 6x: Create Jira ticket for SRE team
4. If burn rate > 1x: Add to weekly SLO review Google Doc
5. Update SLO tracking spreadsheet with current values
```

### Automation 3: Post-Mortem Distribution
```
Create a Make.com scenario that:
1. Trigger: Post-mortem status changes to "Published"
2. Generate PDF from post-mortem content
3. Email to stakeholder distribution list
4. Post summary to #engineering Slack channel
5. Create follow-up Jira tickets for each action item
6. Update incident metrics dashboard
```

---

*Document Control: UI prompts should be iterated with design team feedback. Each screen requires usability testing with target persona.*
