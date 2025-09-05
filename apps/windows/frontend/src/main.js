import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import { TemplateLoader } from './utils/template-loader.js';
import { NavigationManager } from './modules/navigation.js';
import { SnippetTester } from './modules/snippet-tester.js';
import { GroupsManager } from './modules/groups-manager.js';
import { ExpansionManager } from './modules/expansion-manager.js';
import { HotkeyManager } from './modules/hotkey-manager.js';
import { SettingsManager } from './modules/settings-manager.js';
import { EventManager } from './modules/event-manager.js';

// Helper function to escape HTML entities to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Main Application Class
class SnipQApp {
    constructor() {
        this.navigationManager = new NavigationManager();
        this.snippetTester = new SnippetTester();
        this.groupsManager = new GroupsManager();
        this.expansionManager = new ExpansionManager();
        this.hotkeyManager = new HotkeyManager();
        this.settingsManager = new SettingsManager();
        this.eventManager = new EventManager();
    }

    async init() {
        console.log('Initializing SnipQ App...');
        
        try {
            // Load and render the main template
            await this.loadTemplate();
            
            // Set the logo
            document.getElementById('logo').src = logo;
            
            // Initialize all modules
            this.initializeModules();
            
            // Setup inter-module communication
            this.setupModuleCommunication();
            
            console.log('SnipQ App initialized successfully');
            
        } catch (error) {
            console.error('Error initializing app:', error);
            this.showError('Failed to initialize application: ' + escapeHtml(error.message));
        }
    }

    async loadTemplate() {
        try {
            const templateUrl = new URL('./templates/app-template.html', import.meta.url).href;
            const templateContent = await TemplateLoader.loadTemplate(templateUrl);
            TemplateLoader.renderTemplate('#app', templateContent);
        } catch (error) {
            console.error('Error loading template:', error);
            // Fallback to a simple error message
            const escapedErrorMessage = escapeHtml(error.message);
            document.querySelector('#app').innerHTML = `
                <div class="error-container">
                    <h1>Error Loading Application</h1>
                    <p>Failed to load the application template. Please refresh the page.</p>
                    <p>Error: ${escapedErrorMessage}</p>
                </div>
            `;
            throw error;
        }
    }

    initializeModules() {
        // Initialize navigation manager first
        this.navigationManager.init();
        
        // Initialize other modules
        this.snippetTester.init();
        this.groupsManager.init();
        this.expansionManager.init();
        this.hotkeyManager.init();
        this.settingsManager.init(this.navigationManager);
        this.eventManager.init(this.snippetTester, this.hotkeyManager);
    }

    setupModuleCommunication() {
        // Set up callbacks and event handlers between modules
        
        // Navigation manager tab change handler
        this.navigationManager.onTabChanged = (tabName) => {
            if (tabName === 'settings') {
                this.settingsManager.loadSettingsPage();
            }
        };
        
        // Groups manager test trigger callback
        this.groupsManager.setTestTriggerCallback((trigger) => {
            this.snippetTester.testTrigger(trigger);
        });
        
        // Settings change callback for trigger placeholder update
        const originalSaveSettings = this.settingsManager.saveAllSettings.bind(this.settingsManager);
        this.settingsManager.saveAllSettings = async () => {
            await originalSaveSettings();
            // Update trigger placeholder after settings save
            const triggerPrefix = document.getElementById('trigger-prefix')?.value || ':';
            this.snippetTester.updateTriggerPlaceholder(triggerPrefix);
        };
        
        // Expose managers globally for onclick handlers (temporary solution)
        window.navigationManager = this.navigationManager;
        window.snippetTester = this.snippetTester;
        window.groupsManager = this.groupsManager;
        window.expansionManager = this.expansionManager;
        window.hotkeyManager = this.hotkeyManager;
        window.settingsManager = this.settingsManager;
        
        // Global functions for onclick handlers
        window.recordHotkeyForSetting = (action, inputId) => {
            this.hotkeyManager.showHotkeyRecordingModal(action, (hotkey) => {
                const input = document.getElementById(inputId);
                if (input) {
                    input.value = this.hotkeyManager.formatHotkey(hotkey);
                }
                // Mark hotkeys section as changed
                this.navigationManager.markSectionChanged('hotkeys');
                
                // Update pending settings
                if (!this.settingsManager.pendingSettings.hotkeys) {
                    this.settingsManager.pendingSettings.hotkeys = {};
                }
                this.settingsManager.pendingSettings.hotkeys[action] = hotkey;
            });
        };
    }

    showError(message) {
        const escapedMessage = escapeHtml(message);
        document.querySelector('#app').innerHTML = `
            <div class="error-container" style="padding: 2rem; text-align: center;">
                <h1 style="color: #dc3545;">Application Error</h1>
                <p style="margin: 1rem 0;">${escapedMessage}</p>
                <button onclick="location.reload()" style="padding: 0.5rem 1rem; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">
                    Reload Application
                </button>
            </div>
        `;
    }

    destroy() {
        // Clean up resources
        if (this.expansionManager) {
            this.expansionManager.destroy();
        }
    }
}

// Initialize the application when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    const app = new SnipQApp();
    await app.init();
    
    // Store app instance globally for debugging
    window.snipqApp = app;
    
    // Cleanup on page unload
    window.addEventListener('beforeunload', () => {
        app.destroy();
    });
});
