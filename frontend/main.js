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
    combinedLogContent: document.getElementById('combinedLogContent')
};

// State
let overlay = null;
let liveLogs = [];           // All logs
let serverLogs = {};         // Logs by server { 'ip': [{...}, ...] }
let currentServerTab = 'all'; // Current selected server tab
let knownServers = new Set(); // Track servers for tabs

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
        addServerTab(serverIP, hostname);
    }
    serverLogs[serverIP].push(logEntry);

    // Update UI if this server's tab is active or "All" tab
    if (currentServerTab === 'all' || currentServerTab === serverIP) {
        appendToLogView(logEntry);
    }

    // Limit memory (keep last 10000 entries)
    if (liveLogs.length > 10000) {
        liveLogs.shift();
    }
}

function appendToLogView(logEntry) {
    const logContent = elements.combinedLogContent;
    if (!logContent) return;

    const logLine = document.createElement('div');
    logLine.className = 'log-line';
    logLine.dataset.serverIp = logEntry.serverIP;

    // Show server name only in "All" tab
    if (currentServerTab === 'all') {
        logLine.innerHTML = `<span class="log-time">[${logEntry.formattedTime}]</span> ` +
                            `<span class="log-server">[${escapeHtml(logEntry.hostname)}]</span> ` +
                            `<span class="log-text">${escapeHtml(logEntry.line)}</span>`;
    } else {
        logLine.innerHTML = `<span class="log-time">[${logEntry.formattedTime}]</span> ` +
                            `<span class="log-text">${escapeHtml(logEntry.line)}</span>`;
    }
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

function addServerTab(serverIP, hostname) {
    if (knownServers.has(serverIP)) return;
    knownServers.add(serverIP);

    const tabsContainer = document.getElementById('serverTabs');
    if (!tabsContainer) return;

    const tab = document.createElement('button');
    tab.className = 'server-tab';
    tab.dataset.server = serverIP;
    tab.textContent = hostname;
    tab.onclick = () => switchServerTab(serverIP);
    tabsContainer.appendChild(tab);
}

function switchServerTab(serverIP) {
    currentServerTab = serverIP;

    // Update tab buttons
    document.querySelectorAll('.server-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.server === serverIP);
    });

    // Clear and repopulate log content
    const logContent = elements.combinedLogContent;
    if (!logContent) return;
    logContent.innerHTML = '';

    // Get logs to display
    let logsToShow = [];
    if (serverIP === 'all') {
        logsToShow = liveLogs;
    } else {
        logsToShow = serverLogs[serverIP] || [];
    }

    // Populate logs
    logsToShow.forEach(log => {
        appendToLogView(log);
    });
}

function clearLiveLogs() {
    liveLogs = [];
    serverLogs = {};
    knownServers.clear();
    currentServerTab = 'all';

    if (elements.combinedLogContent) {
        elements.combinedLogContent.innerHTML = '';
    }

    const tabsContainer = document.getElementById('serverTabs');
    if (tabsContainer) {
        tabsContainer.innerHTML = '<button class="server-tab active" data-server="all" onclick="switchServerTab(\'all\')">All</button>';
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
window.switchServerTab = switchServerTab;
window.clearLiveLogs = clearLiveLogs;
