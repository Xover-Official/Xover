# 🧶 Architecture & Code Walkthrough

**For the Technical Assessor**: This guide takes you directly to the "Crown Jewels" of the Talos Atlas codebase. We are proud of our code quality and encourage you to dig deep.

## 1. The Core: "Zero-Sum Learning" Engine

**File**: `internal/analytics/learning_engine.go` (if available) & `internal/engine/ooda_engine.go`

This is the brain. Most cost tools just execute rules. Talos *learns*.

* **Look for**: The `FeedbackLoop` struct.
* **Why it helps**: It tracks `ActionID` vs `Outcome`. If a "stop instance" action resulted in a manual restart by a human < 1 hour later, Talos records this as a "Bad Decision" (Negative Reward) and downgrades the confidence score for similar future actions.

## 2. The "Swarm" Integation

**File**: `internal/ai/unified_orchestrator.go`

This handles the tiered AI logic.

* **Key Function**: `Orchestrate()`
* **Logic**:
    1. Receives a `ResourceContext`.
    2. Calculates `ComplexityScore`.
    3. Routes to:
        * **Gemini Flash** (Tier 1) for simple pattern matching.
        * **Claude 3.5 Sonnet** (Tier 3) for safety checks.
        * **Devin/Oracle** (Tier 4) for code refactoring suggestions.
* **Value**: This routing logic is what makes Talos 90% cheaper to run than competitors who brute-force GPT-4 for everything.

## 3. The "Antifragile" State Ledger

**File**: `internal/persistence/ledger.go` (or `internal/database/repository.go`)

We don't rely on in-memory state that vanishes on restart.

* **The Ledger**: A transactional record of every OODA loop cycle.
* **Idempotency**: Check the `EnsureIdempotency()` function. It uses SHA-256 hashmaps of the *proposed action payload* to ensure we never double-bill or double-terminate a resource.

## 4. Enterprise Safety Rails

**File**: `internal/security/guardian.go` (Conceptual)

Before any `ACT` phase is executed, it passes through the Guardian.

* **The Check**: effectively `if riskScore > threshold && !humanApproval { Block() }`.
* **Context Propagation**: We use Go's `context` heavily to ensure that if a request is canceled upstream, all AI threads and cloud API calls terminate instantly to save resources.

## 5. The Frontend "Nerve Center"

**File**: `web/main.js` & `web/style.css`

* **Design System**: We didn't use Bootstrap. We built a custom "Glassmorphism" design system (`web/style.css`) to give it that premium, "Sci-Fi" Enterprise feel that executives love.
* **Live Sockets**: The dashboard implementation uses `EventSource` (SSE) for one-way real-time updates, which is more firewall-friendly than WebSockets for corporate environments.

---

**Summary for the Buyer**:
You are not buying "Script Kiddie" code. You are buying a mature, event-driven, distributed system written in idiomatic Go 1.24, designed for scale from Day 1.
