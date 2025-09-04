import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {GetGroups, GetSnippets, ExpandSnippet, PreviewSnippet, CreateSampleData, GetVaultInfo, GetHotkeys, GetHotkeyStatuses, UpdateHotkeys, ValidateHotkey, RefreshHotkeyStatuses, GetDefaultHotkeys, GetPauseState, TogglePause, GetExpansionStats, TestExpansion, GetExpansionBuffer, ClearExpansionBuffer, SetExpansionEnabled, IsExpansionEnabled, IsSuggestionsEnabled, SetSuggestionsEnabled, GetAppSettings, UpdateSettings, GetDefaultSettings, ExportSettings, ImportSettings} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

document.querySelector('#app').innerHTML = `
    <div class="container">
        <header class="header">
            <img id="logo" class="logo" style="width: 32px; height: 32px;">
            <h1>SnipQ - Snippet Manager</h1>
            <nav class="nav-tabs">
                <button class="nav-tab active" data-tab="main">Main</button>
                <button class="nav-tab" data-tab="settings">Settings</button>
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
        
        <div id="settings-tab" class="tab-content">
            <div class="hotkeys-page">
                <div class="hotkeys-header">
                    <h2>Application Settings</h2>
                    <button class="btn refresh-btn" id="refresh-settings">Refresh</button>
                </div>
                
                <!-- Hotkeys Section -->
                <div class="hotkeys-table-container">
                    <h3 data-section="hotkeys" style="margin: 0 0 1rem 0; padding: 1rem; background: #f8f9fa; color: #495057;">🎹 Hotkeys</h3>
                    <div style="padding: 1.5rem;">
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Expand Now:</label>
                                <div style="display: flex; gap: 0.5rem; align-items: center;">
                                    <input type="text" id="hotkey-expand-now" class="shortcut-display" readonly placeholder="Click to set" style="flex: 1; cursor: pointer;">
                                    <button class="btn secondary small" onclick="recordHotkeyForSetting('expandNow', 'hotkey-expand-now')">Change</button>
                                </div>
                            </div>
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Open Palette:</label>
                                <div style="display: flex; gap: 0.5rem; align-items: center;">
                                    <input type="text" id="hotkey-open-palette" class="shortcut-display" readonly placeholder="Click to set" style="flex: 1; cursor: pointer;">
                                    <button class="btn secondary small" onclick="recordHotkeyForSetting('openPalette', 'hotkey-open-palette')">Change</button>
                                </div>
                            </div>
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Toggle Pause:</label>
                                <div style="display: flex; gap: 0.5rem; align-items: center;">
                                    <input type="text" id="hotkey-toggle-pause" class="shortcut-display" readonly placeholder="Click to set" style="flex: 1; cursor: pointer;">
                                    <button class="btn secondary small" onclick="recordHotkeyForSetting('togglePause', 'hotkey-toggle-pause')">Change</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Text Expansion Section -->
                <div class="hotkeys-table-container">
                    <h3 data-section="expansion" style="margin: 0 0 1rem 0; padding: 1rem; background: #f8f9fa; color: #495057;">⚡ Text Expansion</h3>
                    <div style="padding: 1.5rem;">
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Trigger Prefix:</label>
                                <select id="trigger-prefix" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                    <option value=":">: (colon)</option>
                                    <option value=";">; (semicolon)</option>
                                    <option value="/">//</option>
                                    <option value="\\">\\ (backslash)</option>
                                    <option value=",">, (comma)</option>
                                </select>
                                <small style="color: #6c757d; font-size: 0.8rem;">Character that triggers snippet suggestions</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Expand Key:</label>
                                <select id="expand-key" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                    <option value="Tab">Tab</option>
                                    <option value="Enter">Enter</option>
                                    <option value="Space">Space</option>
                                </select>
                                <small style="color: #6c757d; font-size: 0.8rem;">Key used to trigger expansion of typed triggers</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Typing Timeout (ms):</label>
                                <div style="display: flex; align-items: center; gap: 1rem;">
                                    <input type="range" id="typing-timeout" min="500" max="10000" step="250" value="2000" style="flex: 1;">
                                    <input type="number" id="typing-timeout-value" min="500" max="10000" step="250" value="2000" style="width: 80px; padding: 0.25rem; border: 1px solid #ced4da; border-radius: 4px;">
                                    <label style="display: flex; align-items: center; gap: 0.25rem; cursor: pointer;">
                                        <input type="checkbox" id="no-timeout" style="cursor: pointer;">
                                        <span style="font-size: 0.8rem;">No timeout</span>
                                    </label>
                                </div>
                                <small style="color: #6c757d; font-size: 0.8rem;">How long to wait before clearing the typing buffer</small>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Smart Suggestions Section -->
                <div class="hotkeys-table-container">
                    <h3 data-section="suggestions" style="margin: 0 0 1rem 0; padding: 1rem; background: #f8f9fa; color: #495057;">💡 Smart Suggestions</h3>
                    <div style="padding: 1.5rem;">
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-weight: 500; color: #495057; font-size: 1rem;">
                                    <input type="checkbox" id="suggestions-enabled" checked style="cursor: pointer;">
                                    <span>Enable Smart Suggestions</span>
                                </label>
                                <small style="color: #6c757d; font-size: 0.8rem;">Show IDE-style autocomplete suggestions while typing</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Minimum Query Length:</label>
                                <input type="number" id="min-query-length" min="1" max="5" value="1" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                <small style="color: #6c757d; font-size: 0.8rem;">Minimum characters needed before showing suggestions</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Maximum Suggestions:</label>
                                <input type="number" id="max-suggestions" min="3" max="20" value="10" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                <small style="color: #6c757d; font-size: 0.8rem;">Maximum number of suggestions to display</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Show Delay (ms):</label>
                                <input type="number" id="suggestion-delay" min="0" max="1000" step="50" value="200" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                <small style="color: #6c757d; font-size: 0.8rem;">Delay before showing suggestions (prevents flickering)</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Hide Delay (ms):</label>
                                <input type="number" id="suggestion-hide-delay" min="500" max="5000" step="250" value="1000" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                <small style="color: #6c757d; font-size: 0.8rem;">How long to keep suggestions visible after stopping typing</small>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Advanced Settings Section -->
                <div class="hotkeys-table-container">
                    <h3 data-section="advanced" style="margin: 0 0 1rem 0; padding: 1rem; background: #f8f9fa; color: #495057;">🔧 Advanced Settings</h3>
                    <div style="padding: 1.5rem;">
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-weight: 500; color: #495057;">
                                    <input type="checkbox" id="strict-boundaries" checked style="cursor: pointer;">
                                    <span>Strict Word Boundaries</span>
                                </label>
                                <small style="color: #6c757d; font-size: 0.8rem;">Only expand triggers at word boundaries (recommended)</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-weight: 500; color: #495057;">
                                    <input type="checkbox" id="pause-on-failure" style="cursor: pointer;">
                                    <span>Pause on Expansion Failure</span>
                                </label>
                                <small style="color: #6c757d; font-size: 0.8rem;">Automatically pause when expansion fails</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="font-weight: 500; color: #495057;">Keyboard Buffer Size:</label>
                                <input type="number" id="buffer-size" min="50" max="500" value="100" style="padding: 0.5rem; border: 1px solid #ced4da; border-radius: 4px; background: white;">
                                <small style="color: #6c757d; font-size: 0.8rem;">Maximum characters to track in typing buffer</small>
                            </div>
                            
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-weight: 500; color: #495057;">
                                    <input type="checkbox" id="case-sensitive" style="cursor: pointer;">
                                    <span>Case Sensitive Triggers</span>
                                </label>
                                <small style="color: #6c757d; font-size: 0.8rem;">Whether trigger matching is case sensitive</small>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="hotkeys-actions">
                    <button class="btn success" id="save-settings">💾 Save All Settings</button>
                    <button class="btn secondary" id="reset-settings">🔄 Reset to Defaults</button>
                    <button class="btn warning" id="export-settings">📤 Export Settings</button>
                    <button class="btn secondary" id="import-settings">📥 Import Settings</button>
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

// Settings change tracking
let hasUnsavedChanges = false;
let originalSettings = {};
let sectionChanges = {
    hotkeys: false,
    expansion: false,
    suggestions: false,
    advanced: false
};

// Hotkey management state
let currentHotkeys = {};
let pendingHotkeys = {};
let hotkeyStatuses = {};
let isRecording = false;
let recordingAction = null;

// Switch tab function
function switchTab(tabName) {
    console.log(`Switching to tab: ${tabName}`);
    
    // Check for unsaved changes when switching to main tab
    if (tabName === 'main' && hasUnsavedChanges) {
        const confirmSwitch = confirm(
            'You have unsaved changes in Settings. Are you sure you want to switch tabs?\n\n' +
            'Your changes will be lost if you continue without saving.'
        );
        if (!confirmSwitch) {
            return; // Don't switch tabs
        }
        // Reset changes if user confirms
        hasUnsavedChanges = false;
        resetSectionChanges();
        updateSectionTitles();
        updateSaveButton();
    }
    
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
    if (tabName === 'settings') {
        loadSettingsPage();
    }
}

// Settings change management functions
function resetSectionChanges() {
    sectionChanges = {
        hotkeys: false,
        expansion: false,
        suggestions: false,
        advanced: false
    };
}

function updateSectionTitles() {
    const sections = [
        { id: 'hotkeys', title: '🎹 Hotkeys' },
        { id: 'expansion', title: '⚡ Text Expansion' },
        { id: 'suggestions', title: '💡 Smart Suggestions' },
        { id: 'advanced', title: '🔧 Advanced Settings' }
    ];
    
    sections.forEach(section => {
        const titleElement = document.querySelector(`[data-section="${section.id}"]`);
        if (titleElement) {
            const hasChanges = sectionChanges[section.id];
            titleElement.textContent = hasChanges ? `${section.title} (*)` : section.title;
        }
    });
}

function updateSaveButton() {
    const saveButton = document.getElementById('save-settings');
    if (saveButton) {
        saveButton.disabled = !hasUnsavedChanges;
        saveButton.style.opacity = hasUnsavedChanges ? '1' : '0.5';
    }
}

function markSectionChanged(sectionId) {
    sectionChanges[sectionId] = true;
    hasUnsavedChanges = true;
    updateSectionTitles();
    updateSaveButton();
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
function showHotkeyRecordingModal(action, callback = null) {
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
            if (callback) {
                // Use callback for settings page
                callback(recordedHotkey);
            } else {
                // Use original logic for hotkeys page
                pendingHotkeys[action] = recordedHotkey;
                renderHotkeysTable();
            }
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
    
    // Settings page buttons
    const refreshSettingsBtn = document.getElementById('refresh-settings');
    const saveSettingsBtn = document.getElementById('save-settings');
    const resetSettingsBtn = document.getElementById('reset-settings');
    const exportSettingsBtn = document.getElementById('export-settings');
    const importSettingsBtn = document.getElementById('import-settings');
    
    if (refreshSettingsBtn) refreshSettingsBtn.addEventListener('click', loadSettingsPage);
    if (saveSettingsBtn) saveSettingsBtn.addEventListener('click', saveAllSettings);
    if (resetSettingsBtn) resetSettingsBtn.addEventListener('click', resetAllSettings);
    if (exportSettingsBtn) exportSettingsBtn.addEventListener('click', exportAllSettings);
    if (importSettingsBtn) importSettingsBtn.addEventListener('click', importAllSettings);
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

// Settings Management State
let currentSettings = {};
let pendingSettings = {};

// Settings Management Functions
async function loadSettingsPage() {
    console.log("Loading settings page");
    
    try {
        const [settings, hotkeys] = await Promise.all([
            GetAppSettings(),
            GetHotkeys()
        ]);
        
        console.log("Loaded settings data:", { settings, hotkeys });
        
        currentSettings = settings || {};
        pendingSettings = { ...currentSettings };
        
        populateSettingsForm(pendingSettings);
        populateHotkeysInSettings(hotkeys || {});
        updateTriggerPlaceholder(); // Update placeholder with current prefix
        
    } catch (error) {
        console.error("Error loading settings page:", error);
    }
}

// Update the trigger input placeholder based on current settings
function updateTriggerPlaceholder() {
    const triggerInput = document.getElementById('trigger-input');
    if (triggerInput && currentSettings && currentSettings.triggerPrefix) {
        const prefix = currentSettings.triggerPrefix;
        triggerInput.placeholder = `Enter trigger (e.g., ${prefix}ty, ${prefix}hello)`;
    }
}

function populateSettingsForm(settings) {
    // Trigger prefix
    const triggerPrefix = document.getElementById('trigger-prefix');
    if (triggerPrefix && settings.triggerPrefix) {
        triggerPrefix.value = settings.triggerPrefix;
    }
    
    // Expand key
    const expandKey = document.getElementById('expand-key');
    if (expandKey && settings.expandKey) {
        expandKey.value = settings.expandKey;
    }
    
    // Typing timeout
    const typingTimeout = document.getElementById('typing-timeout');
    const typingTimeoutValue = document.getElementById('typing-timeout-value');
    const noTimeout = document.getElementById('no-timeout');
    
    if (settings.typingTimeout !== undefined) {
        const timeoutValue = settings.typingTimeout === 0 ? 2000 : settings.typingTimeout;
        if (typingTimeout) typingTimeout.value = timeoutValue;
        if (typingTimeoutValue) typingTimeoutValue.value = timeoutValue;
        if (noTimeout) noTimeout.checked = settings.typingTimeout === 0;
    }
    
    // Expansion method
    const expansionMethod = settings.expansionMethod || 'direct';
    const methodRadio = document.querySelector(`input[name="expansion-method"][value="${expansionMethod}"]`);
    if (methodRadio) methodRadio.checked = true;
    
    // Suggestions settings
    const suggestionsEnabled = document.getElementById('suggestions-enabled');
    if (suggestionsEnabled && settings.suggestionsEnabled !== undefined) {
        suggestionsEnabled.checked = settings.suggestionsEnabled;
    }
    
    const minQueryLength = document.getElementById('min-query-length');
    if (minQueryLength && settings.minQueryLength !== undefined) {
        minQueryLength.value = settings.minQueryLength;
    }
    
    const maxSuggestions = document.getElementById('max-suggestions');
    if (maxSuggestions && settings.maxSuggestions !== undefined) {
        maxSuggestions.value = settings.maxSuggestions;
    }
    
    const suggestionDelay = document.getElementById('suggestion-delay');
    if (suggestionDelay && settings.suggestionDelay !== undefined) {
        suggestionDelay.value = settings.suggestionDelay;
    }
    
    const suggestionHideDelay = document.getElementById('suggestion-hide-delay');
    if (suggestionHideDelay && settings.suggestionHideDelay !== undefined) {
        suggestionHideDelay.value = settings.suggestionHideDelay;
    }
    
    // Advanced settings
    const strictBoundaries = document.getElementById('strict-boundaries');
    if (strictBoundaries && settings.strictBoundaries !== undefined) {
        strictBoundaries.checked = settings.strictBoundaries;
    }
    
    const pauseOnFailure = document.getElementById('pause-on-failure');
    if (pauseOnFailure && settings.pauseOnFailure !== undefined) {
        pauseOnFailure.checked = settings.pauseOnFailure;
    }
    
    const bufferSize = document.getElementById('buffer-size');
    if (bufferSize && settings.bufferSize !== undefined) {
        bufferSize.value = settings.bufferSize;
    }
    
    const caseSensitive = document.getElementById('case-sensitive');
    if (caseSensitive && settings.caseSensitive !== undefined) {
        caseSensitive.checked = settings.caseSensitive;
    }
    
    // Setup event listeners for real-time updates
    setupSettingsEventListeners();
    
    // Store original settings for change detection
    originalSettings = { ...settings };
    
    // Reset change tracking
    hasUnsavedChanges = false;
    resetSectionChanges();
    updateSectionTitles();
    updateSaveButton();
}

function populateHotkeysInSettings(hotkeys) {
    // Populate hotkey displays in settings
    const hotkeyInputs = {
        'expandNow': 'hotkey-expand-now',
        'openPalette': 'hotkey-open-palette', 
        'togglePause': 'hotkey-toggle-pause'
    };
    
    Object.entries(hotkeyInputs).forEach(([action, inputId]) => {
        const input = document.getElementById(inputId);
        const hotkey = hotkeys[action];
        if (input && hotkey) {
            input.value = formatHotkey(hotkey);
        }
    });
}

function setupSettingsEventListeners() {
    // Helper function to track changes for specific sections
    function addChangeListener(elementId, sectionId) {
        const element = document.getElementById(elementId);
        if (element) {
            const eventType = element.type === 'checkbox' || element.type === 'radio' ? 'change' : 'input';
            element.addEventListener(eventType, () => {
                markSectionChanged(sectionId);
            });
        }
    }
    
    // Helper function for radio groups
    function addRadioGroupListener(groupName, sectionId) {
        const radios = document.querySelectorAll(`input[name="${groupName}"]`);
        radios.forEach(radio => {
            radio.addEventListener('change', () => {
                markSectionChanged(sectionId);
            });
        });
    }
    
    // Hotkeys section change tracking
    // Note: Hotkey changes are tracked in the recordHotkeyForSetting function
    
    // Expansion section change tracking
    addChangeListener('trigger-prefix', 'expansion');
    addChangeListener('expand-key', 'expansion');
    addChangeListener('typing-timeout', 'expansion');
    addChangeListener('typing-timeout-value', 'expansion');
    addChangeListener('no-timeout', 'expansion');
    
    // Suggestions section change tracking
    addChangeListener('suggestions-enabled', 'suggestions');
    addChangeListener('min-query-length', 'suggestions');
    addChangeListener('max-suggestions', 'suggestions');
    addChangeListener('suggestion-delay', 'suggestions');
    addChangeListener('suggestion-hide-delay', 'suggestions');
    
    // Advanced section change tracking
    addChangeListener('strict-boundaries', 'advanced');
    addChangeListener('pause-on-failure', 'advanced');
    addChangeListener('buffer-size', 'advanced');
    addChangeListener('case-sensitive', 'advanced');
    
    // Typing timeout synchronization
    const typingTimeout = document.getElementById('typing-timeout');
    const typingTimeoutValue = document.getElementById('typing-timeout-value');
    const noTimeout = document.getElementById('no-timeout');
    
    if (typingTimeout && typingTimeoutValue) {
        typingTimeout.addEventListener('input', () => {
            typingTimeoutValue.value = typingTimeout.value;
        });
        
        typingTimeoutValue.addEventListener('input', () => {
            typingTimeout.value = typingTimeoutValue.value;
        });
    }
    
    if (noTimeout && typingTimeout && typingTimeoutValue) {
        noTimeout.addEventListener('change', () => {
            const disabled = noTimeout.checked;
            typingTimeout.disabled = disabled;
            typingTimeoutValue.disabled = disabled;
            if (disabled) {
                typingTimeout.style.opacity = '0.5';
                typingTimeoutValue.style.opacity = '0.5';
            } else {
                typingTimeout.style.opacity = '1';
                typingTimeoutValue.style.opacity = '1';
            }
        });
    }
}

// Global functions for settings (called from onclick)
window.recordHotkeyForSetting = function(action, inputId) {
    console.log(`Recording hotkey for ${action} in settings`);
    showHotkeyRecordingModal(action, (hotkey) => {
        const input = document.getElementById(inputId);
        if (input) {
            input.value = formatHotkey(hotkey);
        }
        // Update pending hotkeys
        if (!pendingSettings.hotkeys) {
            pendingSettings.hotkeys = {};
        }
        pendingSettings.hotkeys[action] = hotkey;
        
        // Mark hotkeys section as changed
        markSectionChanged('hotkeys');
    });
};

async function saveAllSettings() {
    console.log("Saving all settings");
    
    try {
        // Collect all settings from the form
        const settings = collectSettingsFromForm();
        
        console.log("Collected settings:", settings);
        
        // Save settings
        await UpdateSettings(settings);
        
        // Update current settings
        currentSettings = { ...settings };
        pendingSettings = { ...settings };
        originalSettings = { ...settings };
        
        // Reset change tracking
        hasUnsavedChanges = false;
        resetSectionChanges();
        updateSectionTitles();
        updateSaveButton();
        
        // Show success message with more details
        const changedSettings = [];
        if (settings.triggerPrefix !== ':') changedSettings.push(`Prefix: ${settings.triggerPrefix}`);
        if (settings.expandKey !== 'Tab') changedSettings.push(`Expand Key: ${settings.expandKey}`);
        if (settings.expansionMethod !== 'direct') changedSettings.push(`Method: ${settings.expansionMethod}`);
        
        let message = "Settings saved and applied successfully!";
        if (changedSettings.length > 0) {
            message += "\n\nChanges applied immediately:\n• " + changedSettings.join("\n• ");
            message += "\n\nYou can test the new settings right away.";
        }
        
        alert(message);
        
        // Update trigger placeholder with new prefix
        updateTriggerPlaceholder();
        
        // Refresh expansion status to reflect new settings
        updateExpansionStatus();
        
    } catch (error) {
        console.error("Error saving settings:", error);
        alert("Error saving settings: " + error);
    }
}

function collectSettingsFromForm() {
    const settings = {};
    
    // Basic expansion settings
    const triggerPrefix = document.getElementById('trigger-prefix');
    if (triggerPrefix) settings.triggerPrefix = triggerPrefix.value;
    
    const expandKey = document.getElementById('expand-key');
    if (expandKey) settings.expandKey = expandKey.value;
    
    const noTimeout = document.getElementById('no-timeout');
    const typingTimeoutValue = document.getElementById('typing-timeout-value');
    if (noTimeout && typingTimeoutValue) {
        settings.typingTimeout = noTimeout.checked ? 0 : parseInt(typingTimeoutValue.value);
    }
    
    const expansionMethod = document.querySelector('input[name="expansion-method"]:checked');
    if (expansionMethod) settings.expansionMethod = expansionMethod.value;
    
    // Suggestions settings
    const suggestionsEnabled = document.getElementById('suggestions-enabled');
    if (suggestionsEnabled) settings.suggestionsEnabled = suggestionsEnabled.checked;
    
    const minQueryLength = document.getElementById('min-query-length');
    if (minQueryLength) settings.minQueryLength = parseInt(minQueryLength.value);
    
    const maxSuggestions = document.getElementById('max-suggestions');
    if (maxSuggestions) settings.maxSuggestions = parseInt(maxSuggestions.value);
    
    const suggestionDelay = document.getElementById('suggestion-delay');
    if (suggestionDelay) settings.suggestionDelay = parseInt(suggestionDelay.value);
    
    const suggestionHideDelay = document.getElementById('suggestion-hide-delay');
    if (suggestionHideDelay) settings.suggestionHideDelay = parseInt(suggestionHideDelay.value);
    
    // Advanced settings
    const strictBoundaries = document.getElementById('strict-boundaries');
    if (strictBoundaries) settings.strictBoundaries = strictBoundaries.checked;
    
    const pauseOnFailure = document.getElementById('pause-on-failure');
    if (pauseOnFailure) settings.pauseOnFailure = pauseOnFailure.checked;
    
    const bufferSize = document.getElementById('buffer-size');
    if (bufferSize) settings.bufferSize = parseInt(bufferSize.value);
    
    const caseSensitive = document.getElementById('case-sensitive');
    if (caseSensitive) settings.caseSensitive = caseSensitive.checked;
    
    // Include hotkeys if they were changed
    if (pendingSettings.hotkeys) {
        settings.hotkeys = pendingSettings.hotkeys;
    }
    
    return settings;
}

async function resetAllSettings() {
    console.log("Resetting all settings to defaults");
    
    if (confirm("Reset all settings to defaults? This will overwrite your current configuration.")) {
        try {
            const defaultSettings = await GetDefaultSettings();
            pendingSettings = { ...defaultSettings };
            populateSettingsForm(pendingSettings);
            
            // Also get default hotkeys
            const defaultHotkeys = await GetDefaultHotkeys();
            populateHotkeysInSettings(defaultHotkeys);
            
        } catch (error) {
            console.error("Error getting default settings:", error);
            alert("Error getting default settings: " + error);
        }
    }
}

async function exportAllSettings() {
    console.log("Exporting settings");
    
    try {
        const exportData = await ExportSettings();
        
        // Create download link
        const blob = new Blob([exportData], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `snipq-settings-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        
        alert("Settings exported successfully!");
        
    } catch (error) {
        console.error("Error exporting settings:", error);
        alert("Error exporting settings: " + error);
    }
}

async function importAllSettings() {
    console.log("Importing settings");
    
    // Create file input
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = async (event) => {
        const file = event.target.files[0];
        if (!file) return;
        
        try {
            const text = await file.text();
            await ImportSettings(text);
            
            alert("Settings imported successfully! Please restart the application for all changes to take effect.");
            
            // Reload the settings page
            loadSettingsPage();
            
        } catch (error) {
            console.error("Error importing settings:", error);
            alert("Error importing settings: " + error);
        }
    };
    
    input.click();
}
