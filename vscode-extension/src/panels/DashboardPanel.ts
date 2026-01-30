import * as vscode from 'vscode';
import { TalosAPIClient } from '../api/client';

export class DashboardPanel {
    public static currentPanel: DashboardPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private _disposables: vscode.Disposable[] = [];

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, private apiClient: TalosAPIClient) {
        this._panel = panel;

        // Set initial HTML
        this._update();

        // Handle messages from webview
        this._panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'refresh':
                        this._update();
                        break;
                    case 'runOptimization':
                        this.handleOptimization(message.type);
                        break;
                }
            },
            null,
            this._disposables
        );

        // Cleanup
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);
    }

    public static render(context: vscode.ExtensionContext, apiClient: TalosAPIClient) {
        const column = vscode.window.activeTextEditor
            ? vscode.window.activeTextEditor.viewColumn
            : undefined;

        if (DashboardPanel.currentPanel) {
            DashboardPanel.currentPanel._panel.reveal(column);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'talosDashboard',
            'Talos Dashboard',
            column || vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        DashboardPanel.currentPanel = new DashboardPanel(panel, context.extensionUri, apiClient);
    }

    private async _update() {
        const webview = this._panel.webview;

        try {
            const swarmStatus = await this.apiClient.getSwarmStatus();
            const roi = await this.apiClient.getROI();

            this._panel.webview.html = this._getHtmlForWebview(webview, swarmStatus, roi);
        } catch (error) {
            this._panel.webview.html = this._getErrorHtml(String(error));
        }
    }

    private _getHtmlForWebview(webview: vscode.Webview, swarmStatus: any, roi: any): string {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Talos Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            padding: 20px;
            background: linear-gradient(135deg, #1e1e2e 0%, #2d2d44 100%);
            color: #fff;
        }
        .dashboard {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            font-size: 32px;
            font-weight: 700;
            margin-bottom: 30px;
            background: linear-gradient(90deg, #60a5fa 0%, #a78bfa 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border-radius: 16px;
            padding: 24px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .card-title {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 16px;
            color: #7dd3fc;
        }
        .stat {
            font-size: 36px;
            font-weight: 700;
            color: #60a5fa;
        }
        .tier-list {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }
        .tier {
            background: rgba(100, 150, 255, 0.1);
            padding: 12px;
            border-radius: 8px;
            border-left: 3px solid #60a5fa;
        }
        .tier.active {
            border-left-color: #50ff96;
            background: rgba(50, 255, 150, 0.1);
        }
        .btn {
            background: linear-gradient(90deg, #60a5fa 0%, #a78bfa 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            cursor: pointer;
            font-weight: 600;
            transition: transform 0.2s;
        }
        .btn:hover {
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="dashboard">
        <div class="header">🚀 Talos Guardian Dashboard</div>
        
        <div class="grid">
            <div class="card">
                <div class="card-title">💰 ROI</div>
                <div class="stat">${roi.ratio.toFixed(1)}x</div>
                <div>Saved: $${roi.totalSavings.toFixed(2)}</div>
                <div>Cost: $${roi.totalCost.toFixed(2)}</div>
            </div>

            <div class="card">
                <div class="card-title">🤖 Active Tier</div>
                <div class="stat">T${swarmStatus.active_tier}</div>
                <div>${swarmStatus.current_action || 'Idle'}</div>
            </div>

            <div class="card">
                <div class="card-title">📋 Queue</div>
                <div class="stat">${swarmStatus.queue_depth}</div>
                <div>Pending optimizations</div>
            </div>
        </div>

        <div class="card">
            <div class="card-title">AI Swarm Status</div>
            <div class="tier-list">
                ${swarmStatus.tier_status.map((tier: any) => `
                    <div class="tier ${tier.active ? 'active' : ''}">
                        <strong>T${tier.tier}: ${tier.name}</strong> • ${tier.model}<br>
                        Requests: ${tier.requests_today} | Latency: ${Math.round(tier.avg_latency_ms)}ms | Success: ${tier.success_rate}%
                    </div>
                `).join('')}
            </div>
        </div>

        <div style="margin-top: 20px; text-align: center;">
            <button class="btn" onclick="runOptimization()">🚀 Run Optimization</button>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();
        
        function runOptimization() {
            vscode.postMessage({ command: 'runOptimization', type: 'full' });
        }

        // Auto-refresh every 5 seconds
        setInterval(() => {
            vscode.postMessage({ command: 'refresh' });
        }, 5000);
    </script>
</body>
</html>`;
    }

    private _getErrorHtml(error: string): string {
        return `<!DOCTYPE html>
<html>
<body>
    <h1>❌ Dashboard Error</h1>
    <p>Failed to connect to Talos backend:</p>
    <pre>${error}</pre>
    <p>Make sure the Talos backend is running on the configured URL.</p>
</body>
</html>`;
    }

    private async handleOptimization(type: string) {
        try {
            await this.apiClient.runOptimization(type);
            vscode.window.showInformationMessage('✅ Optimization complete!');
            this._update();
        } catch (error) {
            vscode.window.showErrorMessage(`❌ Optimization failed: ${error}`);
        }
    }

    public dispose() {
        DashboardPanel.currentPanel = undefined;
        this._panel.dispose();

        while (this._disposables.length) {
            const disposable = this._disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
