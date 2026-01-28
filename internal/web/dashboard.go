package web

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SocksBalance Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #0f0f1e 0%, #1a1a2e 100%);
            color: #eee;
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        header {
            text-align: center;
            margin-bottom: 40px;
            padding: 30px 20px;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 16px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        h1 {
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 10px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .subtitle {
            color: #aaa;
            font-size: 1rem;
            margin-bottom: 20px;
        }

        .stats-summary {
            display: flex;
            gap: 20px;
            justify-content: center;
            flex-wrap: wrap;
            margin-top: 20px;
        }

        .stat-card {
            background: rgba(255, 255, 255, 0.08);
            padding: 15px 30px;
            border-radius: 12px;
            min-width: 150px;
            text-align: center;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 5px;
        }

        .stat-label {
            color: #aaa;
            font-size: 0.9rem;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .stat-card.total .stat-value { color: #667eea; }
        .stat-card.healthy .stat-value { color: #48bb78; }
        .stat-card.unhealthy .stat-value { color: #f56565; }

        .card {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 16px;
            padding: 30px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .loading {
            text-align: center;
            padding: 60px;
            color: #aaa;
            font-size: 1.2rem;
        }

        .loading::after {
            content: '...';
            animation: dots 1.5s steps(4, end) infinite;
        }

        @keyframes dots {
            0%, 20% { content: '.'; }
            40% { content: '..'; }
            60%, 100% { content: '...'; }
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }

        thead {
            background: rgba(255, 255, 255, 0.08);
        }

        th {
            padding: 16px;
            text-align: left;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 1px;
            color: #aaa;
            border-bottom: 2px solid rgba(255, 255, 255, 0.1);
        }

        td {
            padding: 16px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }

        tbody tr {
            transition: background 0.2s ease;
        }

        tbody tr:hover {
            background: rgba(255, 255, 255, 0.05);
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.85rem;
            font-weight: 600;
        }

        .status-badge.healthy {
            background: rgba(72, 187, 120, 0.2);
            color: #48bb78;
            border: 1px solid rgba(72, 187, 120, 0.4);
        }

        .status-badge.unhealthy {
            background: rgba(245, 101, 101, 0.2);
            color: #f56565;
            border: 1px solid rgba(245, 101, 101, 0.4);
        }

        .latency {
            font-weight: 600;
            font-size: 1rem;
        }

        .latency.fast { color: #48bb78; }
        .latency.medium { color: #ecc94b; }
        .latency.slow { color: #f56565; }
        .latency.unknown { color: #aaa; }

        .backend-name {
            font-weight: 500;
            color: #fff;
        }

        .backend-address {
            color: #aaa;
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
        }

        .last-updated {
            text-align: center;
            margin-top: 20px;
            color: #666;
            font-size: 0.9rem;
        }

        .no-data {
            text-align: center;
            padding: 60px;
            color: #aaa;
            font-size: 1.1rem;
        }

        @media (max-width: 768px) {
            h1 { font-size: 1.8rem; }
            .card { padding: 20px; }
            table { font-size: 0.9rem; }
            th, td { padding: 12px 8px; }
            .stats-summary { gap: 10px; }
            .stat-card { min-width: 120px; padding: 12px 20px; }
        }

        @media (max-width: 480px) {
            body { padding: 10px; }
            h1 { font-size: 1.5rem; }
            .stat-value { font-size: 1.5rem; }
            table { font-size: 0.85rem; }
            th, td { padding: 10px 6px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🚀 SocksBalance Dashboard</h1>
            <p class="subtitle">Real-time SOCKS5 Backend Monitoring</p>
            <div class="stats-summary">
                <div class="stat-card total">
                    <div class="stat-value" id="total-backends">-</div>
                    <div class="stat-label">Total Backends</div>
                </div>
                <div class="stat-card healthy">
                    <div class="stat-value" id="healthy-backends">-</div>
                    <div class="stat-label">Healthy</div>
                </div>
                <div class="stat-card unhealthy">
                    <div class="stat-value" id="unhealthy-backends">-</div>
                    <div class="stat-label">Unhealthy</div>
                </div>
            </div>
        </header>

        <div class="card">
            <div id="content">
                <div class="loading">Loading backend status</div>
            </div>
        </div>

        <div class="last-updated" id="last-updated"></div>
    </div>

    <script>
        // Placeholder for AJAX - will be implemented in STEP4
        function updateDashboard() {
            document.getElementById('content').innerHTML = `
                <div class="no-data">
                    <p>AJAX auto-update will be implemented in STEP4</p>
                    <p style="margin-top: 10px;">Try <a href="/api/stats" style="color: #667eea;">viewing the API directly</a></p>
                </div>
            `;
        }

        // Initial load
        setTimeout(updateDashboard, 1000);
    </script>
</body>
</html>
`
