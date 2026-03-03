# Figma Make Prompts: Design System

## Objective

Generate a scalable design system for Product web and Admin experiences with accessibility and implementation parity.

## Prompt 1: Token Foundation

```text
Create design tokens with semantic naming:
- color: bg, surface, border, text, accent, success, warning, error, info
- spacing: 4, 8, 12, 16, 20, 24, 32, 40, 48
- typography: display, heading, body, caption, mono
- radius: sm, md, lg, xl, full
- shadow and elevation levels
Include dark-mode readiness but prioritize light enterprise baseline.
```

## Prompt 2: Core Components

```text
Build components with variants and state coverage:
- Button, Input, Select, Checkbox, Radio, Switch
- Data Table, Pagination, Tabs, Modal, Drawer, Toast
- Badge, Tag, Tooltip, Breadcrumb, Avatar, Empty State
For each component include states: default, hover, focus, disabled, loading, error.
```

## Prompt 3: Composite Patterns

```text
Create composite patterns:
- dashboard card group
- filter/search toolbar
- table with bulk actions
- inline approval workflow panel
- audit timeline viewer
Map each pattern to expected data density and responsive behavior.
```

## Prompt 4: Responsive and Mobile

```text
Produce responsive variants for desktop (1440), tablet (1024), mobile (390).
Define which interactions collapse into drawer/sheet on mobile.
Include touch target sizing and spacing adjustments.
```

## Prompt 5: Handoff Contract

```text
Generate handoff metadata:
- token export map for CSS variables
- component props matrix
- interaction state tokens
- accessibility constraints
- implementation notes for engineering
```

## Accessibility and Inclusivity Add-On

```text
Audit all components for WCAG AA contrast, keyboard focus visibility, and non-color dependent status indicators.
```

