document.addEventListener('DOMContentLoaded', () => {
    initChart();
    populateActivity();
    populateResources();
    initSSE();
    startOODALoop();
    initModal();

    const runBtn = document.getElementById('run-optimization');
    if (runBtn) {
        runBtn.addEventListener('click', () => {
            runBtn.innerHTML = '<i class="ph ph-circle-notch animate-spin"></i> OPTIMIZING...';
            runBtn.style.opacity = '0.7';

            setTimeout(() => {
                showNotification('Guardian Cycle Triggered', 'The OODA loop has started an autonomous cycle.', 'info');
                runBtn.innerHTML = '<i class="ph ph-lightning"></i> RUN GUARDIAN';
                runBtn.style.opacity = '1';
                addActivityItem({ agent: 'architect', icon: 'ph ph-lightning', msg: 'Manual override: Immediate optimization cycle started', time: 'Now', glow: true }, true);
            }, 1000);
        });
    }

    // Start Simulate Feed if no SSE connection establishes
    setTimeout(() => {
        const feed = document.getElementById('activity-feed');
        if (feed && feed.children.length <= 1) {
            console.log("Starting simulation mode...");
            simulateActivityFeed();
        }
    }, 3000);
});

function initChart() {
    const ctx = document.getElementById('optimizationChart');
    if (!ctx) return;

    const gradient = ctx.getContext('2d').createLinearGradient(0, 0, 0, 400);
    gradient.addColorStop(0, 'rgba(56, 189, 248, 0.4)');
    gradient.addColorStop(1, 'rgba(56, 189, 248, 0)');

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4', 'Week 5', 'Week 6'],
            datasets: [{
                label: 'Costs Saved ($)',
                data: [5000, 12000, 19000, 28000, 36000, 46080],
                borderColor: '#38bdf8',
                borderWidth: 3,
                fill: true,
                backgroundColor: gradient,
                tension: 0.4,
                pointBackgroundColor: '#38bdf8',
                pointRadius: 4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false }
            },
            scales: {
                y: { grid: { color: 'rgba(255, 255, 255, 0.05)' }, ticks: { color: '#94a3b8' } },
                x: { grid: { display: false }, ticks: { color: '#94a3b8' } }
            }
        }
    });
}

function initSSE() {
    const source = new EventSource('/stream');
    source.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            const agent = (data.agent || 'architect').toLowerCase();
            const msg = data.metadata ? `${data.action}: ${data.metadata}` : data.action;
            const time = data.timestamp ? new Date(data.timestamp).toLocaleTimeString() : 'Just now';

            addActivityItem({
                agent: agent,
                icon: getIconForAgent(agent),
                msg: msg,
                time: time
            }, true);

            if (data.status === 'COMPLETED' || data.status === 'APPROVED') {
                showNotification('Swarm Event', msg, 'success');
            } else {
                showNotification('Swarm Signal', msg, 'info');
            }
        } catch (e) {
            console.error("Failed to parse SSE data", e);
        }
    };
    source.onerror = () => {
        // Silent fail, simulation will pick up
    };
}

function getIconForAgent(agent) {
    switch (agent) {
        case 'architect': return 'ph ph-sketch-logo';
        case 'auditor': return 'ph ph-eye';
        case 'builder': return 'ph ph-hammer';
        default: return 'ph ph-info';
    }
}

function initModal() {
    const overlay = document.getElementById('modal-overlay');
    const closeBtns = document.querySelectorAll('.close-modal');
    const confirmBtn = document.getElementById('confirm-approval');

    if (closeBtns) closeBtns.forEach(btn => {
        btn.addEventListener('click', () => overlay.classList.remove('active'));
    });

    if (confirmBtn) confirmBtn.addEventListener('click', () => {
        const resourceId = document.getElementById('modal-resource').innerText;
        overlay.classList.remove('active');
        showNotification('Task Approved', `Relocation of ${resourceId} initiated.`, 'info');
    });
}

function showModal(resource) {
    document.getElementById('modal-resource').innerText = resource.id;
    document.getElementById('modal-change').innerText = `${resource.current} → ${resource.proposed}`;
    document.getElementById('modal-impact').innerText = resource.savings + '/mo';
    document.getElementById('modal-overlay').classList.add('active');
}

