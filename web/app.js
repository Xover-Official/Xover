async function fetchSwarmState() {
    try {
        const response = await fetch('/api/live');
        const data = await response.json();
        updateDashboard(data);
    } catch (error) {
        console.error('Failed to fetch swarm state:', error);
    }
}

function updateDashboard(data) {
    // Update Action
    document.getElementById('current-action').textContent = data.current_action;
    
    // Update Queue
    document.getElementById('queue-depth').textContent = data.queue_depth;

    // Update Health
    const healthElem = document.getElementById('system-health');
    healthElem.textContent = Math.round(data.system_health) + '%';
    healthElem.style.color = data.system_health > 90 ? 'var(--accent-green)' : (data.system_health > 70 ? '#fbbf24' : '#ef4444');

    // Update Projected Savings
    document.getElementById('proj-savings').textContent = '$' + data.projected_savings.toLocaleString(undefined, {minimumFractionDigits: 0, maximumFractionDigits: 0});

    // Update Swarm Grid
    const grid = document.getElementById('swarm-grid');
    grid.innerHTML = '';

    data.tier_status.forEach(tier => {
        const div = document.createElement('div');
        div.className = `tier-box ${getTierClass(tier)}`;
        div.innerHTML = `
            <span class="tier-name">${tier.name}</span>
            <span class="tier-model">${tier.model}</span>
            <div style="margin-top: 0.5rem; font-size: 0.8rem;">
                ${tier.active ? '● Active' : '○ Idle'}
            </div>
        `;
        grid.appendChild(div);
    });

    // Simulate Savings (since API doesn't send it yet, we animate it)
    const savings = 1247.50 + (Math.random() * 0.5);
    document.getElementById('savings-amount').textContent = savings.toFixed(2);
}

function getTierClass(tier) {
    if (!tier.active) return '';
    if (tier.tier >= 3) return 'active-critical'; // Violet pulse for Arbiter+
    return 'active'; // Blue glow for others
}

// Poll every second
setInterval(fetchSwarmState, 1000);
fetchSwarmState();