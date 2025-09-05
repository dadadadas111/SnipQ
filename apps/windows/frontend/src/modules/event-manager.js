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
        EventsOn("hotkey:openPalette", () => {
            console.log("Hotkey event: Open Palette");
            // Switch to main tab and focus input
            if (this.snippetTester) {
                // Trigger tab switch via navigation manager
                document.querySelector('[data-tab="main"]').click();
                setTimeout(() => {
                    this.snippetTester.focusInput();
                }, 100);
            }
        });

        EventsOn("hotkey:pauseToggled", (data) => {
            console.log("Hotkey event: Pause toggled", data);
            if (this.hotkeyManager) {
                this.hotkeyManager.updatePauseStatus(data.paused);
            }
        });

        EventsOn("hotkey:expandNow", () => {
            console.log("Hotkey event: Expand Now");
            if (this.snippetTester && this.snippetTester.triggerInput && this.snippetTester.triggerInput.value) {
                this.snippetTester.testExpansion();
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
