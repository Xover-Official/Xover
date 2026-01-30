# 🔑 API Keys Configuration

## Required API Keys

Talos uses **OpenRouter** as the unified API gateway for all AI models, plus **Devin** for critical operations.

### 1. OpenRouter API Key (PRIMARY)

**Get it here**: <https://openrouter.ai/keys>

OpenRouter provides access to:

- ✅ Gemini 2.0 Flash (Tier 1 - Sentinel) - **FREE**
- ✅ Gemini Pro (Tier 2 - Strategist)
- ✅ Claude 3.5 Sonnet (Tier 3 - Arbiter)  
- ✅ GPT-4o Mini (Tier 4 - Reasoning)

**Cost**: Pay-as-you-go, typically $0.10-$5/day depending on usage

### 2. Devin API Key (OPTIONAL - Tier 5)

**Get it here**: <https://devin.ai/api>

Only needed for critical infrastructure changes (risk > 9.0)

**Cost**: ~$10 per request (rarely used)

### 3. OpenAI API Key (FALLBACK - Optional)

**Get it here**: <https://platform.openai.com/api-keys>

Used as fallback if OpenRouter is unavailable

---

## Setup Instructions

### Option 1: Environment Variables (Recommended)

```bash
# Linux/Mac
export OPENROUTER_API_KEY="sk-or-v1-YOUR-KEY-HERE"
export DEVIN_API_KEY="your-devin-key"       # Optional
export OPENAI_API_KEY="sk-YOUR-OPENAI-KEY"  # Optional fallback

# Windows PowerShell
$env:OPENROUTER_API_KEY="sk-or-v1-YOUR-KEY-HERE"
$env:DEVIN_API_KEY="your-devin-key"
$env:OPENAI_API_KEY="sk-YOUR-OPENAI-KEY"
```

### Option 2: Config File

Edit `config.yaml`:

```yaml
ai:
  openrouter_key: "sk-or-v1-YOUR-KEY-HERE"
  devin_key: "your-devin-key"       # Optional
  openai_key: "sk-YOUR-OPENAI-KEY"  # Optional fallback
```

### Option 3: Vault (Production)

For production deployments, store keys in HashiCorp Vault:

```bash
vault kv put secret/talos/api-keys \
  openrouter="sk-or-v1-YOUR-KEY" \
  devin="your-devin-key" \
  openai="sk-YOUR-OPENAI-KEY"
```

Then configure Talos to read from Vault:

```yaml
ai:
  vault_enabled: true
  vault_path: "secret/talos/api-keys"
```

---

## Security Best Practices

### ✅ DO

- Store keys in environment variables or Vault
- Use `.env` files (add to `.gitignore`)
- Rotate keys every 90 days
- Use different keys for dev/staging/prod
- Monitor usage for anomalies

### ❌ DON'T

- Commit keys to Git
- Share keys in Slack/email
- Use the same key across environments
- Store keys in plain text config files (unless `.gitignore`d)

---

## .env File Template

Create `.env` in project root:

```bash
# Talos API Keys
# DO NOT COMMIT THIS FILE TO GIT!

# Primary API (Required)
OPENROUTER_API_KEY=sk-or-v1-YOUR-KEY-HERE

# Tier 5 Oracle (Optional - only for critical ops)
DEVIN_API_KEY=your-devin-key

# Fallback (Optional)
OPENAI_API_KEY=sk-YOUR-OPENAI-KEY

# Database
DB_PASSWORD=your-secure-password

# Redis (if external)
REDIS_PASSWORD=your-redis-password

# Vault (if enabled)
VAULT_TOKEN=your-vault-token
```

Then add to `.gitignore`:

```bash
echo ".env" >> .gitignore
```

Load in shell:

```bash
# Linux/Mac
export $(cat .env | xargs)

# Or use direnv (recommended)
direnv allow
```

---

## Cost Estimation

### Typical Daily Usage (100 optimizations/day)

| Tier | Model | Calls/Day | Cost/Day |
|:-----|:------|:----------|:---------|
| Tier 1 | Gemini Flash | 60 | **$0.00** (Free) |
| Tier 2 | Gemini Pro | 25 | $0.50 |
| Tier 3 | Claude 3.5 | 10 | $2.00 |
| Tier 4 | GPT-4o Mini | 4 | $0.30 |
| Tier 5 | Devin | 1 | $10.00 |
| **Total** | | **100** | **~$12.80/day** |

**With 30% cache hit rate**: ~$9/day

**Monthly**: ~$270-400

**Cloud savings**: $1,500-3,000/month

**Net savings**: $1,100-2,700/month

**ROI**: 4-10x 🚀

---

## Monitoring Usage

### OpenRouter Dashboard
<https://openrouter.ai/activity>

Track:

- Requests per model
- Cost breakdown
- Rate limit status

### Talos Dashboard
<http://localhost:8080/api/token-breakdown>

```json
{
  "total_tokens": 1250000,
  "total_cost_usd": 87.50,
  "total_savings_usd": 4200.00,
  "net_roi": 4700.0,
  "model_breakdown": {
    "google/gemini-2.0-flash-exp": {
      "tokens": 750000,
      "cost_usd": 0.00,
      "requests": 1200
    },
    "anthropic/claude-3.5-sonnet": {
      "tokens": 300000,
      "cost_usd": 45.00,
      "requests": 100
    }
  }
}
```

---

## Troubleshooting

### "Invalid API key"

- Check key starts with `sk-or-v1-` for OpenRouter
- Verify key is set: `echo $OPENROUTER_API_KEY`
- Try regenerating key on provider dashboard

### "Rate limit exceeded"

- OpenRouter free tier: 200 req/day
- Upgrade to paid tier or add delay between requests
- Enable caching to reduce API calls by 30-50%

### "Model not available"

- Check model name in OpenRouter docs
- Some models require prepaid credits
- Fallback will automatically use cheaper model

### High costs

- Enable Redis caching (saves 30-50%)
- Increase risk threshold to use cheaper tiers
- Set daily spending limit on OpenRouter dashboard

---

## Testing API Keys

```bash
# Test OpenRouter
curl https://openrouter.ai/api/v1/auth/key \
  -H "Authorization: Bearer $OPENROUTER_API_KEY"

# Should return: {"data": {"label": "Talos"}}

# Test in Talos
go run cmd/atlas/main.go --test-api-keys
```

---

## Next Steps

1. ✅ Get OpenRouter API key (<https://openrouter.ai/keys>)
2. ✅ Add $10 credit to OpenRouter account
3. ✅ Set `OPENROUTER_API_KEY` environment variable
4. ✅ (Optional) Get Devin API key for Tier 5
5. ✅ Run `go run cmd/atlas/main.go` to start
6. ✅ Monitor usage in dashboard

**You're ready to launch!** 🚀
