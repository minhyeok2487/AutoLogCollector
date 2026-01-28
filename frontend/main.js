// Wails runtime
const runtime = window.go?.main?.App;

// DOM Elements
const elements = {
    username: document.getElementById('username'),
    password: document.getElementById('password'),
    timeout: document.getElementById('timeout'),
    enableMode: document.getElementById('enableMode'),
    disablePaging: document.getElementById('disablePaging'),
    enablePasswordOptions: document.getElementById('enablePasswordOptions'),
    samePassword: document.getElementById('samePassword'),
    enablePassword: document.getElementById('enablePassword'),
    serversBody: document.getElementById('serversBody'),
    commandsInput: document.getElementById('commandsInput'),
    runBtn: document.getElementById('runBtn'),
    stopBtn: document.getElementById('stopBtn'),
    serverCount: document.getElementById('serverCount'),
    commandCount: document.getElementById('commandCount'),
    progressSection: document.getElementById('progressSection'),
    progressFill: document.getElementById('progressFill'),
    progressText: document.getElementById('progressText'),
    currentServer: document.getElementById('currentServer'),
    resultsSection: document.getElementById('resultsSection'),
    resultsBody: document.getElementById('resultsBody'),
    summary: document.getElementById('summary'),
    logViewerModal: document.getElementById('logViewerModal'),
    logTitle: document.getElementById('logTitle'),
    logContent: document.getElementById('logContent'),
    statusText: document.getElementById('statusText'),
    statusDot: document.getElementById('statusDot'),
    autoScroll: document.getElementById('autoScroll'),
    combinedLogContent: document.getElementById('combinedLogContent'),
    sectionTitle: document.getElementById('sectionTitle'),
    connectionInfo: document.getElementById('connectionInfo')
};

// State
let liveLogs = [];
let serverLogs = {};
let currentServerTab = null;  // null means show all logs
let knownServers = new Set();
let latestUpdateInfo = null;
let currentSection = 'execution';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    loadVersion();
    setupInputListeners();
    addServerRow(); // Add one empty row by default
});

// Setup Wails event listeners
function setupEventListeners() {
    if (window.runtime) {
        window.runtime.EventsOn('progress', handleProgress);
        window.runtime.EventsOn('result', handleResult);
        window.runtime.EventsOn('completed', handleCompleted);
        window.runtime.EventsOn('error', handleError);
        window.runtime.EventsOn('log', handleLog);
        window.runtime.EventsOn('updateProgress', handleUpdateProgress);
        window.runtime.EventsOn('updateError', handleUpdateError);
        window.runtime.EventsOn('updateComplete', handleUpdateComplete);
    }
}

// Setup input listeners for counting
function setupInputListeners() {
    if (elements.commandsInput) {
        elements.commandsInput.addEventListener('input', updateCommandCount);
    }
}

// Load and display version
async function loadVersion() {
    try {
        const version = await runtime.GetCurrentVersion();
        document.getElementById('versionInfo').textContent = `v${version}`;
        const versionStatus = document.getElementById('versionStatus');
        if (versionStatus) {
            versionStatus.textContent = `v${version}`;
        }
    } catch (err) {
        console.error('Failed to load version:', err);
    }
}

// ==================== Navigation ====================

function showSection(section) {
    currentSection = section;

    // Update nav items
    document.querySelectorAll('.nav-item[data-section]').forEach(item => {
        item.classList.toggle('active', item.dataset.section === section);
    });

    // Update section title
    const titles = {
        execution: 'Execution',
        results: 'Results',
        logs: 'Live Logs'
    };
    elements.sectionTitle.textContent = titles[section] || section;

    // Show/hide sections
    document.getElementById('executionSection').style.display = section === 'execution' ? 'flex' : 'none';
    document.getElementById('resultsSection').style.display = section === 'results' ? 'flex' : 'none';
    document.getElementById('logsSection').style.display = section === 'logs' ? 'flex' : 'none';
}

// ==================== Server Table Management ====================

