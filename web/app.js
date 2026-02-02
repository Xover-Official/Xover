/**
 * Talos Atlas Web Orchestration
 * Premium Dashboard Logic
 */

const state = {
    savings: 1247.50,
    progress: 45,
    health: 98
};

async function fetchSwarmState() {
    try {
        const response = await fetch('/api/live');
        if (!response.ok) throw new Error('API Sync Failed');
        
        const data = await response.json();
        updateDashboard(data);
    } catch (error) {
        console.warn('⚡ API Offline - Using Simulated Stream');
        simulateData();
    }
}

function updateDashboard(data) {
    // Update main action text
    if (data.current_action) {
        const actionElem = document.getElementById('current-action');
        if (actionElem.textContent !== data.current_action) {
            actionElem.style.opacity = '0';
            setTimeout(() => {
                actionElem.textContent = data.current_action;
                actionElem.style.opacity = '1';
            }, 300);
        }
    }
    
    // Update Queue
    if (data.queue_depth !== undefined) {
        document.getElementById('queue-depth').textContent = data.queue_depth;
    }

    // Update Health Circle
    if (data.system_health !== undefined) {
        const health = Math.round(data.system_health);
        document.getElementById('system-health').textContent = health + '%';
        const circle = document.getElementById('health-circle');
        circle.setAttribute('stroke-dasharray', `${health}, 100`);
        
        // Color shifts based on health
        if (health > 95) circle.style.stroke = 'var(--accent-success)';
        else if (health > 80) circle.style.stroke = 'var(--accent-warning)';
        else circle.style.stroke = 'var(--accent-error)';
    }

    // Update Projected Savings
    if (data.projected_savings !== undefined) {
        document.getElementById('proj-savings').textContent = 
            data.projected_savings.toLocaleString(undefined, {minimumFractionDigits: 0, maximumFractionDigits: 0});
    }

    // Update Swarm Grid elegantly
    if (data.tier_status) {
        const grid = document.getElementById('swarm-grid');
        grid.innerHTML = '';

        data.tier_status.forEach((tier, index) => {
            const div = document.createElement('div');
            const isActive = tier.active;
            const isHigh = tier.tier >= 3;
            
            div.className = `agent-node ${isActive ? (isHigh ? 'active-high' : 'active') : ''}`;
            div.style.animationDelay = `${index * 0.1}s`;
            
            div.innerHTML = `
                <span class="agent-role">${tier.name}</span>
                <span class="agent-model">${tier.model}</span>
                <div class="agent-status" style="color: ${isActive ? (isHigh ? 'var(--accent-secondary)' : 'var(--accent-success)') : 'var(--text-dim)'}">
                    ${isActive ? '● Processing' : '○ Standby'}
                </div>
            `;
            grid.appendChild(div);
        });
    }

    // Live Savings Ticker
    if (data.actual_savings !== undefined) {
        animateValue('savings-amount', state.savings, data.actual_savings, 2000);
        state.savings = data.actual_savings;
    }
}

// Fallback simulation for visual demo
function simulateData() {
    const mockData = {
        current_action: "Optimizing " + (Math.random() > 0.5 ? "RDS" : "EC2") + " instance clusters...",
        queue_depth: Math.floor(Math.random() * 20) + 5,
        system_health: 95 + (Math.random() * 4),
        projected_savings: 14000 + (Math.random() * 2000),
        actual_savings: state.savings + (Math.random() * 0.05),
        tier_status: [
            { name: "Sentinel", model: "Gemini 2.0 Flash", active: true, tier: 1 },
            { name: "Strategist", model: "Gemini Pro 1.5", active: Math.random() > 0.5, tier: 2 },
            { name: "Arbiter", model: "Claude 3.5 Sonnet", active: Math.random() > 0.8, tier: 3 },
            { name: "Reasoning", model: "GPT-5 Mini", active: Math.random() > 0.7, tier: 3 },
            { name: "Oracle", model: "Devin Oracle", active: Math.random() > 0.9, tier: 4 }
        ]
    };

    updateDashboard(mockData);
    
    // Animate progress bar locally
    state.progress = (state.progress + 1) % 100;
    document.getElementById('action-progress').style.width = state.progress + '%';
    document.getElementById('progress-val').textContent = state.progress + '%';
}

function animateValue(id, start, end, duration) {
    const obj = document.getElementById(id);
    if (!obj) return;
    const range = end - start;
    let startTime = null;

    function step(timestamp) {
        if (!startTime) startTime = timestamp;
        const progress = Math.min((timestamp - startTime) / duration, 1);
        const value = (start + (progress * range)).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2});
        obj.innerHTML = value;
        if (progress < 1) {
            window.requestAnimationFrame(step);
        }
    }
    window.requestAnimationFrame(step);
}

// Init
setInterval(fetchSwarmState, 2000);
fetchSwarmState();
console.log('🛡️ Talos Atlas Intelligence Interface Initialized');