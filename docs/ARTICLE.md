# ⚔️ Talos: The Autonomous Cloud Guardian
>
> *The Comprehensive System Manifesto & Technical Reference*

## 1. Executive Summary

Talos is not just a monitoring tool; it is an **autonomous OODA-loop agent** designed to defend your cloud infrastructure against entropy, inefficiency, and waste. It operates on a "Guardian" philosophy: active defense rather than passive observation.

By combining a **4-Tier AI Swarm**, **Aggressive Idempotency**, and a **"Pulse of Vulcan" UI**, Talos delivers a self-healing, self-optimizing layer above AWS/GCP that feels alive.

---

## 2. Core Architecture: The OODA Loop

At the heart of Talos is the **Observe-Orient-Decide-Act** loop, executing every 10 seconds (configurable).

### Phase 1: OBSERVE 🔭

* **Deep Scanning**: Connects to cloud providers (AWS/GCP) to fetch live resource states.
* **Zombie Detection**: Identifies "Zombie Processes" (tasks interrupted by crashes) via the Ledger.
* **Metrics Collection**: Ingests CPU, Memory, and Network telemetry for utilization analysis.

### Phase 2: ORIENT 🧭

* **Contextualization**: Matches raw resources against business context (Tags: `env=dev`, `critical=true`).
* **Indie Force**: Checks for "Indie-Force" windows (12 AM - 6 AM) to enforce aggressive shutdown policies.
* **Knowledge Retrieval**: Queries `ProjectMemory` (Vector/JSON) for historical context on similar resources.

### Phase 3: DECIDE 🧠 (The AI Swarm)

Talos routes decisions through a tiered intelligence model based on **Risk Score (0.0 - 10.0)**.

| Tier | Name | Model | Risk Scope | Capability |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 1** | **Sentinel** | `Gemini 2.0 Flash` | 0.0 - 3.0 | Real-time pattern matching, log filtering, low-cost scanning. |
| **Tier 2** | **Strategist** | `Gemini Pro` | 3.0 - 7.0 | Cost-benefit analysis, rightsizing recommendations, architectural tweaks. |
| **Tier 3** | **Arbiter** | `Claude 3.5 Sonnet` | 7.0 - 9.0 | **Critical Safety Checks**, code audits, complex refactoring proposals. |
| **Tier 4** | **Oracle** | `Devin AI` | 9.0 - 10.0 | **Autonomous Engineering**. Full multi-file refactoring, migration planning. |
| **Tier 5** | **Reasoning** | `GPT-5 Mini` | Special | **Pure Reasoning**. Used for abstract architectural debates and "Why" analysis. |

### Phase 4: ACT ⚡

* **Idempotent Execution**: Every action is hashed (SHA-256) and recorded in the SQLite Ledger *before* execution.
* **Integrity Check**: Re-verifies checksums to prevent "Action Hallucination."
* **Self-Healing**: If the process crashes here, the next boot sees the `PENDING` state and resumes/rolls back.

---

## 3. Key Features & Capabilities

### 🛡️ Resilience & Safety

* **Local Ledger (SQLite)**: The source of truth. No dependency on external SaaS for state.
* **Crash Recovery**: Auto-resumes interrupted tasks on boot.
* **Personal Mode**: Strict approval gates for any action with Risk > 5.0 (configurable).
* **Adversarial Hardening**: System prompts are reinforced against "Ignore previous instructions" attacks.

### 💰 Cost Intelligence

* **Indie-Force Mode**: A specialized "Founder Mode" that ruthlessly shuts down non-critical dev resources at night.
* **ROI Tracking**: Real-time calculation of *Net Savings* vs. *AI Token Costs*.
* **Arbitrage Engine**: Can (theoretically) move workloads between Spot instances and On-Demand based on pricing.

### 🔮 The "Pulse of Vulcan" Dashboard

* **Glassmorphism**: Premium, translucent UI with frosted glass effects (3 tiers of blur).
* **OODA Visualizer**: Live animation showing the heartbeat of the autonomous loop.
* **Reactive Feedback**:
  * **Blue Glow**: Deep Sleep / Idle.
  * **Cyan Pulse**: Active Observation.
  * **Molten Red**: Critical Alert / High Risk Action.
  * **Golden Shimmer**: Oracle (Tier 4) Engagement.

### 🏗️ Infrastructure & Extensibility

* **Configurable**: `config.yaml` for tuning risk thresholds and API keys.
* **Dockerized**: `docker-compose up` ready for instant deployment.
* **Observability**: `/metrics` endpoint for Prometheus scraping.
* **CI/CD**: GitHub Actions pipeline for automated integrity verification.

---

## 4. Technical Stack

* **Language**: Go (Golang) 1.25+
* **Storage**: SQLite (Ledger), JSON (Logs/Memory)
* **Frontend**: Vanilla JS + CSS3 (Variables, Glassmorphism)
* **AI**: OpenRouter API (Gemini, Claude, GPT-5) + Devin API

---

## 5. Critical Review & Status

### ✅ Functionating Perfectly

* **Swarm Routing**: Correctly routes to Gemini Flash/Pro/Claude/Devin based on risk.
* **GPT-5 Integration**: Successfully added as a specialized reasoning tier.
* **OODA Loop**: The infinite blocking loop in `worker.go` is stable and handles panic recovery.
* **Dashboard**: Connectivity via SSE (Server-Sent Events) is robust.

### ⚠️ Enterprise Roadmap (In Progress)

* **Concurrency**: Migrating to Redis-backed worker swarm for horizontal scaling.
* **Database**: Transitioning from SQLite to PostgreSQL for multi-tenant state.
* **Auth**: Implementing OIDC/SAML for enterprise SSO integration.

---

> *"Talos does not just watch. It waits. It thinks. And when the time is right, it acts."*
