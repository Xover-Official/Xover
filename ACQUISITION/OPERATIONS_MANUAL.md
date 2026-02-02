# 🕹️ Operations Manual: Standing Up the Business

**Objective**: Run Talos Atlas as a "High-Margin, Low-Touch" Micro-SaaS.
**Time Commitment**: ~4 Hours / Week.

---

## 📅 The Monthly "Cadence"

### Week 1: Billing & Growth

* **Action**: Run the `analytics/revenue_report.go` script.
* **Result**: Generates a CSV of all customers and their "Saved Amount."
* **Task**: Upload CSV to Stripe/ChurnZero to trigger "Percentage of Savings" invoices.
* **Time**: 1 Hour.

### Week 2: AI Tuning (The "Gardening")

* **Action**: Review the `ProjectMemory` ledger for "Failed Actions" (Risk Score > 8.0 that were rejected).
* **Task**: If the AI is being too cautious, adjust the `RiskThreshold` in `config.yaml` from 5.0 -> 6.0.
* **Note**: This is the "Secret Sauce" tuning. It keeps the product sticky.
* **Time**: 1 Hour.

### Week 3: Marketing Automation

* **Action**: Post a snapshot of the aggregate "Global Savings" from the Dashboard to LinkedIn/Twitter.
* **Content**: "Talos blocked 4,000 idle instances this week. That's $12k returned to founders."
* **Time**: 30 Minutes.

### Week 4: Software Updates

* **Action**: Merge dependency updates (`go get -u ./...`).
* **Task**: Check for new models from OpenAI/Anthropic. If "GPT-5" drops, update the `adapter` string strings.
* **Time**: 1.5 Hours.

---

## 🛠️ The Support Playbook (FAQ)

**Q: "Talos shut down a server I needed!"**
**A**:

1. Check the Dashboard "Audit Log" tab.
2. Find the Resource ID.
3. Click "Rollback" (Talos will spin it back up via Terraform/AWS API).
4. Tag the resource `talos:ignore=true` to prevent future touches.

**Q: "How do I install this on Azure?"**
**A**: Send them the `docs/AZURE_DEPLOY.md`. It's a one-line Docker command.

**Q: "Is my data safe?"**
**A**: Remind them: "Talos runs on *your* servers. We don't see your data. Check the `TECHNICAL_DUE_DILIGENCE.md` for the architecture proof."

---

## 🚨 Emergency Protocols ("Red Button")

If the AI starts behaving erratically (e.g., trying to delete everything):

1. **Kill Switch**: Set environment variable `TALOS_GLOBAL_LOCK=true`.
    * This instantly freezes all `ACT` phase operations globally.
    * Only `OBSERVE` (ReadOnly) will continue.
2. **Diagnostics**: Check the `zap` logs for `error` level events.
3. **Restore**: Once patched, set `TALOS_GLOBAL_LOCK=false`.

*(Note: We have never had to use this in production. The `Guardian` logic prevents mass-deletion events proactively.)*
