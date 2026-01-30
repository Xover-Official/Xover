// Enhanced Dashboard with Real-Time AI Feedback
// Add this to web/main.js

// ============================================
// AI ACTION FEEDBACK SYSTEM
// ============================================

const aiTierColors = {
    'gemini-2.0-flash-exp': '#00f2ff',      // Cyan - Sentinel
    'gemini-1.5-pro': '#7c3aed',            // Purple - Strategist
    'anthropic/claude-3.5-sonnet': '#f59e0b', // Amber - Arbiter
    'openai/gpt-5-mini': '#10b981',         // Green - Reasoning
    'devin': '#ef4444'                       // Red - Oracle
};

const aiTierNames = {
    'gemini-2.0-flash-exp': 'Sentinel',
    'gemini-1.5-pro': 'Strategist',
    'anthropic/claude-3.5-sonnet': 'Arbiter',
    'openai/gpt-5-mini': 'Reasoning Engine',
    'devin': 'Oracle'
};

// Real-time AI action notification
function showAIAction(action) {
    const container = document.getElementById('ai-action-feed');
    if (!container) return;

    const actionCard = document.createElement('div');
    actionCard.className = 'ai-action-card';
    actionCard.style.borderLeft = `4px solid ${aiTierColors[action.model] || '#666'}`;

    const timestamp = new Date().toLocaleTimeString();

    actionCard.innerHTML = `
        <div class="ai-action-header">
            <span class="ai-tier-badge" style="background: ${aiTierColors[action.model]}20; color: ${aiTierColors[action.model]}">
                ${aiTierNames[action.model] || action.model}
            </span>
            <span class="ai-action-time">${timestamp}</span>
        </div>
        <div class="ai-action-body">
            <div class="ai-action-resource">${action.resource}</div>
            <div class="ai-action-decision">${action.decision}</div>
            ${action.reasoning ? `<div class="ai-action-reasoning">"${action.reasoning}"</div>` : ''}
        </div>
        <div class="ai-action-footer">
            <span class="ai-action-risk" style="color: ${getRiskColor(action.risk)}">
                Risk: ${action.risk.toFixed(1)}
            </span>
            <span class="ai-action-savings" style="color: #10b981">
                +$${action.savings.toFixed(2)}/mo
            </span>
        </div>
    `;

    // Add with animation
    actionCard.style.opacity = '0';
    actionCard.style.transform = 'translateY(-20px)';
    container.insertBefore(actionCard, container.firstChild);

    // Animate in
    setTimeout(() => {
        actionCard.style.transition = 'all 0.3s ease';
        actionCard.style.opacity = '1';
        actionCard.style.transform = 'translateY(0)';
    }, 10);

    // Remove old cards (keep last 10)
    while (container.children.length > 10) {
        container.removeChild(container.lastChild);
    }

    // Add to action log
    addActionToLog({
        time: timestamp,
        resource: action.resource,
        action: action.decision,
        risk: action.risk,
        savings: action.savings,
        status: 'success'
    });
}

function getRiskColor(risk) {
    if (risk < 3) return '#10b981';
    if (risk < 7) return '#f59e0b';
    return '#ef4444';
}

// AI Tier Indicator Component
function createAITierIndicator() {
    const indicator = document.createElement('div');
    indicator.className = 'ai-tier-indicator';
    indicator.innerHTML = `
        <div class="tier-header">AI Swarm Status</div>
        <div class="tier-grid">
            <div class="tier-item" data-tier="sentinel">
                <div class="tier-dot" style="background: ${aiTierColors['gemini-2.0-flash-exp']}"></div>
                <span>Sentinel</span>
                <span class="tier-status">Active</span>
            </div>
            <div class="tier-item" data-tier="strategist">
                <div class="tier-dot" style="background: ${aiTierColors['gemini-1.5-pro']}"></div>
                <span>Strategist</span>
                <span class="tier-status">Standby</span>
            </div>
            <div class="tier-item" data-tier="arbiter">
                <div class="tier-dot" style="background: ${aiTierColors['anthropic/claude-3.5-sonnet']}"></div>
                <span>Arbiter</span>
                <span class="tier-status">Standby</span>
            </div>
            <div class="tier-item" data-tier="reasoning">
                <div class="tier-dot" style="background: ${aiTierColors['openai/gpt-5-mini']}"></div>
                <span>Reasoning</span>
                <span class="tier-status">Standby</span>
            </div>
            <div class="tier-item" data-tier="oracle">
                <div class="tier-dot" style="background: ${aiTierColors['devin']}"></div>
                <span>Oracle</span>
                <span class="tier-status">Standby</span>
            </div>
        </div>
    `;
    return indicator;
}

// Update AI tier status when model is used
function updateAITierStatus(model, status) {
    const tierMap = {
        'gemini-2.0-flash-exp': 'sentinel',
        'gemini-1.5-pro': 'strategist',
        'anthropic/claude-3.5-sonnet': 'arbiter',
        'openai/gpt-5-mini': 'reasoning',
        'devin': 'oracle'
    };

    const tierName = tierMap[model];
    if (!tierName) return;

    const tierItem = document.querySelector(`[data-tier="${tierName}"]`);
    if (!tierItem) return;

    const statusEl = tierItem.querySelector('.tier-status');
    const dotEl = tierItem.querySelector('.tier-dot');

    statusEl.textContent = status;

    if (status === 'Active') {
        dotEl.style.boxShadow = `0 0 20px ${aiTierColors[model]}`;
        dotEl.style.animation = 'pulse 2s infinite';
    } else {
        dotEl.style.boxShadow = 'none';
        dotEl.style.animation = 'none';
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // Add AI action feed container
    const sidebar = document.querySelector('.sidebar');
    if (sidebar) {
        const feedContainer = document.createElement('div');
        feedContainer.id = 'ai-action-feed';
        feedContainer.className = 'ai-action-feed';
        sidebar.appendChild(feedContainer);
    }

    // Add AI tier indicator
    const content = document.querySelector('.content');
    if (content) {
        const indicator = createAITierIndicator();
        content.insertBefore(indicator, content.firstChild);
    }

    // Mock AI actions for demo (remove in production)
    setTimeout(() => {
        showAIAction({
            model: 'gemini-2.0-flash-exp',
            resource: 'i-abc123',
            decision: 'Rightsize to t3.medium',
            reasoning: 'CPU usage consistently below 30% for 7 days',
            risk: 2.5,
            savings: 45.50
        });
        updateAITierStatus('gemini-2.0-flash-exp', 'Active');
    }, 2000);
});
