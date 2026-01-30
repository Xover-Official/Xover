// ROI Chart, Risk Heatmap, and Action Log JavaScript
// Append this to main.js

// ============================================
// ROI CHART
// ============================================
let roiChart = null;
let roiData = {
    labels: [],
    savings: [],
    costs: []
};

function initROIChart() {
    const canvas = document.getElementById('roiChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    roiChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: roiData.labels,
            datasets: [
                {
                    label: 'Total Savings ($)',
                    data: roiData.savings,
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    tension: 0.4,
                    fill: true
                },
                {
                    label: 'AI Costs ($)',
                    data: roiData.costs,
                    borderColor: '#ef4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.1)',
                    tension: 0.4,
                    fill: true
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    labels: { color: '#fff' }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: { color: '#fff' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                },
                x: {
                    ticks: { color: '#fff' },
                    grid: { color: 'rgba(255, 255, 255, 0.1)' }
                }
            }
        }
    });
}

async function updateROIChart() {
    try {
        const response = await fetch('/api/roi');
        const data = await response.json();

        if (data.error) return;

        const now = new Date().toLocaleTimeString();
        roiData.labels.push(now);
        roiData.savings.push(data.total_savings_usd || 0);
        roiData.costs.push(data.total_cost_usd || 0);

        if (roiData.labels.length > 20) {
            roiData.labels.shift();
            roiData.savings.shift();
            roiData.costs.shift();
        }

        if (roiChart) roiChart.update();
    } catch (err) {
        console.error('ROI update failed:', err);
    }
}

// ============================================
// RISK HEATMAP
// ============================================
function updateRiskHeatmap(resources) {
    const heatmap = document.getElementById('riskHeatmap');
    if (!heatmap) return;

    if (!resources || resources.length === 0) {
        heatmap.innerHTML = '<p style="opacity: 0.5;">No resources scanned</p>';
        return;
    }

    heatmap.innerHTML = resources.map(r => {
        const risk = r.risk || 0;
        const riskClass = risk < 3 ? 'risk-low' : risk < 7 ? 'risk-medium' : 'risk-high';
        return `
            <div class="heatmap-cell ${riskClass}">
                <div class="resource-id">${r.id || 'Unknown'}</div>
                <div class="risk-score">${risk.toFixed(1)}</div>
                <div style="font-size: 0.7rem; opacity: 0.8;">Risk</div>
            </div>
        `;
    }).join('');
}

// ============================================
// ACTION LOG
// ============================================
const actionLogData = [];

function addActionToLog(action) {
    actionLogData.unshift(action);
    if (actionLogData.length > 50) actionLogData.pop();
    renderActionLog();
}

function renderActionLog() {
    const tbody = document.getElementById('actionLogBody');
    if (!tbody) return;

    if (actionLogData.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align: center; opacity: 0.5;">Waiting...</td></tr>';
        return;
    }

    tbody.innerHTML = actionLogData.map(a => `
        <tr>
            <td>${a.time}</td>
            <td><code>${a.resource}</code></td>
            <td>${a.action}</td>
            <td>${a.risk.toFixed(1)}</td>
            <td style="color: #10b981;">$${a.savings.toFixed(2)}</td>
            <td><span class="status-badge status-${a.status}">${a.status.toUpperCase()}</span></td>
        </tr>
    `).join('');
}

// Initialize on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        initROIChart();
        setInterval(updateROIChart, 10000);
        updateROIChart();
    });
} else {
    initROIChart();
    setInterval(updateROIChart, 10000);
    updateROIChart();
}
