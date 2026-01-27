// Wails runtime
const runtime = window.go?.main?.App;

// DOM Elements
const elements = {
    username: document.getElementById('username'),
    password: document.getElementById('password'),
    concurrent: document.getElementById('concurrent'),
    serversFile: document.getElementById('serversFile'),
    commandsFile: document.getElementById('commandsFile'),
    runBtn: document.getElementById('runBtn'),
    stopBtn: document.getElementById('stopBtn'),
    serverCount: document.getElementById('serverCount'),
    commandCount: document.getElementById('commandCount'),
    serversPreview: document.getElementById('serversPreview'),
    commandsPreview: document.getElementById('commandsPreview'),
    progressSection: document.getElementById('progressSection'),
    progressFill: document.getElementById('progressFill'),
    progressText: document.getElementById('progressText'),
    currentServer: document.getElementById('currentServer'),
    resultsSection: document.getElementById('resultsSection'),
    resultsBody: document.getElementById('resultsBody'),
    summary: document.getElementById('summary'),
    logSection: document.getElementById('logSection'),
    logTitle: document.getElementById('logTitle'),
    logContent: document.getElementById('logContent'),
    statusText: document.getElementById('statusText'),
    // Live Logs elements
    autoScroll: document.getElementById('autoScroll'),
    serverFilter: document.getElementById('serverFilter'),
    combinedLogContent: document.getElementById('combinedLogContent'),
    logsCombinedView: document.getElementById('logsCombinedView'),
    logsSplitView: document.getElementById('logsSplitView')
};

// State
let overlay = null;
let liveLogs = [];           // All logs
let serverLogs = {};         // Logs by server { 'ip': [{...}, ...] }
let logsViewMode = 'combined'; // 'combined' or 'split'
let knownServers = new Set(); // Track servers for filter dropdown

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
});

// Setup Wails event listeners
function setupEventListeners() {
    if (window.runtime) {
        window.runtime.EventsOn('progress', handleProgress);
        window.runtime.EventsOn('result', handleResult);
        window.runtime.EventsOn('completed', handleCompleted);
        window.runtime.EventsOn('error', handleError);
        window.runtime.EventsOn('log', handleLog);
    }
}

// Tab switching
function switchTab(tabName) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.tab === tabName);
    });
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.toggle('active', content.id === `tab-${tabName}`);
    });
}

// Live Logs handling
function handleLog(data) {
    const { serverIP, hostname, line } = data;
    const timestamp = Date.now();

    const logEntry = {
        serverIP,
        hostname,
        line,
        timestamp,
        formattedTime: new Date(timestamp).toLocaleTimeString()
    };

    // Store log
    liveLogs.push(logEntry);

    // Store by server
    if (!serverLogs[serverIP]) {
        serverLogs[serverIP] = [];
        addServerToFilter(serverIP, hostname);
        if (logsViewMode === 'split') {
            createServerLogPanel(serverIP, hostname);
        }
    }
    serverLogs[serverIP].push(logEntry);

    // Update UI
    updateLogsUI(logEntry);

    // Limit memory (keep last 10000 entries)
    if (liveLogs.length > 10000) {
        liveLogs.shift();
    }
}

function updateLogsUI(logEntry) {
    const currentFilter = elements.serverFilter?.value || 'all';

    if (logsViewMode === 'combined') {
        if (currentFilter === 'all' || currentFilter === logEntry.serverIP) {
            appendToCombinedLog(logEntry);
        }
    } else {
        appendToSplitLog(logEntry);
    }
}

function appendToCombinedLog(logEntry) {
    const logContent = elements.combinedLogContent;
    if (!logContent) return;

    const logLine = document.createElement('div');
    logLine.className = 'log-line';
    logLine.dataset.serverIp = logEntry.serverIP;
    logLine.innerHTML = `<span class="log-time">[${logEntry.formattedTime}]</span> ` +
                        `<span class="log-server">[${escapeHtml(logEntry.hostname)}]</span> ` +
                        `<span class="log-text">${escapeHtml(logEntry.line)}</span>`;
    logContent.appendChild(logLine);

    // Auto-scroll
    if (elements.autoScroll?.checked) {
        logContent.scrollTop = logContent.scrollHeight;
    }

    // Limit DOM nodes
    while (logContent.children.length > 5000) {
        logContent.removeChild(logContent.firstChild);
    }
}

function appendToSplitLog(logEntry) {
    const panelId = `log-panel-${logEntry.serverIP.replace(/\./g, '-')}`;
    const panel = document.getElementById(panelId);
    if (!panel) return;

    const content = panel.querySelector('.panel-content');
    if (!content) return;

    const logLine = document.createElement('div');
    logLine.className = 'log-line';
    logLine.innerHTML = `<span class="log-time">[${logEntry.formattedTime}]</span> ` +
                        `<span class="log-text">${escapeHtml(logEntry.line)}</span>`;
    content.appendChild(logLine);

    // Auto-scroll
    if (elements.autoScroll?.checked) {
        content.scrollTop = content.scrollHeight;
    }

    // Limit DOM nodes
    while (content.children.length > 2000) {
        content.removeChild(content.firstChild);
    }
}