function addServerRow(ip = '', hostname = '') {
    const tbody = elements.serversBody;
    if (!tbody) return;

    const row = document.createElement('tr');
    row.innerHTML = `
        <td><input type="text" placeholder="192.168.1.1" value="${escapeHtml(ip)}" onchange="updateServerCount()"></td>
        <td><input type="text" placeholder="Router1" value="${escapeHtml(hostname)}" onchange="updateServerCount()"></td>
        <td><button type="button" class="delete-btn" onclick="removeServerRow(this)">&times;</button></td>
    `;
    tbody.appendChild(row);
    updateServerCount();

    // Focus on the IP input of the new row
    const ipInput = row.querySelector('input');
    if (ipInput && !ip) {
        ipInput.focus();
    }
}

function removeServerRow(btn) {
    const row = btn.closest('tr');
    if (row) {
        row.remove();
        updateServerCount();
    }
}

function updateServerCount() {
    const servers = getServersFromTable();
    elements.serverCount.textContent = servers.length;
}

function getServersFromTable() {
    const rows = elements.serversBody?.querySelectorAll('tr') || [];
    const servers = [];

    rows.forEach(row => {
        const inputs = row.querySelectorAll('input');
        if (inputs.length >= 2) {
            const ip = inputs[0].value.trim();
            const hostname = inputs[1].value.trim();
            if (ip) {
                servers.push({ ip, hostname: hostname || ip });
            }
        }
    });

    return servers;
}

function clearServersTable() {
    if (elements.serversBody) {
        elements.serversBody.innerHTML = '';
    }
    updateServerCount();
}

// Import CSV file
async function importCSV() {
    try {
        const servers = await runtime.ImportServersFromCSV();
        if (servers && servers.length > 0) {
            clearServersTable();
            servers.forEach(server => {
                addServerRow(server.ip, server.hostname);
            });
        }
    } catch (err) {
        showError('Failed to import CSV: ' + err);
    }
}

// ==================== Commands Management ====================

function getCommandsFromTextarea() {
    const text = elements.commandsInput?.value || '';
    return text.split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0);
}

function updateCommandCount() {
    const commands = getCommandsFromTextarea();
    elements.commandCount.textContent = commands.length;
}

// ==================== Settings Menu ====================

function toggleSettingsMenu() {
    const menu = document.getElementById('settingsMenu');
    menu.classList.toggle('show');
}

function closeSettingsMenu() {
    const menu = document.getElementById('settingsMenu');
    menu.classList.remove('show');
}

document.addEventListener('click', (e) => {
    if (!e.target.closest('.sidebar-footer')) {
        closeSettingsMenu();
    }
});

// ==================== Enable Mode Options ====================

function toggleEnablePassword() {
    const enableMode = elements.enableMode?.checked;
    if (elements.enablePasswordOptions) {
        elements.enablePasswordOptions.style.display = enableMode ? 'flex' : 'none';
    }
}

function toggleEnablePasswordInput() {
    const samePassword = elements.samePassword?.checked;
    if (elements.enablePassword) {
        elements.enablePassword.style.display = samePassword ? 'none' : 'block';
        if (samePassword) {
            elements.enablePassword.value = '';
        }
    }
}

// ==================== Update Functions ====================

async function checkForUpdates() {
    try {
        const info = await runtime.CheckForUpdates();
        if (!info) return;

        if (info.available) {
            latestUpdateInfo = info;
            showUpdateModal(info);
        } else {
            alert('You are using the latest version.');
        }
    } catch (err) {
        showError('Failed to check for updates: ' + err);
    }
}

function showUpdateModal(info) {
    document.getElementById('updateVersionInfo').innerHTML =
        `<strong>Current:</strong> ${info.currentVersion}<br>` +
        `<strong>Latest:</strong> ${info.latestVersion}`;

    document.getElementById('updateReleaseNotes').textContent =
        info.releaseNotes || 'No release notes available.';

    document.getElementById('updateModal').style.display = 'flex';
    document.getElementById('updateProgress').style.display = 'none';
    document.getElementById('downloadUpdateBtn').style.display = 'block';
    document.getElementById('restartBtn').style.display = 'none';
}

function closeUpdateModal() {
    document.getElementById('updateModal').style.display = 'none';
}

