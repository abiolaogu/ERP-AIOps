# ADR-0002: Why We Chose Vitastor Over Ceph

- **Status**: Accepted
- **Date**: 2026-03-02
- **Deciders**: Platform Lead, DevOps Lead, Security Architect, Engineering Lead
- **Technical Story**: Storage architecture evaluation for self-managed cloud-native environments

## Context

The platform requires distributed block/object storage support for stateful workloads in virtualized and Kubernetes-managed Environments. The team evaluated Ceph and Vitastor for performance, operational complexity, integration effort, and hardware efficiency in a constrained private-cloud footprint.

## Decision

We selected **Vitastor** as the preferred storage backend for target private-cloud workloads where low-latency performance and operational simplicity are prioritized. Ceph remains a known fallback for environments that require broad ecosystem compatibility and mature managed-service patterns.

## Decision Drivers

- Predictable low-latency behavior in NVMe-heavy clusters
- Lower operational overhead for small-to-medium team operations
- Simpler tuning profile for targeted workload classes
- Better fit for near-term platform team capacity

## Alternatives Considered

1. **Ceph as primary**
   - Pros: mature ecosystem, broad community support, rich feature set
   - Cons: higher operational complexity and tuning overhead
2. **Dual-stack Vitastor + Ceph**
   - Pros: flexibility for workload classes
   - Cons: duplicated runbooks, split expertise, higher support burden
3. **Managed cloud storage only**
   - Pros: low operations burden
   - Cons: does not satisfy private-cloud and data residency constraints

## Consequences

### Positive

- Faster operational onboarding for platform engineers
- Lower Day-2 complexity for maintenance and troubleshooting
- Strong baseline performance for targeted stateful workloads

### Negative

- Smaller ecosystem compared to Ceph
- Requires explicit knowledge transfer and internal standards

### Neutral

- Keep interoperability path documented if migration to Ceph is needed later.

## Implementation Notes

- Maintain storage abstraction in deployment manifests.
- Keep workload-level storage class mapping documented.
- Review storage decision quarterly against SLO and incident trends.

## References

- [CI/CD and deployment architecture](../../03-quality-devops/CI-CD.md)
- [Disaster recovery plan](../../03-quality-devops/DR-Plan.md)
- [Operations manual](../../04-user-ops/Operations-Manual.md)

