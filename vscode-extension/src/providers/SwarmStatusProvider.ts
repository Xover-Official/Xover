import * as vscode from 'vscode';
import { TalosAPIClient, TierStatus } from '../api/client';

export class SwarmStatusProvider implements vscode.TreeDataProvider<SwarmItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<SwarmItem | undefined | null | void> = new vscode.EventEmitter<SwarmItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<SwarmItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private apiClient: TalosAPIClient) { }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: SwarmItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: SwarmItem): Promise<SwarmItem[]> {
        if (!element) {
            try {
                const status = await this.apiClient.getSwarmStatus();
                return status.tier_status.map(tier => new SwarmItem(
                    `T${tier.tier}: ${tier.name}`,
                    tier.model,
                    tier.active,
                    tier.status,
                    [
                        `Requests: ${tier.requests_today}`,
                        `Latency: ${Math.round(tier.avg_latency_ms)}ms`,
                        `Success: ${tier.success_rate.toFixed(1)}%`
                    ],
                    vscode.TreeItemCollapsibleState.Collapsed
                ));
            } catch (error) {
                return [new SwarmItem('❌ Backend Offline', 'Check connection', false, 'error', [], vscode.TreeItemCollapsibleState.None)];
            }
        } else {
            return element.stats.map(stat => new SwarmItem(stat, '', false, 'info', [], vscode.TreeItemCollapsibleState.None));
        }
    }
}

class SwarmItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly description: string,
        public readonly active: boolean,
        public readonly status: string,
        public readonly stats: string[],
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);

        this.description = description;

        // Set icon based on status
        if (active) {
            this.iconPath = new vscode.ThemeIcon('pulse', new vscode.ThemeColor('charts.green'));
        } else if (status === 'healthy') {
            this.iconPath = new vscode.ThemeIcon('circle-filled', new vscode.ThemeColor('charts.green'));
        } else if (status === 'degraded') {
            this.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('charts.yellow'));
        } else if (status === 'error') {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('charts.red'));
        } else {
            this.iconPath = new vscode.ThemeIcon('circle-outline');
        }
    }
}