function addServerToFilter(serverIP, hostname) {
    if (knownServers.has(serverIP)) return;
    knownServers.add(serverIP);

    const option = document.createElement('option');
    option.value = serverIP;
    option.textContent = `${hostname} (${serverIP})`;
    elements.serverFilter?.appendChild(option);
}

function createServerLogPanel(serverIP, hostname) {
    const splitView = elements.logsSplitView;
    if (!splitView) return;

    const panelId = `log-panel-${serverIP.replace(/\./g, '-')}`;
    if (document.getElementById(panelId)) return;

    const panel = document.createElement('div');
    panel.className = 'server-log-panel';
    panel.id = panelId;
    panel.innerHTML = `
        <div class="panel-header">
            <span>${escapeHtml(hostname)} (${escapeHtml(serverIP)})</span>
            <button onclick="closeServerPanel('${escapeHtml(serverIP)}')">X</button>
        </div>
        <pre class="panel-content"></pre>
    `;
    splitView.appendChild(panel);

    // Populate with existing logs
    const logs = serverLogs[serverIP] || [];
    const content = panel.querySelector('.panel-content');
    logs.forEach(log => {
        const logLine = document.createElement('div');
        logLine.className = 'log-line';
        logLine.innerHTML = `<span class="log-time">[${log.formattedTime}]</span> ` +
                            `<span class="log-text">${escapeHtml(log.line)}</span>`;
        content.appendChild(logLine);
    });
}

function closeServerPanel(serverIP) {
    const panelId = `log-panel-${serverIP.replace(/\./g, '-')}`;
    const panel = document.getElementById(panelId);
    if (panel) {
        panel.remove();
    }
}

function setLogsView(mode) {
    logsViewMode = mode;

    // Update toggle buttons
    document.querySelectorAll('.logs-view-toggle button').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.view === mode);
    });

    if (mode === 'combined') {
        elements.logsCombinedView.style.display = 'block';
        elements.logsSplitView.style.display = 'none';
    } else {
        elements.logsCombinedView.style.display = 'none';
        elements.logsSplitView.style.display = 'grid';

        // Create panels for all known servers
        knownServers.forEach(serverIP => {
            const logs = serverLogs[serverIP];
            if (logs && logs.length > 0) {
                createServerLogPanel(serverIP, logs[0].hostname);
            }
        });
    }
}

function filterLogs() {
    const filter = elements.serverFilter?.value || 'all';
    const logContent = elements.combinedLogContent;
    if (!logContent) return;

    // Show/hide log lines based on filter
    Array.from(logContent.children).forEach(line => {
        if (filter === 'all' || line.dataset.serverIp === filter) {
            line.style.display = '';
        } else {
            line.style.display = 'none';
        }
    });
}

function clearLiveLogs() {
    liveLogs = [];
    serverLogs = {};
    knownServers.clear();

    if (elements.combinedLogContent) {
        elements.combinedLogContent.innerHTML = '';
    }
    if (elements.logsSplitView) {
        elements.logsSplitView.innerHTML = '';
    }
    if (elements.serverFilter) {
        elements.serverFilter.innerHTML = '<option value="all">All Servers</option>';
    }
}

// File selection
async function selectServersFile() {
    try {
        const file = await runtime.SelectServersFile();
        if (file) {
            elements.serversFile.value = file;
            await previewServers();
        }
    } catch (err) {
        showError('Failed to select file: ' + err);
    }
}

async function selectCommandsFile() {
    try {
        const file = await runtime.SelectCommandsFile();
        if (file) {
            elements.commandsFile.value = file;
            await previewCommands();
        }
    } catch (err) {
        showError('Failed to select file: ' + err);
    }
}

// Preview functions
async function previewServers() {
    try {
        const servers = await runtime.PreviewServers();
        if (servers && servers.length > 0) {
            elements.serverCount.textContent = servers.length;
            elements.serversPreview.innerHTML = '<ul>' +
                servers.map(s => `<li><strong>${s.hostname}</strong> (${s.ip})</li>`).join('') +
                '</ul>';
        } else {
            elements.serverCount.textContent = '0';
            elements.serversPreview.innerHTML = '<p class="placeholder">No servers loaded</p>';
        }
    } catch (err) {
        showError('Failed to preview servers: ' + err);
    }
}

async function previewCommands() {
    try {
        const commands = await runtime.PreviewCommands();
        if (commands && commands.length > 0) {
            elements.commandCount.textContent = commands.length;
            elements.commandsPreview.innerHTML = '<ul>' +
                commands.map(c => `<li><code>${escapeHtml(c)}</code></li>`).join('') +
                '</ul>';
        } else {
            elements.commandCount.textContent = '0';
            elements.commandsPreview.innerHTML = '<p class="placeholder">No commands loaded</p>';
        }
    } catch (err) {
        showError('Failed to preview commands: ' + err);
    }
}

