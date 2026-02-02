# 🛡️ Security Policy

## Reporting a Vulnerability

We take the security of TALOS seriously. If you discover a security vulnerability within this project, please follow the steps below to report it responsibly.

### 🛡️ How to Manage Secrets

TALOS is designed for enterprise environments. **Never** commit secrets, API keys, or credentials to the source code.

1. Use the provided `.env.template` to create your own `.env` file.
2. Use environment variables in production (K8s Secrets, AWS Secrets Manager, HashiCorp Vault).
3. All sensitive variables are postfixed with `_KEY`, `_SECRET`, or `_PASSWORD` in the configuration files.

### 📞 Security Contact Info

Please report security vulnerabilities via email to: **<security@talos.io>**

Include the following information in your report:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting)
- Location of the issue (e.g., file name, function, API endpoint)
- Potential impact of the issue
- Steps to reproduce the issue
- Any suggested fix or mitigation

### 📜 Disclosure Policy

- We will acknowledge receipt of your report within 48 hours.
- We will provide an estimated timeframe for a fix.
- We will notify you once the vulnerability has been patched.
- We ask you not to disclose the issue publicly until we have had a chance to fix it and release a patch.

## 🔒 Security Features

- **JWT Authentication**: Secure, token-based authentication for all dashboard and API interactions.
- **RBAC (Role-Based Access Control)**: Granular permissions for Viewers, Editors, and Admins.
- **SSO Integration**: Out-of-the-box support for Google Workspace, Okta, and Azure AD.
- **Audit Logs**: Every action taken by the AI or a user is recorded in an immutable ledger.
- **Adversarial Guardrails**: AI prompts are hardened against injection attacks.

## ⚖️ License & Compliance

TALOS is licensed under the MIT License. It is designed to be compliant with standard cloud security practices and can be audited using standard Go security tools.