function showNotification(title, message, type) {
    const container = document.getElementById('notification-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    const icon = type === 'success' ? 'ph ph-check-circle' : 'ph ph-info';

    toast.innerHTML = `
        <i class="${icon}"></i>
        <div class="toast-content">
            <strong style="display:block; font-size:0.9rem;">${title}</strong>
            <span style="font-size:0.8rem; color:var(--text-dim);">${message}</span>
        </div>
    `;

    container.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(20px)';
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

const activityData = [
    { agent: 'architect', icon: 'ph ph-eye', msg: 'Guardian initialized. Entering OODA loop...', time: '1m ago' }
];

function populateActivity() {
    const feed = document.getElementById('activity-feed');
    if (!feed) return;
    feed.innerHTML = '';
    activityData.forEach(item => addActivityItem(item));
}

function addActivityItem(item, prepend = false) {
    const feed = document.getElementById('activity-feed');
    if (!feed) return;

    const div = document.createElement('div');
    div.className = 'feed-item';
    const glowClass = item.glow ? 'agent-glow' : '';
    div.innerHTML = `
        <i class="${item.icon} ${glowClass}" style="color: ${getAgentColor(item.agent)}"></i>
        <div class="feed-content">
            <p>${item.msg}</p>
            <span class="feed-time">${item.time}</span>
        </div>
    `;
    if (prepend) {
        feed.prepend(div);
    } else {
        feed.appendChild(div);
    }
}

function getAgentColor(agent) {
    switch (agent) {
        case 'architect': return '#7000ff';
        case 'auditor': return '#ff00c8';
        case 'builder': return '#00f2ff';
        default: return '#fff';
    }
}

const resources = [
    { id: 'db-prod-01', current: 'db.m5.xlarge', proposed: 'db.m5.large', savings: '$450.00', risk: 'Low' },
    { id: 'web-srv-04', current: 't3.medium', proposed: 't3.nano', savings: '$25.00', risk: 'High' },
    { id: 'cache-redis', current: 'cache.m5.large', proposed: 'cache.t3.small', savings: '$120.00', risk: 'Med' }
];

function populateResources() {
    const body = document.getElementById('resources-body');
    if (!body) return;
    body.innerHTML = '';
    resources.forEach(res => {
        const row = document.createElement('tr');

        if (res.risk === 'High') {
            row.classList.add('vulcan-pulse');
        }

        row.innerHTML = `
            <td><strong>${res.id}</strong></td>
            <td>${res.current}</td>
            <td>${res.proposed}</td>
            <td class="positive" style="color:var(--accent-success);">${res.savings}</td>
            <td><span class="badge-risk ${res.risk.toLowerCase().substring(0, 3)}">${res.risk}</span></td>
            <td><button class="btn-action" onclick="handleApprove('${res.id}')">APPROVE</button></td>
        `;
        body.appendChild(row);
    });
}

window.handleApprove = function (id) {
    const res = resources.find(r => r.id === id);
    if (res) showModal(res);
}

function startOODALoop() {
    const steps = ['step-observe', 'step-orient', 'step-decide', 'step-act'];
    let current = 0;

    // Check if elements exist first
    if (!document.getElementById(steps[0])) return;

    // Initial activation
    document.getElementById(steps[0]).classList.add('active');

    setInterval(() => {
        // Reset all
        steps.forEach(id => {
            const el = document.getElementById(id);
            if (el) el.classList.remove('active');
        });

        // Activate current
        const currEl = document.getElementById(steps[current]);
        if (currEl) currEl.classList.add('active');

        // Next
        current = (current + 1) % steps.length;
    }, 1500);
}

function simulateActivityFeed() {
    const events = [
        { agent: 'sentinel', msg: 'Detected sustained CPU spike on resource i-0f9a8b7c' },
        { agent: 'strategist', msg: 'Analyzing cost patterns for potential spot instance migration' },
        { agent: 'auditor', msg: 'Compliance check passed for new security group rules' },
        { agent: 'builder', msg: 'Auto-scaling group resized to optimize for current load' },
        { agent: 'architect', msg: 'OODA Loop Cycle 42 completed. Optimization score: 94%' },
        { agent: 'sentinel', msg: 'Scanning RDS snapshots for retention policy adherence' }
    ];

    setInterval(() => {
        const evt = events[Math.floor(Math.random() * events.length)];
        addActivityItem({
            agent: evt.agent,
            icon: getIconForAgent(evt.agent),
            msg: evt.msg,
            time: 'Just now',
            glow: true
        }, true);

        const feed = document.getElementById('activity-feed');
        if (feed && feed.children.length > 20) {
            feed.removeChild(feed.lastChild);
        }
    }, 4500);
}
