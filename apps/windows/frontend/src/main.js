import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {GetGroups, GetSnippets, ExpandSnippet, PreviewSnippet, CreateSampleData, GetVaultInfo, GetHotkeys, GetHotkeyStatuses, UpdateHotkeys, ValidateHotkey, RefreshHotkeyStatuses, GetDefaultHotkeys, GetPauseState, TogglePause, GetExpansionStats, TestExpansion, GetExpansionBuffer, ClearExpansionBuffer, SetExpansionEnabled, IsExpansionEnabled} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

document.querySelector('#app').innerHTML = `
    <div class="container">
        <header class="header">
            <img id="logo" class="logo" style="width: 32px; height: 32px;">
            <h1>SnipQ - Snippet Manager</h1>
            <nav class="nav-tabs">
                <button class="nav-tab active" data-tab="main">Main</button>
                <button class="nav-tab" data-tab="hotkeys">Hotkeys</button>
            </nav>
        </header>
        
        <div id="main-tab" class="tab-content active">
            <div class="main-content">
                <div class="sidebar">
                    <h3>Groups</h3>
                    <div id="vault-info" class="vault-info"></div>
                    <div id="groups-list" class="groups-list"></div>
                    
                    <div class="expansion-status">
                        <h4>Auto-Expansion</h4>
                        <div id="expansion-info" class="expansion-info">
                            <div class="status-item">
                                <span class="label">Status:</span>
                                <span id="expansion-status" class="status-value">Loading...</span>
                            </div>
                            <div class="status-item">
                                <span class="label">Buffer:</span>
                                <code id="expansion-buffer" class="buffer-display">-</code>
                            </div>
                            <div class="expansion-controls">
                                <button class="btn small" id="toggle-expansion">Toggle</button>
                                <button class="btn small" id="clear-buffer">Clear Buffer</button>
                                <button class="btn small" id="refresh-expansion">Refresh</button>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="content">
                    <div class="test-section">
                        <h3>Test Snippet Expansion</h3>
                        <div class="input-box">
                            <input class="input" id="trigger-input" type="text" placeholder="Enter trigger (e.g., :ty, :hello)" autocomplete="off" />
                            <button class="btn primary" id="expand-btn">🚀 Expand</button>
                            <button class="btn secondary" id="preview-btn">👁️ Preview</button>
                            <button class="btn-clear" id="clear-btn" title="Clear field">✕</button>
                        </div>
                        <div class="result" id="result">Enter a trigger above to test expansion</div>
                    </div>
                    
                    <div class="snippets-section">
                        <h3>Snippets</h3>
                        <div id="snippets-list" class="snippets-list"></div>
                    </div>
                </div>
            </div>
        </div>
        
        <div id="hotkeys-tab" class="tab-content">
            <div class="hotkeys-page">
                <div class="hotkeys-header">
                    <h2>Hotkey Settings</h2>
                    <div class="hotkey-status">
                        <div class="pause-control">
                            <span id="pause-status" class="pause-indicator">Active</span>
                            <button class="btn pause-toggle" id="pause-toggle-btn">
                                <span id="pause-btn-text">Pause</span>
                            </button>
                        </div>
                        <button class="btn refresh-btn" id="refresh-hotkeys">Refresh</button>
                    </div>
                </div>
                
                <div class="hotkeys-table-container">
                    <table class="hotkeys-table">
                        <thead>
                            <tr>
                                <th>Action</th>
                                <th>Current Shortcut</th>
                                <th>OS Registration Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="hotkeys-table-body">
                        </tbody>
                    </table>
                </div>
                
                <div class="hotkeys-actions">
                    <button class="btn success" id="save-hotkeys">Save Changes</button>
                    <button class="btn secondary" id="revert-hotkeys">Revert</button>
                    <button class="btn warning" id="reset-defaults">Reset to Defaults</button>
                </div>
            </div>
        </div>
    </div>
`;
document.getElementById('logo').src = logo;

let triggerInput = document.getElementById("trigger-input");
let resultElement = document.getElementById("result");
let groupsList = document.getElementById("groups-list");
let snippetsList = document.getElementById("snippets-list");
let vaultInfo = document.getElementById("vault-info");
let currentSelectedGroup = null;

// Navigation functionality
let currentTab = 'main';
const navTabs = document.querySelectorAll('.nav-tab');
const tabContents = document.querySelectorAll('.tab-content');

