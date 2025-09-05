// Settings management functionality
import { 
    GetAppSettings, 
    UpdateSettings, 
    GetDefaultSettings, 
    ExportSettings, 
    ImportSettings,
    GetHotkeys
} from '../../wailsjs/go/main/App';

export class SettingsManager {
    constructor() {
        this.currentSettings = {};
        this.pendingSettings = {};
        this.originalSettings = {};
        this.navigationManager = null;
    }

    init(navigationManager) {
        this.navigationManager = navigationManager;
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Settings page buttons
        const refreshSettingsBtn = document.getElementById('refresh-settings');
        const saveSettingsBtn = document.getElementById('save-settings');
        const resetSettingsBtn = document.getElementById('reset-settings');
        const exportSettingsBtn = document.getElementById('export-settings');
        const importSettingsBtn = document.getElementById('import-settings');
        
        if (refreshSettingsBtn) refreshSettingsBtn.addEventListener('click', () => this.loadSettingsPage());
        if (saveSettingsBtn) saveSettingsBtn.addEventListener('click', () => this.saveAllSettings());
        if (resetSettingsBtn) resetSettingsBtn.addEventListener('click', () => this.resetAllSettings());
        if (exportSettingsBtn) exportSettingsBtn.addEventListener('click', () => this.exportAllSettings());
        if (importSettingsBtn) importSettingsBtn.addEventListener('click', () => this.importAllSettings());
    }

    async loadSettingsPage() {
        console.log("Loading settings page");
        
        try {
            const [settings, hotkeys] = await Promise.all([
                GetAppSettings(),
                GetHotkeys()
            ]);
            
            console.log("Loaded settings data:", { settings, hotkeys });
            
            this.currentSettings = settings || {};
            this.pendingSettings = { ...this.currentSettings };
            
            this.populateSettingsForm(this.pendingSettings);
            this.populateHotkeysInSettings(hotkeys || {});
            
            // Store original settings for change detection
            this.originalSettings = { ...settings };
            
            // Reset change tracking
            if (this.navigationManager) {
                this.navigationManager.hasUnsavedChanges = false;
                this.navigationManager.resetSectionChanges();
                this.navigationManager.updateSectionTitles();
                this.navigationManager.updateSaveButton();
            }
            
        } catch (error) {
            console.error("Error loading settings page:", error);
        }
    }

    populateSettingsForm(settings) {
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
        this.setupSettingsEventListeners();
    }

    populateHotkeysInSettings(hotkeys) {
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
                input.value = this.formatHotkey(hotkey);
            }
        });
    }

    formatHotkey(hotkey) {
        if (!hotkey.enabled || !hotkey.key) {
            return 'Not set';
        }
        
        const modifiers = hotkey.modifiers || [];
        const parts = [...modifiers, hotkey.key];
        return parts.join(' + ');
    }

    setupSettingsEventListeners() {
        // Helper function to track changes for specific sections
        const addChangeListener = (elementId, sectionId) => {
            const element = document.getElementById(elementId);
            if (element) {
                element.addEventListener('input', () => {
                    if (this.navigationManager) {
                        this.navigationManager.markSectionChanged(sectionId);
                    }
                });
                element.addEventListener('change', () => {
                    if (this.navigationManager) {
                        this.navigationManager.markSectionChanged(sectionId);
                    }
                });
            }
        };
        
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
                    typingTimeout.value = 2000;
                    typingTimeoutValue.value = 2000;
                }
            });
        }
    }

    recordHotkeyForSetting(action, inputId) {
        console.log(`Recording hotkey for ${action} in settings`);
        // This would integrate with the hotkey manager
        // For now, just a placeholder
        alert(`Recording hotkey for ${action} - this would open the hotkey recorder`);
    }

    async saveAllSettings() {
        console.log("Saving all settings");
        
        try {
            // Collect all settings from the form
            const settings = this.collectSettingsFromForm();
            
            console.log("Collected settings:", settings);
            
            // Save settings
            await UpdateSettings(settings);
            
            // Update current settings
            this.currentSettings = { ...settings };
            this.pendingSettings = { ...settings };
            this.originalSettings = { ...settings };
            
            // Reset change tracking
            if (this.navigationManager) {
                this.navigationManager.hasUnsavedChanges = false;
                this.navigationManager.resetSectionChanges();
                this.navigationManager.updateSectionTitles();
                this.navigationManager.updateSaveButton();
            }
            
            // Show success message
            let message = "Settings saved and applied successfully!";
            alert(message);
            
        } catch (error) {
            console.error("Error saving settings:", error);
            alert("Error saving settings: " + error);
        }
    }

    collectSettingsFromForm() {
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
        if (this.pendingSettings.hotkeys) {
            settings.hotkeys = this.pendingSettings.hotkeys;
        }
        
        return settings;
    }

    async resetAllSettings() {
        console.log("Resetting all settings to defaults");
        
        if (confirm("Reset all settings to defaults? This will overwrite your current configuration.")) {
            try {
                const defaultSettings = await GetDefaultSettings();
                this.populateSettingsForm(defaultSettings);
                alert("Settings reset to defaults. Don't forget to save!");
            } catch (error) {
                console.error("Error resetting settings:", error);
                alert("Error resetting settings: " + error);
            }
        }
    }

    async exportAllSettings() {
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

    async importAllSettings() {
        console.log("Importing settings");
        
        // Create file input
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.json';
        input.onchange = async (event) => {
            const file = event.target.files[0];
            if (!file) return;
            
            try {
                const content = await file.text();
                await ImportSettings(content);
                
                alert("Settings imported successfully! The application will reload the settings.");
                
                // Reload settings page
                this.loadSettingsPage();
                
            } catch (error) {
                console.error("Error importing settings:", error);
                alert("Error importing settings: " + error);
            }
        };
        
        input.click();
    }
}
