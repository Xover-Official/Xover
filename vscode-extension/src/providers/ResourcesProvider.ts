import * as vscode from 'vscode';
import { TalosAPIClient } from '../api/client';

export class ResourcesProvider implements vscode.TreeDataProvider<ResourceItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ResourceItem | undefined | null | void> = new vscode.EventEmitter<ResourceItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ResourceItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private apiClient: TalosAPIClient) { }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: ResourceItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: ResourceItem): Promise<ResourceItem[]> {
        if (!element) {
            try {
                const resources = await this.apiClient.getResources();
                return resources.map(r => new ResourceItem(
                    r.name,
                    r.type,
                    r.provider,
                    r.cost_per_month,
                    r.optimization_score
                ));
            } catch (error) {
                return [new ResourceItem('No resources found', '', '', 0, 0)];
            }
        }
        return [];
    }
}

class ResourceItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly resourceType: string,
        public readonly provider: string,
        public readonly costPerMonth: number,
        public readonly optimizationScore: number
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);

        this.description = `$${costPerMonth.toFixed(2)}/mo`;
        this.tooltip = `${provider} ${resourceType} • Score: ${optimizationScore}/100`;

        // Icon based on optimization score
        if (optimizationScore >= 80) {
            this.iconPath = new vscode.ThemeIcon('check', new vscode.ThemeColor('charts.green'));
        } else if (optimizationScore >= 50) {
            this.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('charts.yellow'));
        } else {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('charts.red'));
        }

        this.contextValue = 'resource';
    }
}