// Execution control
async function startExecution() {
    const username = elements.username.value.trim();
    const password = elements.password.value;
    const concurrent = parseInt(elements.concurrent.value) || 5;

    if (!username || !password) {
        showError('Please enter username and password');
        return;
    }

    if (!elements.serversFile.value) {
        showError('Please select a servers file');
        return;
    }

    if (!elements.commandsFile.value) {
        showError('Please select a commands file');
        return;
    }

    // Clear previous logs
    clearLiveLogs();

    try {
        const success = await runtime.StartExecution(username, password, concurrent);
        if (success) {
            setRunningState(true);
            elements.resultsBody.innerHTML = '';
            elements.progressSection.style.display = 'block';
            elements.resultsSection.style.display = 'block';
            elements.progressFill.style.width = '0%';
            elements.progressText.textContent = '0 / 0';
            elements.currentServer.textContent = '';
            elements.summary.innerHTML = '';
            setStatus(`Running (${concurrent} parallel)...`);
        }
    } catch (err) {
        showError('Failed to start: ' + err);
    }
}

async function stopExecution() {
    try {
        await runtime.StopExecution();
        setRunningState(false);
        setStatus('Stopped');
    } catch (err) {
        showError('Failed to stop: ' + err);
    }
}

// Event handlers
function handleProgress(data) {
    const { current, total, hostname, ip, status } = data;
    const percent = (current / total) * 100;

    elements.progressFill.style.width = percent + '%';
    elements.progressText.textContent = `${current} / ${total}`;

    if (status === 'connecting') {
        elements.currentServer.textContent = `Connecting to ${hostname} (${ip})...`;
    } else if (status === 'success') {
        elements.currentServer.textContent = `${hostname}: Success`;
    } else if (status === 'failed') {
        elements.currentServer.textContent = `${hostname}: Failed`;
    }
}

function handleResult(data) {
    const { hostname, ip, success, error, logPath, duration } = data;

    const row = document.createElement('tr');
    row.innerHTML = `
        <td>${escapeHtml(hostname)}</td>
        <td>${escapeHtml(ip)}</td>
        <td class="${success ? 'status-success' : 'status-failed'}">${success ? 'Success' : 'Failed'}</td>
        <td>${(duration / 1000).toFixed(1)}s</td>
        <td>
            ${success && logPath ? `<button onclick="viewLog('${escapeHtml(logPath)}', '${escapeHtml(hostname)}')">View Log</button>` :
              error ? `<span title="${escapeHtml(error)}">Error</span>` : '-'}
        </td>
    `;
    elements.resultsBody.appendChild(row);
}

function handleCompleted(data) {
    const { success, fail, total, logDir } = data;

    setRunningState(false);

    elements.summary.innerHTML = `
        <span class="success">Success: ${success}</span> |
        <span class="fail">Failed: ${fail}</span> |
        Total: ${total} |
        Logs: ${logDir}
    `;

    setStatus(`Completed: ${success} success, ${fail} failed`);
}

function handleError(message) {
    showError(message);
}

// Log viewer
async function viewLog(path, hostname) {
    try {
        const content = await runtime.ReadLogFile(path);
        elements.logTitle.textContent = hostname + '.log';
        elements.logContent.textContent = content;

        // Show overlay
        overlay = document.createElement('div');
        overlay.className = 'overlay';
        overlay.onclick = closeLogViewer;
        document.body.appendChild(overlay);

        elements.logSection.style.display = 'flex';
    } catch (err) {
        showError('Failed to read log: ' + err);
    }
}

function closeLogViewer() {
    elements.logSection.style.display = 'none';
    if (overlay) {
        overlay.remove();
        overlay = null;
    }
}

// Open logs folder
async function openLogsFolder() {
    try {
        await runtime.OpenLogsFolder();
    } catch (err) {
        showError('Failed to open logs folder: ' + err);
    }
}

// UI helpers
function setRunningState(running) {
    elements.runBtn.disabled = running;
    elements.stopBtn.disabled = !running;
    elements.username.disabled = running;
    elements.password.disabled = running;
    elements.concurrent.disabled = running;
}

function setStatus(text) {
    elements.statusText.textContent = text;
}

function showError(message) {
    alert(message);
    setStatus('Error: ' + message);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Expose functions to window for onclick handlers
window.selectServersFile = selectServersFile;
window.selectCommandsFile = selectCommandsFile;
window.startExecution = startExecution;
window.stopExecution = stopExecution;
window.viewLog = viewLog;
window.closeLogViewer = closeLogViewer;
window.openLogsFolder = openLogsFolder;
window.switchTab = switchTab;
window.setLogsView = setLogsView;
window.filterLogs = filterLogs;
window.clearLiveLogs = clearLiveLogs;
window.closeServerPanel = closeServerPanel;
