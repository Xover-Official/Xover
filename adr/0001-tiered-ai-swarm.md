# ADR 0001: Tiered AI Swarm Orchestration

## Status

Accepted

## Context

TALOS needs to perform complex cloud infrastructure optimizations. Traditional rules-based engines are often too rigid, while using high-end LLMs (like GPT-4o or Claude 3.5 Sonnet) for every resource scan is prohibitively expensive and slow.

## Decision

We will implement a **Tiered AI Swarm** architecture. This approach categorizes infrastructure decisions into levels of risk and complexity, routing them to the most cost-effective model capable of handling the task.

### The Tiers

1. **Sentinel (Gemini 1.5 Flash)**: Used for 24/7 observation and pattern recognition. Low cost, high speed.
2. **Strategist (Gemini 1.5 Pro)**: Used for medium-complexity rightsizing and scheduling plans.
3. **Arbiter (Claude 3.5 Sonnet)**: Used for high-risk decisions (e.g., terminating production databases) where safety and reasoning are paramount.
4. **Oracle (Devin/GPT-4o)**: Reserved for architectural crossroads and extreme complexity cases.

## Consequences

- **Positive**: Significantly reduced operating costs (up to 95% compared to using GPT-4o for everything).
- **Positive**: Improved system safety by using superior reasoning models for critical actions.
- **Negative**: Increased complexity in the `UnifiedOrchestrator` to manage model routing and fallback logic.
- **Negative**: Dependency on multiple AI providers.
