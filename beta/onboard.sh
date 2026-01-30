#!/bin/bash
# Beta Tester Onboarding Script
# Run this to automatically set up a beta tester environment

set -e

echo "🧪 Talos Beta Onboarding"
echo "======================="
echo ""

# Step 1: Collect user information
read -p "Enter your email: " BETA_EMAIL
read -p "Enter your cloud provider (aws/azure/gcp): " CLOUD_PROVIDER
read -p "Enter your company name: " COMPANY_NAME

echo ""
echo "✅ Creating your beta environment..."

# Step 2: Generate unique beta ID
BETA_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

# Step 3: Create beta user config
mkdir -p ~/.talos/beta
cat > ~/.talos/beta/user.json <<EOF
{
  "beta_id": "$BETA_ID",
  "email": "$BETA_EMAIL",
  "cloud_provider": "$CLOUD_PROVIDER",
  "company": "$COMPANY_NAME",
  "joined_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "beta_version": "1.0.0-beta"
}
EOF

# Step 4: Clone starter config
cat > ~/.talos/config.yaml <<EOF
guardian:
  mode: "personal"
  risk_threshold: 5.0
  indie_force: true
  dry_run: true  # Start in safe mode for beta

ai:
  openrouter_key: ""  # Add your key here
  devin_key: ""       # Optional
  gpt_5_key: ""       # Optional

database:
  type: "sqlite"  # Beta users can use SQLite

network:
  dashboard_port: 8080
  enable_sse: true

beta:
  enabled: true
  user_id: "$BETA_ID"
  slack_webhook: "https://hooks.slack.com/beta-talos"  # For feedback
EOF

echo "✅ Configuration created!"
echo ""

# Step 5: Download and install
echo "📥 Installing Talos..."
curl -fsSL https://get.talos.dev/beta | bash

# Step 6: Send welcome notification
echo "📧 Sending welcome email..."
curl -X POST https://api.talos.dev/beta/onboard \
  -H "Content-Type: application/json" \
  -d "{\"beta_id\": \"$BETA_ID\", \"email\": \"$BETA_EMAIL\"}"

echo ""
echo "🎉 Onboarding Complete!"
echo ""
echo "📋 Next Steps:"
echo "   1. Edit ~/.talos/config.yaml and add your API keys"
echo "   2. Join beta Slack: https://talos-beta.slack.com/join/$BETA_ID"
echo "   3. Watch tutorial video: https://youtu.be/talos-beta-tutorial"
echo "   4. Run: cd ~/.talos && docker-compose up -d"
echo ""
echo "📞 Need help? Reply to the welcome email or ping us in Slack!"
echo ""
echo "Your Beta ID: $BETA_ID"
echo "Keep this for future reference."