async function downloadUpdate() {
    if (!latestUpdateInfo || !latestUpdateInfo.downloadURL) {
        showError('No download URL available');
        return;
    }

    document.getElementById('downloadUpdateBtn').style.display = 'none';
    document.getElementById('updateProgress').style.display = 'block';
    document.getElementById('updateProgressText').textContent = 'Starting download...';

    try {
        const success = await runtime.DownloadAndInstallUpdate(latestUpdateInfo.downloadURL);
        if (!success) {
            document.getElementById('downloadUpdateBtn').style.display = 'block';
            document.getElementById('updateProgress').style.display = 'none';
        }
    } catch (err) {
        showError('Update failed: ' + err);
        document.getElementById('downloadUpdateBtn').style.display = 'block';
        document.getElementById('updateProgress').style.display = 'none';
    }
}

function handleUpdateProgress(data) {
    const { downloaded, total, percent } = data;
    document.getElementById('updateProgressFill').style.width = percent + '%';

    const downloadedMB = (downloaded / 1024 / 1024).toFixed(1);
    const totalMB = (total / 1024 / 1024).toFixed(1);
    document.getElementById('updateProgressText').textContent =
        `Downloading... ${downloadedMB} MB / ${totalMB} MB (${percent.toFixed(0)}%)`;
}

function handleUpdateError(message) {
    showError(message);
    document.getElementById('downloadUpdateBtn').style.display = 'block';
    document.getElementById('updateProgress').style.display = 'none';
}

function handleUpdateComplete(message) {
    document.getElementById('updateProgressText').textContent = 'Update installed successfully!';
    document.getElementById('updateProgressFill').style.width = '100%';
    document.getElementById('restartBtn').style.display = 'block';
    alert(message);
}

async function restartApp() {
    try {
        await runtime.RestartApp();
    } catch (err) {
        showError('Failed to restart: ' + err);
    }
}

// ==================== Live Logs ====================

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

    liveLogs.push(logEntry);

    if (!serverLogs[serverIP]) {
        serverLogs[serverIP] = [];
        addServerTab(serverIP, hostname);
    }
    serverLogs[serverIP].push(logEntry);

    // Show log if no filter or matches current filter
    if (!currentServerTab || currentServerTab === serverIP) {
        appendToLogView(logEntry);
    }

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

    logLine.innerHTML = `<span class="log-time">[${logEntry.formattedTime}]</span> ` +
                        `<span class="log-server">[${escapeHtml(logEntry.hostname)}]</span> ` +
                        `<span class="log-text">${escapeHtml(logEntry.line)}</span>`;
    logContent.appendChild(logLine);

    // Auto-scroll to bottom
    if (elements.autoScroll?.checked) {
        requestAnimationFrame(() => {
            logContent.scrollTop = logContent.scrollHeight;
        });
    }

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

    document.querySelectorAll('.server-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.server === serverIP);
    });

    const logContent = elements.combinedLogContent;
    if (!logContent) return;
    logContent.innerHTML = '';

    let logsToShow = serverLogs[serverIP] || [];
    logsToShow.forEach(log => appendToLogView(log));
}

function clearLiveLogs() {
    liveLogs = [];
    serverLogs = {};
    knownServers.clear();
    currentServerTab = null;

    if (elements.combinedLogContent) {
        elements.combinedLogContent.innerHTML = '';
    }

    const tabsContainer = document.getElementById('serverTabs');
    if (tabsContainer) {
        tabsContainer.innerHTML = '';
    }
}

// ==================== Execution Control ====================

