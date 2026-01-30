// Talos AI Swarm Visualization
// Real-time visualization of AI tier activity

class SwarmVisualizer {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.updateInterval = null;
        this.init();
    }

    init() {
        this.render();
        this.startUpdates();
    }

    render() {
        this.container.innerHTML = `
            <div class="swarm-container">
                <h2 class="swarm-title">AI Swarm Status</h2>
                <div class="swarm-grid" id="swarm-grid"></div>
                <div class="swarm-activity" id="swarm-activity"></div>
            </div>
        `;
    }

    async update() {
        try {
            const response = await fetch('/api/swarm/live');
            const data = await response.json();

            this.updateGrid(data.tier_status, data.active_tier);
            this.updateActivity(data.current_action, data.queue_depth);
        } catch (error) {
            console.error('Failed to update swarm visualization:', error);
        }
    }

    updateGrid(tiers, activeTier) {
        const grid = document.getElementById('swarm-grid');

        grid.innerHTML = tiers.map(tier => `
            <div class="tier-node ${tier.active ? 'active' : ''} ${tier.tier === activeTier ? 'current' : ''}" 
                 data-tier="${tier.tier}">
                <div class="tier-header">
                    <span class="tier-number">T${tier.tier}</span>
                    <span class="tier-name">${tier.name}</span>
                </div>
                <div class="tier-model">${tier.model}</div>
                <div class="tier-stats">
                    <div class="stat">
                        <span class="stat-label">Requests</span>
                        <span class="stat-value">${tier.requests_today}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">Latency</span>
                        <span class="stat-value">${Math.round(tier.avg_latency_ms)}ms</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">Success</span>
                        <span class="stat-value">${tier.success_rate.toFixed(1)}%</span>
                    </div>
                </div>
                <div class="tier-status status-${tier.status}">
                    ${tier.status}
                </div>
                ${tier.active ? '<div class="pulse"></div>' : ''}
            </div>
        `).join('');
    }

    updateActivity(action, queueDepth) {
        const activity = document.getElementById('swarm-activity');

        activity.innerHTML = `
            <div class="current-action">
                <span class="action-icon">🤖</span>
                <span class="action-text">${action || 'Idle'}</span>
            </div>
            ${queueDepth > 0 ? `
                <div class="queue-status">
                    <span class="queue-icon">📋</span>
                    <span class="queue-text">${queueDepth} in queue</span>
                </div>
            ` : ''}
        `;
    }

    startUpdates() {
        // Initial update
        this.update();

        // Update every 2 seconds
        this.updateInterval = setInterval(() => this.update(), 2000);
    }

    stop() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
        }
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    const visualizer = new SwarmVisualizer('swarm-visualization');
});
