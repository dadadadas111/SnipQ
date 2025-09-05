// Hotkey management functionality
import { 
    GetHotkeys, 
    GetHotkeyStatuses, 
    UpdateHotkeys, 
    ValidateHotkey, 
    RefreshHotkeyStatuses, 
    GetDefaultHotkeys, 
    GetPauseState, 
    TogglePause 
} from '../../wailsjs/go/main/App';

export class HotkeyManager {
    constructor() {
        this.currentHotkeys = {};
        this.pendingHotkeys = {};
        this.hotkeyStatuses = {};
        this.isRecording = false;
        this.recordingAction = null;
        this._hotkeyModalOpen = false;
        this._currentModal = null;
        this._escapeHandler = null;
    }

    init() {
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Hotkey page buttons
        const refreshBtn = document.getElementById('refresh-hotkeys');
        const saveBtn = document.getElementById('save-hotkeys');
        const revertBtn = document.getElementById('revert-hotkeys');
        const resetBtn = document.getElementById('reset-defaults');
        const pauseToggleBtn = document.getElementById('pause-toggle-btn');
        
        if (refreshBtn) refreshBtn.addEventListener('click', () => this.refreshHotkeyStatuses());
        if (saveBtn) saveBtn.addEventListener('click', () => this.saveHotkeys());
        if (revertBtn) revertBtn.addEventListener('click', () => this.revertHotkeys());
        if (resetBtn) resetBtn.addEventListener('click', () => this.resetToDefaults());
        if (pauseToggleBtn) pauseToggleBtn.addEventListener('click', () => this.togglePause());
    }

    async loadHotkeysPage() {
        console.log("Loading hotkeys page");
        
        try {
            // Load current hotkeys and statuses
            const [hotkeys, statuses, pauseState] = await Promise.all([
                GetHotkeys(),
                GetHotkeyStatuses(),
                GetPauseState()
            ]);
            
            console.log("Loaded hotkeys data:", { hotkeys, statuses, pauseState });
            
            this.currentHotkeys = hotkeys || {};
            this.pendingHotkeys = { ...this.currentHotkeys };
            this.hotkeyStatuses = statuses || {};
            
            this.renderHotkeysTable();
            this.updatePauseStatus(pauseState);
            
        } catch (error) {
            console.error("Error loading hotkeys page:", error);
        }
    }