async function startExecution() {
    const username = elements.username.value.trim();
    const password = elements.password.value;
    const timeout = parseInt(elements.timeout.value) || 1;
    const enableMode = elements.enableMode?.checked ?? false;
    const disablePaging = elements.disablePaging?.checked ?? true;

    // Get enable password (use login password if "same" is checked)
    let enablePwd = '';
    if (enableMode) {
        const sameAsLogin = elements.samePassword?.checked ?? true;
        enablePwd = sameAsLogin ? password : (elements.enablePassword?.value || '');
    }

    if (!username || !password) {
        showError('Please enter username and password');
        return;
    }

    const servers = getServersFromTable();
    if (servers.length === 0) {
        showError('Please add at least one server');
        return;
    }

    const commands = getCommandsFromTextarea();
    if (commands.length === 0) {
        showError('Please enter at least one command');
        return;
    }

    clearLiveLogs();

    try {
        // Send servers and commands to backend
        await runtime.SetServers(servers);
        await runtime.SetCommands(commands);

        const success = await runtime.StartExecution(username, password, timeout, enableMode, disablePaging, enablePwd);
        if (success) {
            setRunningState(true);
            elements.resultsBody.innerHTML = '';
            elements.progressSection.style.display = 'block';
            elements.progressFill.style.width = '0%';
            elements.progressText.textContent = '0 / 0';
            elements.currentServer.textContent = '';
            elements.summary.innerHTML = '';
            setStatus('Running...');
            updateConnectionInfo(`Executing on ${servers.length} servers`);
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
        updateConnectionInfo('Execution stopped');
    } catch (err) {
        showError('Failed to stop: ' + err);
    }
}

// ==================== Event Handlers ====================

function handleProgress(data) {
    const { current, total, hostname, ip, status } = data;
    const percent = (current / total) * 100;

    elements.progressFill.style.width = percent + '%';
    elements.progressText.textContent = `${current} / ${total}`;

    if (status === 'connecting') {
        elements.currentServer.textContent = `Connecting to ${hostname} (${ip})...`;
        updateConnectionInfo(`Connecting to ${hostname}`);
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
            ${success && logPath ? `<button class="btn-secondary" onclick="viewLog('${escapeHtml(logPath)}', '${escapeHtml(hostname)}')">View</button>` :
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
    updateConnectionInfo('No active connections');

    // Auto switch to results section
    showSection('results');
}

function handleError(message) {
    showError(message);
}

// ==================== Log Viewer ====================

async function viewLog(path, hostname) {
    try {
        const content = await runtime.ReadLogFile(path);
        elements.logTitle.textContent = hostname + '.log';
        elements.logContent.textContent = content;
        elements.logViewerModal.style.display = 'flex';
    } catch (err) {
        showError('Failed to read log: ' + err);
    }
}

function closeLogViewer() {
    elements.logViewerModal.style.display = 'none';
}

async function openLogsFolder() {
    try {
        await runtime.OpenLogsFolder();
    } catch (err) {
        showError('Failed to open logs folder: ' + err);
    }
}

async function exportResults() {
    try {
        const path = await runtime.ExportResults();
        if (path) {
            alert('Excel exported to:\n' + path);
        } else {
            showError('No results to export');
        }
    } catch (err) {
        showError('Failed to export Excel: ' + err);
    }
}

// ==================== UI Helpers ====================

function setRunningState(running) {
    elements.runBtn.disabled = running;
    elements.stopBtn.disabled = !running;
    elements.username.disabled = running;
    elements.password.disabled = running;
    elements.timeout.disabled = running;
    if (elements.enableMode) elements.enableMode.disabled = running;
    if (elements.disablePaging) elements.disablePaging.disabled = running;
    if (elements.samePassword) elements.samePassword.disabled = running;
    if (elements.enablePassword) elements.enablePassword.disabled = running;

    // Update status dot
    if (elements.statusDot) {
        elements.statusDot.className = 'status-dot' + (running ? ' running' : '');
    }

    // Disable server table inputs
    const serverInputs = elements.serversBody?.querySelectorAll('input, button') || [];
    serverInputs.forEach(el => el.disabled = running);

    // Disable commands textarea
    if (elements.commandsInput) {
        elements.commandsInput.disabled = running;
    }
}

function setStatus(text) {
    elements.statusText.textContent = text;
}

function updateConnectionInfo(text) {
    if (elements.connectionInfo) {
        elements.connectionInfo.textContent = text;
    }
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

// ==================== Expose Functions ====================

window.showSection = showSection;
window.addServerRow = addServerRow;
window.removeServerRow = removeServerRow;
window.importCSV = importCSV;
window.updateServerCount = updateServerCount;
window.startExecution = startExecution;
window.stopExecution = stopExecution;
window.viewLog = viewLog;
window.closeLogViewer = closeLogViewer;
window.openLogsFolder = openLogsFolder;
window.switchServerTab = switchServerTab;
window.clearLiveLogs = clearLiveLogs;
window.checkForUpdates = checkForUpdates;
window.closeUpdateModal = closeUpdateModal;
window.downloadUpdate = downloadUpdate;
window.restartApp = restartApp;
window.toggleSettingsMenu = toggleSettingsMenu;
window.closeSettingsMenu = closeSettingsMenu;
window.exportResults = exportResults;
window.toggleEnablePassword = toggleEnablePassword;
window.toggleEnablePasswordInput = toggleEnablePasswordInput;
