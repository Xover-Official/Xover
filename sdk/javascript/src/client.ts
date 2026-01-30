/**
 * Talos Guardian JavaScript/TypeScript SDK
 * Official client for browser and Node.js
 */

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

export interface SwarmStatus {
    active_tier: number;
    tier_status: TierStatus[];
    current_action: string;
    queue_depth: number;
}

export interface Resource {
    id: string;
    type: string;
    provider: string;
    region: string;
    cost_per_month: number;
    optimization_score: number;
    tags: Record<string, string>;
}

export interface ROI {
    ratio: number;
    total_savings: number;
    total_cost: number;
    net_profit: number;
}

export interface OptimizationRequest {
    type?: string;
    risk_limit?: number;
    dry_run?: boolean;
}

export interface OptimizationResponse {
    optimizations_found: number;
    estimated_savings: number;
    actions_applied: number;
    status: string;
}

export class TalosClient {
    private baseURL: string;
    private apiKey?: string;

    constructor(baseURL: string = 'http://localhost:8080', apiKey?: string) {
        this.baseURL = baseURL.replace(/\/$/, '');
        this.apiKey = apiKey;
    }

    private async request<T>(
        method: string,
        path: string,
        body?: any
    ): Promise<T> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
        };

        if (this.apiKey) {
            headers['Authorization'] = `Bearer ${this.apiKey}`;
        }

        const response = await fetch(`${this.baseURL}${path}`, {
            method,
            headers,
            body: body ? JSON.stringify(body) : undefined,
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || `HTTP ${response.status}`);
        }

        return response.json();
    }

    /**
     * Check system health
     */
    async health(): Promise<any> {
        return this.request('GET', '/health');
    }

    /**
     * Get AI swarm status
     */
    async getSwarmStatus(): Promise<SwarmStatus> {
        return this.request('GET', '/api/swarm/live');
    }

    /**
     * Run optimization
     */
    async runOptimization(
        options: OptimizationRequest = {}
    ): Promise<OptimizationResponse> {
        const payload = {
            type: options.type || 'full',
            risk_limit: options.risk_limit || 7.0,
            dry_run: options.dry_run !== undefined ? options.dry_run : true,
        };

        return this.request('POST', '/api/optimize', payload);
    }

    /**
     * Get cloud resources
     */
    async getResources(
        provider?: string,
        resourceType?: string
    ): Promise<Resource[]> {
        let path = '/api/resources';
        const params = new URLSearchParams();

        if (provider) params.append('provider', provider);
        if (resourceType) params.append('type', resourceType);

        if (params.toString()) {
            path += `?${params.toString()}`;
        }

        return this.request('GET', path);
    }

    /**
     * Get ROI metrics
     */
    async getROI(): Promise<ROI> {
        return this.request('GET', '/api/roi');
    }

    /**
     * Chat with AI
     */
    async chat(message: string): Promise<string> {
        const response = await this.request<{ response: string }>(
            'POST',
            '/api/ai/chat',
            { message }
        );
        return response.response;
    }

    /**
     * Get recommendations
     */
    async getRecommendations(): Promise<any[]> {
        return this.request('GET', '/api/recommendations');
    }

    /**
     * Subscribe to real-time updates via Server-Sent Events
     */
    subscribeToUpdates(
        callback: (event: any) => void
    ): EventSource {
        const eventSource = new EventSource(`${this.baseURL}/api/events`);

        eventSource.onmessage = (event) => {
            callback(JSON.parse(event.data));
        };

        eventSource.onerror = (error) => {
            console.error('EventSource error:', error);
        };

        return eventSource;
    }
}

// Example usage
export default TalosClient;

// For CommonJS
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { TalosClient };
}