// Hotkey management state
let currentHotkeys = {};
let pendingHotkeys = {};
let hotkeyStatuses = {};
let isRecording = false;
let recordingAction = null;

// Switch tab function
function switchTab(tabName) {
    console.log(`Switching to tab: ${tabName}`);
    
    // Update nav buttons
    navTabs.forEach(tab => {
        tab.classList.remove('active');
        if (tab.dataset.tab === tabName) {
            tab.classList.add('active');
        }
    });
    
    // Update tab content
    tabContents.forEach(content => {
        content.classList.remove('active');
        if (content.id === `${tabName}-tab`) {
            content.classList.add('active');
        }
    });
    
    currentTab = tabName;
    
    // Load data for the active tab
    if (tabName === 'hotkeys') {
        loadHotkeysPage();
    }
}

// Add navigation event listeners
navTabs.forEach(tab => {
    tab.addEventListener('click', () => {
        switchTab(tab.dataset.tab);
    });
});

// Focus on trigger input
if (triggerInput) {
    triggerInput.focus();
}

// Add Enter key support
if (triggerInput) {
    triggerInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            testExpansion();
        }
    });
}

// Load groups on startup
loadGroups();
checkVaultInfo();
updateExpansionStatus();

// Set up periodic refresh of expansion status
setInterval(updateExpansionStatus, 2000);

// Listen for hotkey events from the backend
EventsOn("hotkey:openPalette", function() {
    console.log("Hotkey event: Open Palette");
    // Switch to main tab and focus input
    switchTab('main');
    if (triggerInput) {
        triggerInput.focus();
        triggerInput.select();
    }
});

EventsOn("hotkey:pauseToggled", function(data) {
    console.log("Hotkey event: Pause toggled", data);
    updatePauseStatus(data.paused);
});

EventsOn("hotkey:expandNow", function() {
    console.log("Hotkey event: Expand Now");
    if (triggerInput && triggerInput.value) {
        testExpansion();
    }
});

// Listen for focus-input events from the tray
EventsOn("focus-input", function() {
    console.log("Focus-input event received from tray");
    if (triggerInput) {
        triggerInput.focus();
        triggerInput.select(); // Select any existing text
    }
});

// Hotkey Management Functions

async function loadHotkeysPage() {
    console.log("Loading hotkeys page");
    
    try {
        // Load current hotkeys and statuses
        const [hotkeys, statuses, pauseState] = await Promise.all([
            GetHotkeys(),
            GetHotkeyStatuses(),
            GetPauseState()
        ]);
        
        console.log("Loaded hotkeys data:", { hotkeys, statuses, pauseState });
        
        currentHotkeys = hotkeys || {};
        pendingHotkeys = { ...currentHotkeys };
        hotkeyStatuses = statuses || {};
        
        renderHotkeysTable();
        updatePauseStatus(pauseState);
        
    } catch (error) {
        console.error("Error loading hotkeys page:", error);
    }
}

