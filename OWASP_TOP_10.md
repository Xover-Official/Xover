# 🛡️ OWASP Top 10 Compliance Statement

TALOS is built with enterprise security as a first-class citizen. This document outlines how we address the OWASP Top 10 (2021) risks.

## A01:2021-Broken Access Control

- **RBAC Implemented**: Role-Based Access Control is enforced at the middleware level in the dashboard.
- **JWT Protection**: All API endpoints require a valid JWT with appropriate claims.
- **Principle of Least Privilege**: Workers only have the permissions necessary to modify resources tagged for optimization.

## A02:2021-Cryptographic Failures

- **TLS Enforcement**: Recommended for all production deployments (AWS ALB/GCP Ingress).
- **Secure Secret Storage**: Credentials are never stored in the database; they are managed via environment variables and cloud-native secret managers.

## A03:2021-Injection

- **SQL Injection**: We use `pgx/v5` with parameterized queries for all database operations. Concatenation of user input into queries is strictly forbidden and audited.
- **Prompt Injection**: LLM contexts are hardened with system instructions to ignore adversarial prompts.

## A04:2021-Insecure Design

- **Autonomous Guardrails**: The OODA loop includes a "Risk Scoring" phase that prevents high-risk actions from executing without human approval.
- **Idempotency**: Every action has a unique checksum to prevent unintended reprocessing.

## A05:2021-Security Misconfiguration

- **Docker-First Policy**: We provide hardened Production Dockerfiles that run with non-root users.
- **Minimal Surface**: Only necessary ports (8080/8081) are exposed.

## A06:2021-Vulnerable and Outdated Components

- **Dependency Audit**: Regular `go mod tidy` and scanning with `gosec` and `trivy`.
- **Minimal Core**: We minimize the use of third-party libraries in the critical request path.

## A07:2021-Identification and Authentication Failures

- **SSO Integration**: Out-of-the-box support for Enterprise SSO (Google/Okta/Azure) to leverage corporate security policies.
- **Session Management**: JWTs have strict expiration (24h default) and are signed with HS256/RS256.

## A08:2021-Software and Data Integrity Failures

- **Checksum Integrity**: Optimization actions are validated against a SHA256 checksum of the resource state before execution.
- **CI/CD Integration**: Pipelines verify the integrity of builds before deployment.

## A09:2021-Security Logging and Monitoring Failures

- **Zap Logging**: All security-relevant events (Login, Policy Failure, AI Action) are logged via structured JSON (Zap).
- **OpenTelemetry**: Distributed tracing allows for auditing the full lifecycle of a request.

## A10:2021-Server-Side Request Forgery (SSRF)

- **Bounded Proxies**: AI clients are configured to only talk to verified endpoints (OpenRouter/Anthropic/Google).
- **Network Isolation**: Recommended deployment in private subnets with NAT gateways.

---

**Status**: ✅ COMPLIANT (Audit Date: Feb 2026)