    renderHotkeysTable() {
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
            const hotkey = this.pendingHotkeys[actionKey] || { modifiers: [], key: '', enabled: false };
            const status = this.hotkeyStatuses[actionKey] || { status: 'Not Registered', error: '' };
            
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>
                    <div class="action-name">${actionInfo.name}</div>
                    <div class="action-description">${actionInfo.description}</div>
                </td>
                <td>
                    <div class="shortcut-display" id="shortcut-${actionKey}">
                        ${this.formatHotkey(hotkey)}
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
                                   onchange="window.hotkeyManager.toggleHotkeyEnabled('${actionKey}', this.checked)">
                            <span>Enabled</span>
                        </label>
                        <div class="hotkey-buttons">
                            <button class="btn small change" onclick="window.hotkeyManager.changeHotkey('${actionKey}')">Change</button>
                            <button class="btn small reset" onclick="window.hotkeyManager.resetHotkey('${actionKey}')">Reset</button>
                        </div>
                    </div>
                </td>
            `;
            tbody.appendChild(row);
        });
    }

    formatHotkey(hotkey) {
        if (!hotkey.enabled || !hotkey.key) {
            return '<em>Not set</em>';
        }
        
        const modifiers = hotkey.modifiers || [];
        const parts = [...modifiers, hotkey.key];
        return parts.join(' + ');
    }

    updatePauseStatus(paused) {
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

    toggleHotkeyEnabled(action, enabled) {
        console.log(`Toggling ${action} enabled: ${enabled}`);
        if (!this.pendingHotkeys[action]) {
            this.pendingHotkeys[action] = { modifiers: [], key: '', enabled: false };
        }
        this.pendingHotkeys[action].enabled = enabled;
        this.renderHotkeysTable();
    }

    changeHotkey(action) {
        console.log(`Changing hotkey for ${action}`);
        this.showHotkeyRecordingModal(action);
    }

    async resetHotkey(action) {
        console.log(`Resetting hotkey for ${action}`);
        try {
            const defaults = await GetDefaultHotkeys();
            if (defaults[action]) {
                this.pendingHotkeys[action] = { ...defaults[action] };
                this.renderHotkeysTable();
            }
        } catch (error) {
            console.error("Error resetting hotkey:", error);
        }
    }

    async refreshHotkeyStatuses() {
        console.log("Refreshing hotkey statuses");
        try {
            const statuses = await RefreshHotkeyStatuses();
            this.hotkeyStatuses = statuses || {};
            this.renderHotkeysTable();
        } catch (error) {
            console.error("Error refreshing hotkey statuses:", error);
        }
    }

    async saveHotkeys() {
        console.log("Saving hotkeys", this.pendingHotkeys);
        try {
            await UpdateHotkeys(this.pendingHotkeys);
            this.currentHotkeys = { ...this.pendingHotkeys };
            
            // Refresh statuses after saving
            await this.refreshHotkeyStatuses();
            
            alert("Hotkeys saved successfully!");
        } catch (error) {
            console.error("Error saving hotkeys:", error);
            alert("Error saving hotkeys: " + error);
        }
    }

    revertHotkeys() {
        console.log("Reverting hotkeys");
        this.pendingHotkeys = { ...this.currentHotkeys };
        this.renderHotkeysTable();
    }

    async resetToDefaults() {
        console.log("Resetting to default hotkeys");
        if (confirm("Reset all hotkeys to defaults? This will overwrite your current settings.")) {
            try {
                const defaults = await GetDefaultHotkeys();
                this.pendingHotkeys = { ...defaults };
                this.renderHotkeysTable();
            } catch (error) {
                console.error("Error getting default hotkeys:", error);
            }
        }
    }

    async togglePause() {
        console.log("Toggling pause state");
        try {
            const newState = await TogglePause();
            console.log("Pause toggled to:", newState);
            this.updatePauseStatus(newState);
        } catch (error) {
            console.error("Error toggling pause:", error);
        }
    }

    // Hotkey Recording Modal
    showHotkeyRecordingModal(action, callback = null) {
        // Re-entrancy guard: return early if modal already open
        if (this._hotkeyModalOpen) {
            console.log('Modal already open, ignoring request');
            return;
        }
        
        console.log(`Showing recording modal for ${action}`);
        this._hotkeyModalOpen = true;
        
        // Get the friendly action name
        const actions = {
            'openPalette': 'Open Palette',
            'togglePause': 'Toggle Pause',
            'expandNow': 'Expand Now'
        };
        const actionName = actions[action] || action;
        
        const modal = document.createElement('div');
        modal.className = 'hotkey-modal';
        // Make modal focusable and accessible
        modal.tabIndex = -1;
        modal.setAttribute('role', 'dialog');
        modal.setAttribute('aria-modal', 'true');
        modal.setAttribute('aria-labelledby', 'modal-title');
        
        modal.innerHTML = `
            <div class="hotkey-modal-content" style="position: relative;">
                <button id="close-modal-x" style="position: absolute; top: 16px; right: 16px; background: none; border: none; font-size: 1.7rem; color: #495057; cursor: pointer; z-index: 10;" title="Close">&times;</button>
                <h3 id="modal-title">Record Hotkey for ${actionName}</h3>
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
        this._currentModal = modal;
        
        // Focus the modal after appending
        modal.focus();
        
        const recordingDisplay = modal.querySelector('#recording-display');
        const recordingError = modal.querySelector('#recording-error');
        const saveBtn = modal.querySelector('#save-recording');
        const closeXBtn = modal.querySelector('#close-modal-x');
        
        let recordedHotkey = null;
        
        // Key event handler
        const handleKeyDown = (e) => {
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
                this.showRecordingError("You must use at least one modifier key (Ctrl, Alt, Shift, or Win)", recordingError);
                return;
            }
            
            const key = this.normalizeKey(e.key);
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
                    recordingDisplay.textContent = this.formatHotkey(hotkey);
                    recordingDisplay.classList.remove('listening');
                    saveBtn.style.display = 'inline-block';
                    this.hideRecordingError(recordingError);
                })
                .catch(error => {
                    console.error("Hotkey validation failed:", error);
                    this.showRecordingError(error.toString(), recordingError);
                });
        };
        
        const closeModal = () => {
            // Remove all event listeners
            document.removeEventListener('keydown', handleKeyDown);
            if (this._escapeHandler) {
                document.removeEventListener('keydown', this._escapeHandler);
                this._escapeHandler = null;
            }
            
            // Clear re-entrancy flag and current modal reference
            this._hotkeyModalOpen = false;
            this._currentModal = null;
            
            // Remove modal from DOM
            if (modal.parentNode) {
                document.body.removeChild(modal);
            }
        };
        
        // Escape key handler (separate from the recording handler)
        this._escapeHandler = (e) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                closeModal();
            }
        };
        
        // Event listeners
        document.addEventListener('keydown', handleKeyDown);
        document.addEventListener('keydown', this._escapeHandler);
        
        saveBtn.addEventListener('click', () => {
            if (recordedHotkey) {
                if (callback) {
                    callback(recordedHotkey);
                } else {
                    this.pendingHotkeys[action] = recordedHotkey;
                    this.renderHotkeysTable();
                }
            }
            closeModal();
        });
        
        closeXBtn.addEventListener('click', closeModal);
        modal.addEventListener('click', (e) => {
            if (e.target === modal) closeModal();
        });
    }

    showRecordingError(message, errorElement) {
        errorElement.textContent = message;
        errorElement.style.display = 'block';
    }
    
    hideRecordingError(errorElement) {
        errorElement.style.display = 'none';
    }
    
    normalizeKey(key) {
        // Normalize common key names
        const keyMap = {
            ' ': 'Space',
            '.': 'Period',
            ',': 'Comma',
            'Enter': 'Enter'
        };
        
        return keyMap[key] || key.toUpperCase();
    }
}
