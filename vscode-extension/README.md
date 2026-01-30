# Talos Guardian - VS Code Extension

Transform VS Code into an AI-powered cloud optimization command center.

## 🚀 Features

- **AI Swarm Dashboard**: Real-time visualization of 5-tier AI system
- **Cost Optimization**: Track ROI, savings, and cloud spend
- **Cloud Resources**: View and manage resources across AWS, GCP, Azure
- **AI Console**: Chat directly with the AI swarm
- **Autonomous Mode**: Let AI automatically optimize your infrastructure
- **Recommendations**: Get actionable optimization suggestions

## 📦 Installation

### From VS Code Marketplace

1. Open VS Code
2. Press `Ctrl+Shift+X` (Windows/Linux) or `Cmd+Shift+X` (Mac)
3. Search for "Talos Guardian"
4. Click Install

### From Source

```bash
cd vscode-extension
npm install
npm run compile
# Press F5 to launch extension in debug mode
```

## ⚙️ Configuration

1. Make sure Talos backend is running:

```bash
cd .. # Navigate to Talos root
go run cmd/atlas/main.go
```

1. Configure extension settings:
   - `talos.backend.url`: Backend URL (default: `http://localhost:8080`)
   - `talos.autonomous.enabled`: Enable autonomous mode
   - `talos.ui.theme`: Dashboard theme (glassmorphic/dark/light)

## 🎯 Quick Start

1. Press `Ctrl+Shift+P` (Windows/Linux) or `Cmd+Shift+P` (Mac)
2. Type `Talos` to see available commands:
   - **Talos: Show Dashboard** - Open main dashboard
   - **Talos: AI Console** - Chat with AI
   - **Talos: Run Optimization** - Start optimization
   - **Talos: View ROI** - See cost metrics
   - **Talos: Toggle Autonomous Mode** - Enable/disable AI autonomy

## 📊 Dashboard Features

### AI Swarm Status

- Live tier activity visualization
- Token consumption tracking
- Success rate metrics
- Latency monitoring

### Cloud Resources

- Real-time resource inventory
- Cost per resource
- Optimization scores
- Provider breakdown

### Recommendations

- AI-generated optimization suggestions
- Estimated savings
- Risk assessment
- One-click application

## 🤖 Autonomous Mode

Enable autonomous mode to let AI:

- Automatically optimize resources
- Right-size instances
- Clean up unused resources
- Implement cost-saving measures

**Safety Features:**

- Dry-run preview
- Approval workflows
- Automatic rollback
- Audit trail

## 🔐 Security

- Local-only mode (no external calls)
- Secure JWT authentication
- Optional mTLS for remote backends
- Secrets stored in VS Code SecretStorage

## 📖 Commands

| Command | Description |
|:--------|:------------|
| `talos.showDashboard` | Open Talos dashboard |
| `talos.connectBackend` | Configure backend connection |
| `talos.runOptimization` | Run optimization workflow |
| `talos.showAIConsole` | Open AI chat console |
| `talos.viewROI` | View ROI metrics |
| `talos.toggleAutonomous` | Toggle autonomous mode |
| `talos.viewHistory` | View action history |

## 🛠️ Development

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Watch mode
npm run watch

# Package extension
npm run package
```

## 📝 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions welcome! Please read CONTRIBUTING.md first.

## 💰 Pricing

- **Free**: Local-only mode
- **Pro ($29/mo)**: Remote backend, full AI swarm
- **Enterprise**: Custom pricing, white-label, SLA

## 🔗 Links

- [Documentation](https://talos.dev/docs)
- [GitHub](https://github.com/talos/extension)
- [Discord Community](https://discord.gg/talos)

---

**Made with ❤️ by the Talos team**
