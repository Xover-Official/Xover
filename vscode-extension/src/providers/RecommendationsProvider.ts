import * as vscode from 'vscode';
import { TalosAPIClient } from '../api/client';

export class RecommendationsProvider implements vscode.TreeDataProvider<RecommendationItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<RecommendationItem | undefined | null | void> = new vscode.EventEmitter<RecommendationItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<RecommendationItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private apiClient: TalosAPIClient) { }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: RecommendationItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: RecommendationItem): Promise<RecommendationItem[]> {
        if (!element) {
            try {
                const recommendations = await this.apiClient.getRecommendations();
                return recommendations.map(r => new RecommendationItem(
                    r.title,
                    r.description,
                    r.savings,
                    r.risk_score,
                    r.action
                ));
            } catch (error) {
                return [new RecommendationItem('No recommendations', 'System is optimized', 0, 0, null)];
            }
        }
        return [];
    }
}

class RecommendationItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly description: string,
        public readonly savings: number,
        public readonly riskScore: number,
        public readonly action: any
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);

        this.description = savings > 0 ? `Save $${savings.toFixed(2)}/mo` : '';
        this.tooltip = `${description}\nRisk: ${riskScore}/10`;

        // Icon based on risk
        if (riskScore < 3) {
            this.iconPath = new vscode.ThemeIcon('lightbulb', new vscode.ThemeColor('charts.green'));
        } else if (riskScore < 7) {
            this.iconPath = new vscode.ThemeIcon('info', new vscode.ThemeColor('charts.yellow'));
        } else {
            this.iconPath = new vscode.ThemeIcon('alert', new vscode.ThemeColor('charts.red'));
        }

        if (action) {
            this.command = {
                command: 'talos.applyRecommendation',
                title: 'Apply Recommendation',
                arguments: [action]
            };
        }

        this.contextValue = 'recommendation';
    }
}
