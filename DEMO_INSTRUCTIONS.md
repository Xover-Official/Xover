# 🚀 Talos Atlas - Sales Demo Instructions

## Overview

The Talos Atlas platform now includes a complete "Sales Funnel" experience:

1. **Landing Page**: High-conversion landing page (`landing.html`)
2. **Demo Access**: Simulated SSO Login (`login.html`)
3. **Premium Dashboard**: Full interactive dashboard with OODA loop visualization (`index.html`)

## How to Run the Demo

### Option 1: Quick Frontend Demo (Recommended for Sales)

Since the frontend is now fully simulated for demo purposes, you can run it with any static file server.

**Using Python:**

```powershell
python -m http.server 8080 --directory web
```

Then open: [http://localhost:8080/landing.html](http://localhost:8080/landing.html)

### Option 2: Full Enterprise Server

If you have the Go environment configured:

```powershell
go run cmd/dashboard/main.go
```

Then open: [http://localhost:8080/landing.html](http://localhost:8080/landing.html)

## Demo Script

1. **Start at Landing Page**: Scroll through the "Features" and "Pricing".
2. **Click 'Start Optimizing Now'**: This takes you to the Login Portal.
3. **Authentication**: Click "Continue with Okta" or "Launch Interactive Demo".
    * *Note: This simulates a secure handshake and redirects you automatically.*
4. **Dashboard Reveal**: Show the **OODA Loop** tracker and **Live Neural Feed**.
5. **Interactive Elements**:
    * Watch the **Realized Savings** ticker increment.
    * Click "APPROVE" on a resource to show the risk-modal.
    * The "Run Guardian" button triggers a simulated optimization cycle.

## Key Talking Points

* **"Zero-Sum Learning"**: Explain how the AI learns from every optimization.
* **"Antifragile Infrastructure"**: We don't just cut costs; we increase stability.
* **"Swarm Intelligence"**: Show the grid of agents (Sentinel, Strategist, etc.) working in parallel.