function renderHotkeysTable() {
    const tbody = document.getElementById('hotkeys-table-body');
    if (!tbody) return;
    
    console.log("Rendering hotkeys table");
    
    const actions = {
        'openPalette': {
            name: 'Open Palette',
            description: 'Opens the snippet search palette'
        },
        'togglePause': {
            name: 'Toggle Pause',
            description: 'Pauses/resumes snippet expansion'
        },
        'expandNow': {
            name: 'Expand Now',
            description: 'Expands the current trigger immediately'
        }
    };
    
    tbody.innerHTML = '';
    
    Object.entries(actions).forEach(([actionKey, actionInfo]) => {
        const hotkey = pendingHotkeys[actionKey] || { modifiers: [], key: '', enabled: false };
        const status = hotkeyStatuses[actionKey] || { status: 'Not Registered', error: '' };
        
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>
                <div class="action-name">${actionInfo.name}</div>
                <div class="action-description">${actionInfo.description}</div>
            </td>
            <td>
                <div class="shortcut-display" id="shortcut-${actionKey}">
                    ${formatHotkey(hotkey)}
                </div>
            </td>
            <td>
                <span class="status-badge status-${status.status.toLowerCase().replace(' ', '-')}" id="status-${actionKey}">
                    ${status.status}
                </span>
                ${status.error ? `<div class="status-error">${status.error}</div>` : ''}
            </td>
            <td>
                <div class="hotkey-actions">
                    <label class="enabled-toggle">
                        <input type="checkbox" ${hotkey.enabled ? 'checked' : ''} 
                               onchange="toggleHotkeyEnabled('${actionKey}', this.checked)">
                        <span>Enabled</span>
                    </label>
                    <div class="hotkey-buttons">
                        <button class="btn small change" onclick="changeHotkey('${actionKey}')">Change</button>
                        <button class="btn small reset" onclick="resetHotkey('${actionKey}')">Reset</button>
                    </div>
                </div>
            </td>
        `;
        tbody.appendChild(row);
    });
}

function formatHotkey(hotkey) {
    if (!hotkey.enabled || !hotkey.key) {
        return '<em>Not set</em>';
    }
    
    const modifiers = hotkey.modifiers || [];
    const parts = [...modifiers, hotkey.key];
    return parts.join(' + ');
}

function updatePauseStatus(paused) {
    const pauseStatus = document.getElementById('pause-status');
    const pauseToggleBtn = document.getElementById('pause-toggle-btn');
    const pauseBtnText = document.getElementById('pause-btn-text');
    
    if (pauseStatus) {
        pauseStatus.textContent = paused ? 'Paused' : 'Active';
        pauseStatus.className = `pause-indicator ${paused ? 'paused' : ''}`;
    }
    
    if (pauseToggleBtn) {
        pauseToggleBtn.className = `btn pause-toggle ${paused ? 'paused' : ''}`;
    }
    
    if (pauseBtnText) {
        pauseBtnText.textContent = paused ? 'Resume' : 'Pause';
    }
}

// Global function for pause toggle (called from onclick)
window.togglePause = async function() {
    console.log("Toggling pause state");
    try {
        const newState = await TogglePause();
        console.log("Pause toggled to:", newState);
        updatePauseStatus(newState);
    } catch (error) {
        console.error("Error toggling pause:", error);
    }
};

// Global functions for hotkey management (called from onclick)
window.toggleHotkeyEnabled = function(action, enabled) {
    console.log(`Toggling ${action} enabled: ${enabled}`);
    if (!pendingHotkeys[action]) {
        pendingHotkeys[action] = { modifiers: [], key: '', enabled: false };
    }
    pendingHotkeys[action].enabled = enabled;
    renderHotkeysTable();
};

window.changeHotkey = function(action) {
    console.log(`Changing hotkey for ${action}`);
    showHotkeyRecordingModal(action);
};

window.resetHotkey = async function(action) {
    console.log(`Resetting hotkey for ${action}`);
    try {
        const defaults = await GetDefaultHotkeys();
        if (defaults[action]) {
            pendingHotkeys[action] = { ...defaults[action] };
            renderHotkeysTable();
        }
    } catch (error) {
        console.error("Error resetting hotkey:", error);
    }
};

async function refreshHotkeyStatuses() {
    console.log("Refreshing hotkey statuses");
    try {
        const statuses = await RefreshHotkeyStatuses();
        hotkeyStatuses = statuses || {};
        renderHotkeysTable();
    } catch (error) {
        console.error("Error refreshing hotkey statuses:", error);
    }
}

async function saveHotkeys() {
    console.log("Saving hotkeys", pendingHotkeys);
    try {
        await UpdateHotkeys(pendingHotkeys);
        currentHotkeys = { ...pendingHotkeys };
        
        // Refresh statuses after saving
        await refreshHotkeyStatuses();
        
        alert("Hotkeys saved successfully!");
    } catch (error) {
        console.error("Error saving hotkeys:", error);
        alert("Error saving hotkeys: " + error);
    }
}

function revertHotkeys() {
    console.log("Reverting hotkeys");
    pendingHotkeys = { ...currentHotkeys };
    renderHotkeysTable();
}

async function resetToDefaults() {
    console.log("Resetting to default hotkeys");
    if (confirm("Reset all hotkeys to defaults? This will overwrite your current settings.")) {
        try {
            const defaults = await GetDefaultHotkeys();
            pendingHotkeys = { ...defaults };
            renderHotkeysTable();
        } catch (error) {
            console.error("Error getting default hotkeys:", error);
        }
    }
}

// Hotkey Recording Modal
function showHotkeyRecordingModal(action) {
    console.log(`Showing recording modal for ${action}`);
    
    // Get the friendly action name
    const actions = {
        'openPalette': 'Open Palette',
        'togglePause': 'Toggle Pause',
        'expandNow': 'Expand Now'
    };
    const actionName = actions[action] || action;
    
    const modal = document.createElement('div');
    modal.className = 'hotkey-modal';
    modal.innerHTML = `
        <div class="hotkey-modal-content" style="position: relative;">
            <button id="close-modal-x" style="position: absolute; top: 16px; right: 16px; background: none; border: none; font-size: 1.7rem; color: #495057; cursor: pointer; z-index: 10;" title="Close">&times;</button>
            <h3>Record Hotkey for ${actionName}</h3>
            <p>Press your desired key combination. Use Ctrl, Alt, Shift, or Win + another key.</p>
            <div class="recording-display listening" id="recording-display">
                Press your hotkey combination...
            </div>
            <div class="recording-error" id="recording-error" style="display: none;"></div>
            <div class="modal-actions">
                <button class="btn success" id="save-recording" style="display: none;">Save</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    const recordingDisplay = modal.querySelector('#recording-display');
    const recordingError = modal.querySelector('#recording-error');
    const saveBtn = modal.querySelector('#save-recording');
    const closeXBtn = modal.querySelector('#close-modal-x');
    
    let recordedHotkey = null;
    
    // Key event handler
    function handleKeyDown(e) {
        e.preventDefault();
        
        // Don't record single modifier keys
        if (['Control', 'Alt', 'Shift', 'Meta'].includes(e.key)) {
            return;
        }
        
        const modifiers = [];
        if (e.ctrlKey) modifiers.push('Ctrl');
        if (e.altKey) modifiers.push('Alt');
        if (e.shiftKey) modifiers.push('Shift');
        if (e.metaKey) modifiers.push('Win');
        
        // Require at least one modifier
        if (modifiers.length === 0) {
            showRecordingError("You must use at least one modifier key (Ctrl, Alt, Shift, or Win)");
            return;
        }
        
        const key = normalizeKey(e.key);
        const hotkey = {
            modifiers: modifiers,
            key: key,
            enabled: true
        };
        
        console.log("Recorded hotkey:", hotkey);
        
        // Validate the hotkey
        ValidateHotkey(hotkey)
            .then(() => {
                recordedHotkey = hotkey;
                recordingDisplay.textContent = formatHotkey(hotkey);
                recordingDisplay.classList.remove('listening');
                saveBtn.style.display = 'inline-block';
                hideRecordingError();
            })
            .catch(error => {
                console.error("Hotkey validation failed:", error);
                showRecordingError(error.toString());
            });
    }
    
    function showRecordingError(message) {
        recordingError.textContent = message;
        recordingError.style.display = 'block';
    }
    
    function hideRecordingError() {
        recordingError.style.display = 'none';
    }
    
    function normalizeKey(key) {
        // Normalize common key names
        const keyMap = {
            ' ': 'Space',
            '.': 'Period',
            ',': 'Comma',
            'Enter': 'Enter'
        };
        
        return keyMap[key] || key.toUpperCase();
    }
    
    // Event listeners
    document.addEventListener('keydown', handleKeyDown);
    
    saveBtn.addEventListener('click', () => {
        if (recordedHotkey) {
            pendingHotkeys[action] = recordedHotkey;
            renderHotkeysTable();
        }
        closeModal();
    });
    
    closeXBtn.addEventListener('click', closeModal);
    modal.addEventListener('click', (e) => {
        if (e.target === modal) closeModal();
    });
    
    function closeModal() {
        document.removeEventListener('keydown', handleKeyDown);
        document.body.removeChild(modal);
    }
    
    // Focus the modal for key capture
    modal.focus();
}

// Add event listeners for main tab buttons
document.addEventListener('DOMContentLoaded', () => {
    // Main tab buttons
    const expandBtn = document.getElementById('expand-btn');
    const previewBtn = document.getElementById('preview-btn');
    const clearBtn = document.getElementById('clear-btn');
    
    if (expandBtn) expandBtn.addEventListener('click', testExpansion);
    if (previewBtn) previewBtn.addEventListener('click', testPreview);
    if (clearBtn) clearBtn.addEventListener('click', clearField);
    
    // Hotkey page buttons
    const refreshBtn = document.getElementById('refresh-hotkeys');
    const saveBtn = document.getElementById('save-hotkeys');
    const revertBtn = document.getElementById('revert-hotkeys');
    const resetBtn = document.getElementById('reset-defaults');
    const pauseToggleBtn = document.getElementById('pause-toggle-btn');
    
    if (refreshBtn) refreshBtn.addEventListener('click', refreshHotkeyStatuses);
    if (saveBtn) saveBtn.addEventListener('click', saveHotkeys);
    if (revertBtn) revertBtn.addEventListener('click', revertHotkeys);
    if (resetBtn) resetBtn.addEventListener('click', resetToDefaults);
    if (pauseToggleBtn) pauseToggleBtn.addEventListener('click', togglePause);
});

// Clear field function
function clearField() {
    triggerInput.value = '';
    resultElement.innerHTML = 'Enter a trigger above to test expansion';
    triggerInput.focus();
}

// Test expansion function
function testExpansion() {
    let trigger = triggerInput.value;
    if (trigger === "") {
        resultElement.innerText = "Please enter a trigger";
        return;
    }

    console.log('Testing expansion for:', trigger);
    try {
        ExpandSnippet(trigger)
            .then((result) => {
                console.log('Expansion result:', result);
                resultElement.innerHTML = `
                    <div><strong>Expanded:</strong> ${result.output}</div>
                    <div><strong>Snippet:</strong> ${result.usedSnippet}</div>
                    <div><strong>Parameters:</strong></div>
                    <pre>${JSON.stringify(result.usedParams, null, 2)}</pre>
                `;
            })
            .catch((err) => {
                console.error('Expansion error:', err);
                resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Expansion exception:', err);
        resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

function testPreview() {
    let trigger = triggerInput.value;
    if (trigger === "") {
        resultElement.innerText = "Please enter a trigger";
        return;
    }

    console.log('Testing preview for:', trigger);
    try {
        PreviewSnippet(trigger)
            .then((result) => {
                console.log('Preview result:', result);
                resultElement.innerHTML = `<strong>Preview:</strong> ${result}`;
            })
            .catch((err) => {
                console.error('Preview error:', err);
                resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Preview exception:', err);
        resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

function loadGroups() {
    console.log('Loading groups...');
    try {
        GetGroups()
            .then((groups) => {
                console.log('Groups loaded:', groups);
                groupsList.innerHTML = '';
                
                if (!groups || groups.length === 0) {
                    groupsList.innerHTML = '<div class="error">No groups found. The vault might be empty.</div>';
                    return;
                }
                
                groups.forEach(group => {
                    const groupElement = document.createElement('div');
                    groupElement.className = 'group-item';
                    groupElement.innerHTML = `
                        <div class="group-header" onclick="loadSnippets('${group.id}')" data-group-id="${group.id}">
                            <span>${group.icon || '📁'} ${group.name}</span>
                            <span class="group-id">${group.id}</span>
                        </div>
                    `;
                    groupsList.appendChild(groupElement);
                });
            })
            .catch((err) => {
                console.error('Error loading groups:', err);
                groupsList.innerHTML = `<div class="error">Error loading groups: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception loading groups:', err);
        groupsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

window.loadSnippets = function(groupId) {
    console.log('Loading snippets for group:', groupId);
    
    // Update group selection state
    currentSelectedGroup = groupId;
    document.querySelectorAll('.group-header').forEach(header => {
        header.classList.remove('active');
    });
    document.querySelector(`[data-group-id="${groupId}"]`).classList.add('active');
    
    try {
        GetSnippets(groupId)
            .then((snippets) => {
                console.log('Snippets loaded:', snippets);
                snippetsList.innerHTML = `<h4>Snippets in ${groupId}</h4>`;
                
                if (!snippets || snippets.length === 0) {
                    snippetsList.innerHTML += '<div class="error">No snippets found in this group.</div>';
                    return;
                }
                
                snippets.forEach(snippet => {
                    const snippetElement = document.createElement('div');
                    snippetElement.className = 'snippet-item';
                    snippetElement.innerHTML = `
                        <div class="snippet-header">
                            <strong>${snippet.trigger}</strong> - ${snippet.name}
                        </div>
                        <div class="snippet-description">${snippet.description || ''}</div>
                        <div class="snippet-template"><pre>${snippet.template}</pre></div>
                        <div class="snippet-actions">
                            <button class="btn small" onclick="testTrigger('${snippet.trigger}')">Test This</button>
                        </div>
                    `;
                    snippetsList.appendChild(snippetElement);
                });
            })
            .catch((err) => {
                console.error('Error loading snippets:', err);
                snippetsList.innerHTML = `<div class="error">Error loading snippets: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception loading snippets:', err);
        snippetsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
};

// Helper function to test a specific trigger
window.testTrigger = function(trigger) {
    triggerInput.value = trigger;
    testExpansion();
    // Scroll to top to view the test section
    document.querySelector('.test-section').scrollIntoView({ 
        behavior: 'smooth', 
        block: 'start' 
    });
};

function checkVaultInfo() {
    try {
        GetVaultInfo()
            .then((info) => {
                console.log('Vault info:', info);
                vaultInfo.innerHTML = `
                    <div class="vault-info-content">
                        <small>Groups: ${info.groups} | Snippets: ${info.snippets}</small>
                    </div>
                `;
            })
            .catch((err) => {
                console.error('Error getting vault info:', err);
                vaultInfo.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception getting vault info:', err);
        vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
};

function createSampleData() {
    try {
        CreateSampleData()
            .then(() => {
                console.log('Sample data created');
                vaultInfo.innerHTML = '<div class="success">Sample data created!</div>';
                // Refresh the UI
                loadGroups();
                checkVaultInfo();
            })
            .catch((err) => {
                console.error('Error creating sample data:', err);
                vaultInfo.innerHTML = `<div class="error">Error creating sample data: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception creating sample data:', err);
        vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

// Add event listeners for all buttons when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Main tab buttons
    const expandBtn = document.getElementById('expand-btn');
    const previewBtn = document.getElementById('preview-btn');
    const clearBtn = document.getElementById('clear-btn');
    
    if (expandBtn) expandBtn.addEventListener('click', testExpansion);
    if (previewBtn) previewBtn.addEventListener('click', testPreview);
    if (clearBtn) clearBtn.addEventListener('click', clearField);
    
    // Hotkey page buttons
    const refreshBtn = document.getElementById('refresh-hotkeys');
    const saveBtn = document.getElementById('save-hotkeys');
    const revertBtn = document.getElementById('revert-hotkeys');
    const resetBtn = document.getElementById('reset-defaults');
    const pauseToggleBtn = document.getElementById('pause-toggle-btn');
    
    if (refreshBtn) refreshBtn.addEventListener('click', refreshHotkeyStatuses);
    if (saveBtn) saveBtn.addEventListener('click', saveHotkeys);
    if (revertBtn) revertBtn.addEventListener('click', revertHotkeys);
    if (resetBtn) resetBtn.addEventListener('click', resetToDefaults);
    if (pauseToggleBtn) pauseToggleBtn.addEventListener('click', togglePause);
    
    // Expansion control buttons
    const toggleExpansionBtn = document.getElementById('toggle-expansion');
    const clearBufferBtn = document.getElementById('clear-buffer');
    const refreshExpansionBtn = document.getElementById('refresh-expansion');
    
    if (toggleExpansionBtn) toggleExpansionBtn.addEventListener('click', toggleExpansion);
    if (clearBufferBtn) clearBufferBtn.addEventListener('click', clearExpansionBuffer);
    if (refreshExpansionBtn) refreshExpansionBtn.addEventListener('click', updateExpansionStatus);
});

// Expansion Status Functions
async function updateExpansionStatus() {
    try {
        const [stats, enabled] = await Promise.all([
            GetExpansionStats(),
            IsExpansionEnabled()
        ]);
        
        const statusElement = document.getElementById('expansion-status');
        const bufferElement = document.getElementById('expansion-buffer');
        
        if (statusElement) {
            statusElement.textContent = enabled ? 'Enabled' : 'Disabled';
            statusElement.className = `status-value ${enabled ? 'enabled' : 'disabled'}`;
        }
        
        if (bufferElement) {
            const buffer = stats.currentBuffer || '';
            bufferElement.textContent = buffer || '-';
            bufferElement.title = buffer; // Show full buffer on hover
        }
        
    } catch (error) {
        console.error('Error updating expansion status:', error);
    }
}

async function toggleExpansion() {
    try {
        const currentState = await IsExpansionEnabled();
        await SetExpansionEnabled(!currentState);
        updateExpansionStatus();
        console.log('Expansion toggled to:', !currentState);
    } catch (error) {
        console.error('Error toggling expansion:', error);
    }
}

async function clearExpansionBuffer() {
    try {
        await ClearExpansionBuffer();
        updateExpansionStatus();
        console.log('Expansion buffer cleared');
    } catch (error) {
        console.error('Error clearing expansion buffer:', error);
    }
}
