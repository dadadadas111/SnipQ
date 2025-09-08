// Event management for backend events
import { EventsOn } from '../../wailsjs/runtime/runtime';

export class EventManager {
    constructor() {
        this.snippetTester = null;
        this.hotkeyManager = null;
    }

    init(snippetTester, hotkeyManager) {
        this.snippetTester = snippetTester;
        this.hotkeyManager = hotkeyManager;
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Listen for hotkey events from the backend
        EventsOn("hotkey:toggleWindow", () => {
            console.log("Hotkey event: Toggle Window");
            // Window toggle is handled by the backend, no frontend action needed
        });

        EventsOn("hotkey:expansionToggled", (data) => {
            console.log("Hotkey event: Expansion toggled", data);
            if (this.hotkeyManager) {
                this.hotkeyManager.updatePauseStatus(data.paused);
            }
        });

        EventsOn("hotkey:showSuggestions", () => {
            console.log("Hotkey event: Show Suggestions");
            // Switch to main tab and focus input to trigger suggestions
            if (this.snippetTester) {
                const mainTabElement = document.querySelector('[data-tab="main"]');
                if (mainTabElement) {
                    mainTabElement.click();
                    setTimeout(() => {
                        this.snippetTester.focusInput();
                        // Add a colon to trigger suggestions if input is empty
                        if (!this.snippetTester.triggerInput.value) {
                            this.snippetTester.triggerInput.value = ':';
                            this.snippetTester.triggerInput.dispatchEvent(new Event('input'));
                        }
                    }, 100);
                }
            }
        });

        EventsOn("hotkey:quickExpand", () => {
            console.log("Hotkey event: Quick Expand");
            if (this.snippetTester && this.snippetTester.triggerInput && this.snippetTester.triggerInput.value) {
                this.snippetTester.testExpansion();
            }
        });

        EventsOn("hotkey:focusSearch", () => {
            console.log("Hotkey event: Focus Search");
            // Switch to main tab and focus input
            if (this.snippetTester) {
                const mainTabElement = document.querySelector('[data-tab="main"]');
                if (mainTabElement) {
                    mainTabElement.click();
                    setTimeout(() => {
                        this.snippetTester.focusInput();
                    }, 100);
                }
            }
        });

        // Listen for focus-input events from the tray
        EventsOn("focus-input", () => {
            console.log("Focus-input event received from tray");
            if (this.snippetTester) {
                this.snippetTester.focusInput();
            }
        });
    }
}
