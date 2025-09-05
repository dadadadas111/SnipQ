// Navigation and tab management functionality

export class NavigationManager {
    constructor() {
        this.currentTab = 'main';
        this.navTabs = null;
        this.tabContents = null;
        this.hasUnsavedChanges = false;
        this.sectionChanges = {
            hotkeys: false,
            expansion: false,
            suggestions: false,
            advanced: false
        };
    }

    init() {
        this.navTabs = document.querySelectorAll('.nav-tab');
        this.tabContents = document.querySelectorAll('.tab-content');
        this.setupEventListeners();
    }

    setupEventListeners() {
        this.navTabs.forEach(tab => {
            tab.addEventListener('click', () => {
                this.switchTab(tab.dataset.tab);
            });
        });
    }

    switchTab(tabName) {
        console.log(`Switching to tab: ${tabName}`);
        
        // Check for unsaved changes when switching to main tab
        if (tabName === 'main' && this.hasUnsavedChanges) {
            const confirmSwitch = confirm(
                'You have unsaved changes in Settings. Are you sure you want to switch tabs?\n\n' +
                'Your changes will be lost if you continue without saving.'
            );
            if (!confirmSwitch) {
                return; // Don't switch tabs
            }
            // Reset changes if user confirms
            this.hasUnsavedChanges = false;
            this.resetSectionChanges();
            this.updateSectionTitles();
            this.updateSaveButton();
        }
        
        // Update nav buttons
        this.navTabs.forEach(tab => {
            tab.classList.remove('active');
            if (tab.dataset.tab === tabName) {
                tab.classList.add('active');
            }
        });
        
        // Update tab content
        this.tabContents.forEach(content => {
            content.classList.remove('active');
            if (content.id === `${tabName}-tab`) {
                content.classList.add('active');
            }
        });
        
        this.currentTab = tabName;
        
        // Emit tab change event
        this.onTabChanged(tabName);
    }

    onTabChanged(tabName) {
        // This will be overridden by the main app
        console.log(`Tab changed to: ${tabName}`);
    }

    // Settings change management functions
    resetSectionChanges() {
        this.sectionChanges = {
            hotkeys: false,
            expansion: false,
            suggestions: false,
            advanced: false
        };
    }

    updateSectionTitles() {
        const sections = [
            { id: 'hotkeys', title: '🎹 Hotkeys' },
            { id: 'expansion', title: '⚡ Text Expansion' },
            { id: 'suggestions', title: '💡 Smart Suggestions' },
            { id: 'advanced', title: '🔧 Advanced Settings' }
        ];
        
        sections.forEach(section => {
            const titleElement = document.querySelector(`[data-section="${section.id}"]`);
            if (titleElement) {
                const hasChanges = this.sectionChanges[section.id];
                titleElement.textContent = hasChanges ? `${section.title} (*)` : section.title;
            }
        });
    }

    updateSaveButton() {
        const saveButton = document.getElementById('save-settings');
        if (saveButton) {
            saveButton.disabled = !this.hasUnsavedChanges;
            saveButton.style.opacity = this.hasUnsavedChanges ? '1' : '0.5';
        }
    }

    markSectionChanged(sectionId) {
        this.sectionChanges[sectionId] = true;
        this.hasUnsavedChanges = true;
        this.updateSectionTitles();
        this.updateSaveButton();
    }
}
