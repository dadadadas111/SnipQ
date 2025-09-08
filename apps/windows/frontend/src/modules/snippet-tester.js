// Snippet testing and expansion functionality
import { ExpandSnippet, PreviewSnippet } from '../../wailsjs/go/main/App';

export class SnippetTester {
    constructor() {
        this.triggerInput = null;
        this.resultElement = null;
    }

    init() {
        this.triggerInput = document.getElementById("trigger-input");
        this.resultElement = document.getElementById("result");
        this.setupEventListeners();
        this.setupExampleListeners();
        
        // Focus on trigger input
        if (this.triggerInput) {
            this.triggerInput.focus();
        }
    }

    setupEventListeners() {
        // Main tab buttons
        const expandBtn = document.getElementById('expand-btn');
        const previewBtn = document.getElementById('preview-btn');
        const clearBtn = document.getElementById('clear-btn');
        
        if (expandBtn) expandBtn.addEventListener('click', () => this.testExpansion());
        if (previewBtn) previewBtn.addEventListener('click', () => this.testPreview());
        if (clearBtn) clearBtn.addEventListener('click', () => this.clearField());

        // Add Enter key support
        if (this.triggerInput) {
            this.triggerInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.testExpansion();
                }
            });
        }
    }

    setupExampleListeners() {
        // Add click listeners to all example trigger buttons
        const exampleButtons = document.querySelectorAll('.example-trigger');
        exampleButtons.forEach(button => {
            button.addEventListener('click', () => {
                const trigger = button.getAttribute('data-trigger');
                if (trigger) {
                    console.log('Example trigger clicked:', trigger);
                    this.testTrigger(trigger);
                }
            });
        });
    }

    clearField() {
        this.triggerInput.value = '';
        this.resultElement.innerHTML = 'Enter a trigger above to test expansion';
        this.triggerInput.focus();
    }

    testExpansion() {
        let trigger = this.triggerInput.value;
        if (trigger === "") {
            this.resultElement.innerText = "Please enter a trigger";
            return;
        }

        console.log('Testing expansion for:', trigger);
        try {
            ExpandSnippet(trigger)
                .then((result) => {
                    console.log('Expansion result:', result);
                    this.resultElement.innerHTML = `
                        <div><strong>Expanded:</strong> ${result.output}</div>
                        <div><strong>Snippet:</strong> ${result.usedSnippet}</div>
                        <div><strong>Parameters:</strong></div>
                        <pre>${JSON.stringify(result.usedParams, null, 2)}</pre>
                    `;
                })
                .catch((err) => {
                    console.error('Expansion error:', err);
                    this.resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
                });
        } catch (err) {
            console.error('Expansion exception:', err);
            this.resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
        }
    }

    testPreview() {
        let trigger = this.triggerInput.value;
        if (trigger === "") {
            this.resultElement.innerText = "Please enter a trigger";
            return;
        }

        console.log('Testing preview for:', trigger);
        try {
            PreviewSnippet(trigger)
                .then((result) => {
                    console.log('Preview result:', result);
                    this.resultElement.innerHTML = `<strong>Preview:</strong> ${result}`;
                })
                .catch((err) => {
                    console.error('Preview error:', err);
                    this.resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
                });
        } catch (err) {
            console.error('Preview exception:', err);
            this.resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
        }
    }

    // Helper function to test a specific trigger (used by other modules)
    testTrigger(trigger) {
        this.triggerInput.value = trigger;
        this.testExpansion();
        // Scroll to top to view the test section
        document.querySelector('.test-section').scrollIntoView({ 
            behavior: 'smooth', 
            block: 'start' 
        });
    }

    focusInput() {
        if (this.triggerInput) {
            this.triggerInput.focus();
            this.triggerInput.select(); // Select any existing text
        }
    }

    updateTriggerPlaceholder(prefix) {
        if (this.triggerInput && prefix) {
            this.triggerInput.placeholder = `Enter trigger (e.g., ${prefix}ty, ${prefix}hello)`;
        }
    }
}
