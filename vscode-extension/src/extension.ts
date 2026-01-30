import * as vscode from 'vscode';
import { TalosAPIClient } from './api/client';
import { DashboardPanel } from './panels/DashboardPanel';
import { AIConsolePanel } from './panels/AIConsolePanel';
import { SwarmStatusProvider } from './providers/SwarmStatusProvider';
import { ResourcesProvider } from './providers/ResourcesProvider';
import { RecommendationsProvider } from './providers/RecommendationsProvider';

export function activate(context: vscode.ExtensionContext) {
    console.log('🚀 Talos Guardian extension activated');

    // Initialize API client
    const config = vscode.workspace.getConfiguration('talos');
    const backendUrl = config.get<string>('backend.url', 'http://localhost:8080');
    const apiClient = new TalosAPIClient(backendUrl, context);

    // Register providers
    const swarmProvider = new SwarmStatusProvider(apiClient);
    const resourcesProvider = new ResourcesProvider(apiClient);
    const recommendationsProvider = new RecommendationsProvider(apiClient);

    vscode.window.registerTreeDataProvider('talos.swarmStatus', swarmProvider);
    vscode.window.registerTreeDataProvider('talos.resources', resourcesProvider);
    vscode.window.registerTreeDataProvider('talos.recommendations', recommendationsProvider);

    // Command: Show Dashboard
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.showDashboard', () => {
            DashboardPanel.render(context, apiClient);
        })
    );

    // Command: Connect Backend
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.connectBackend', async () => {
            const url = await vscode.window.showInputBox({
                prompt: 'Enter Talos backend URL',
                value: backendUrl,
                placeHolder: 'http://localhost:8080'
            });

            if (url) {
                await config.update('backend.url', url, vscode.ConfigurationTarget.Global);
                apiClient.updateBaseURL(url);
                vscode.window.showInformationMessage(`✅ Connected to ${url}`);

                // Refresh providers
                swarmProvider.refresh();
                resourcesProvider.refresh();
                recommendationsProvider.refresh();
            }
        })
    );

    // Command: Run Optimization
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.runOptimization', async () => {
            const result = await vscode.window.showQuickPick(
                ['Optimize Current Project', 'Scan Cloud Resources', 'Analyze Costs'],
                { placeHolder: 'Select optimization type' }
            );

            if (result) {
                vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: `Talos: ${result}`,
                    cancellable: false
                }, async (progress) => {
                    progress.report({ increment: 0 });

                    try {
                        const response = await apiClient.runOptimization(result);
                        progress.report({ increment: 100 });

                        vscode.window.showInformationMessage(
                            `✅ Optimization complete! Saved $${response.savings}/month`
                        );
                    } catch (error) {
                        vscode.window.showErrorMessage(`❌ Optimization failed: ${error}`);
                    }
                });
            }
        })
    );

    // Command: AI Console
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.showAIConsole', () => {
            AIConsolePanel.render(context, apiClient);
        })
    );

    // Command: View ROI
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.viewROI', async () => {
            try {
                const roi = await apiClient.getROI();
                vscode.window.showInformationMessage(
                    `💰 ROI: ${roi.ratio.toFixed(1)}x | Saved: $${roi.totalSavings} | Cost: $${roi.totalCost}`
                );
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to fetch ROI: ${error}`);
            }
        })
    );

    // Command: Toggle Autonomous Mode
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.toggleAutonomous', async () => {
            const current = config.get<boolean>('autonomous.enabled', false);
            const newValue = !current;

            if (newValue) {
                const confirm = await vscode.window.showWarningMessage(
                    '⚠️ Autonomous mode allows AI to make changes automatically. Continue?',
                    'Enable',
                    'Cancel'
                );

                if (confirm !== 'Enable') {
                    return;
                }
            }

            await config.update('autonomous.enabled', newValue, vscode.ConfigurationTarget.Global);
            vscode.window.showInformationMessage(
                `🤖 Autonomous mode: ${newValue ? 'ENABLED' : 'DISABLED'}`
            );
        })
    );

    // Command: View History
    context.subscriptions.push(
        vscode.commands.registerCommand('talos.viewHistory', async () => {
            try {
                const history = await apiClient.getHistory();
                // TODO: Show history in webview
                vscode.window.showInformationMessage(`📜 ${history.length} actions in history`);
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to load history: ${error}`);
            }
        })
    );

    // Status bar item
    const statusBarItem = vscode.window.createStatusBarItem(
        vscode.StatusBarAlignment.Left,
        100
    );
    statusBarItem.text = '$(pulse) Talos';
    statusBarItem.tooltip = 'Talos Guardian is active';
    statusBarItem.command = 'talos.showDashboard';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Auto-connect on startup
    apiClient.connect().then(() => {
        vscode.window.showInformationMessage('✅ Talos Guardian connected');
        swarmProvider.refresh();
        resourcesProvider.refresh();
        recommendationsProvider.refresh();
    }).catch((error) => {
        vscode.window.showWarningMessage(`⚠️ Talos backend not reachable: ${error.message}`);
    });

    // Watch for config changes
    vscode.workspace.onDidChangeConfiguration(e => {
        if (e.affectsConfiguration('talos.backend.url')) {
            const newUrl = config.get<string>('backend.url');
            if (newUrl) {
                apiClient.updateBaseURL(newUrl);
                apiClient.connect();
            }
        }
    });
}

export function deactivate() {
    console.log('🛑 Talos Guardian extension deactivated');
}
