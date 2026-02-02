# 🎥 TALOS: The 5-Minute Enterprise Demo

Since the autonomous browser agent is currently limited by the execution environment, use this script to perform the live demo yourself. The system has been pre-configured with a "Simulation Mode" that auto-generates enterprise-scale data for presentation purposes.

## 🚀 Setup (1 Minute)

1. **Start the Dashboard**:
    Open your terminal and run the following command to start the dashboard with all modules active:

    ```resh
    go run cmd/dashboard/main.go cmd/dashboard/cache.go cmd/dashboard/auth_handlers.go cmd/dashboard/token_handlers.go
    ```

    *Wait for the log message: `Starting server on :8080...`*

2. **Access the Console**:
    Open your web browser to: [http://localhost:8080](http://localhost:8080)
    *Note: If redirected to login, use any credentials (simulation mode accepts all).*

---

## 🎬 The Script (4 Minutes)

### **Minute 1: The "Wow" Factor (Visuals)**

- **Action**: Land on the dashboard and pause.
- **Narrative**: "Welcome to Talos. What you're seeing isn't just a dashboard—it's the brain of an autonomous cloud guardian. Notice the 'System Status' pulse in the top right? That's the heartbeat of the OODA loop running in real-time."
- **Focus**:
  - **Realized Savings Card**: Watch the ticker count up (simulated live savings).
  - **Health Circle**: Point out the 98% health score, indicating stability despite aggressive optimization.

### **Minute 2: The AI Swarm (Core Tech)**

- **Action**: Hover over the "Collective Swarm Intelligence" grid nodes.
- **Narrative**: "This is where the magic happens. Talos uses a multi-tiered AI swarm.
  - **Sentinel (Green)**: That's Gemini Flash, scanning thousands of resources for pennies.
  - **Strategist (Blue)**: Gemini Pro, planning deeper optimizations.
  - **Arbiter (Violet)**: That's Claude 3.5 Sonnet. See it verify high-risk decisions? It only activates when safety is paramount."

### **Minute 3: The OODA Loop (Logic)**

- **Action**: Watch the "Current OODA Operation" text change.
- **Narrative**: "Talos doesn't just read logs. It Observes, Orients, Decides, and Acts. Right now, it's in the 'Decide' phase, weighing the risk of rightsizing that RDS instance versus the cost savings. It's using T.O.P.A.Z. logic to ensuring zero downtime."

### **Minute 4: The Impact (ROI)**

- **Action**: Click the "View Full Logs" button (or scroll to the Task Pipeline).
- **Narrative**: "Look at the Queue. 12 pending optimizations. In a traditional DevOps team, that's a backlog. for Talos, that's 30 seconds of work. We're projecting $14,970 in annual savings just from today's actions."

---

## 🛑 Cleanup

Press `Ctrl+C` in your terminal to shut down the simulation.
