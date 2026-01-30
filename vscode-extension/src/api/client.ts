import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';

export interface SwarmStatus {
    timestamp: string;
    active_tier: number;
    tier_status: TierStatus[];
    current_action: string;
    queue_depth: number;
}

export interface TierStatus {
    tier: number;
    name: string;
    model: string;
    active: boolean;
    requests_today: number;
    avg_latency_ms: number;
    success_rate: number;
    status: string;
}

export interface ROIData {
    ratio: number;
    totalSavings: number;
    totalCost: number;
    netProfit: number;
}

export class TalosAPIClient {
    private client: AxiosInstance;
    private context: vscode.ExtensionContext;
    private baseURL: string;

    constructor(baseURL: string, context: vscode.ExtensionContext) {
        this.baseURL = baseURL;
        this.context = context;
        this.client = axios.create({
            baseURL,
            timeout: 10000,
            headers: {
                'Content-Type': 'application/json'
            }
        });

        // Add auth interceptor
        this.client.interceptors.request.use(async (config) => {
            const token = await this.getAuthToken();
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        });
    }

    updateBaseURL(url: string) {
        this.baseURL = url;
        this.client.defaults.baseURL = url;
    }

    async connect(): Promise<boolean> {
        try {
            const response = await this.client.get('/healthz');
            return response.status === 200;
        } catch (error) {
            throw new Error(`Failed to connect to Talos backend at ${this.baseURL}`);
        }
    }

    async getSwarmStatus(): Promise<SwarmStatus> {
        const response = await this.client.get<SwarmStatus>('/api/swarm/live');
        return response.data;
    }

    async getROI(): Promise<ROIData> {
        const response = await this.client.get<ROIData>('/api/roi');
        return response.data;
    }

    async runOptimization(type: string): Promise<any> {
        const response = await this.client.post('/api/optimize', { type });
        return response.data;
    }

    async getHistory(): Promise<any[]> {
        const response = await this.client.get('/api/history');
        return response.data;
    }

    async getResources(): Promise<any[]> {
        const response = await this.client.get('/api/resources');
        return response.data;
    }

    async getRecommendations(): Promise<any[]> {
        const response = await this.client.get('/api/recommendations');
        return response.data;
    }

    async chatWithAI(message: string): Promise<string> {
        const response = await this.client.post('/api/ai/chat', { message });
        return response.data.response;
    }

    private async getAuthToken(): Promise<string | undefined> {
        return await this.context.secrets.get('talos.authToken');
    }

    async setAuthToken(token: string) {
        await this.context.secrets.store('talos.authToken', token);
    }
}
