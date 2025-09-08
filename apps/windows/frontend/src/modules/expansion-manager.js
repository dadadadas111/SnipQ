// Expansion status and control functionality
import { 
    GetExpansionStats, 
    IsExpansionEnabled, 
    SetExpansionEnabled, 
    ClearExpansionBuffer 
} from '../../wailsjs/go/main/App';

export class ExpansionManager {
    constructor() {
        this.statusElement = null;
        this.bufferElement = null;
        this.refreshInterval = null;
    }

    init() {
        this.statusElement = document.getElementById('expansion-status');
        this.bufferElement = document.getElementById('expansion-buffer');
        
        this.setupEventListeners();
        this.updateExpansionStatus();
        
        // Set up periodic refresh of expansion status
        this.refreshInterval = setInterval(() => this.updateExpansionStatus(), 2000);
    }

    destroy() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
    }

    setupEventListeners() {
        const toggleExpansionBtn = document.getElementById('toggle-expansion');
        const clearBufferBtn = document.getElementById('clear-buffer');
        const refreshExpansionBtn = document.getElementById('refresh-expansion');
        
        if (toggleExpansionBtn) toggleExpansionBtn.addEventListener('click', () => this.toggleExpansion());
        if (clearBufferBtn) clearBufferBtn.addEventListener('click', () => this.clearExpansionBuffer());
        if (refreshExpansionBtn) refreshExpansionBtn.addEventListener('click', () => this.updateExpansionStatus());
    }

    async updateExpansionStatus() {
        try {
            const [stats, enabled] = await Promise.all([
                GetExpansionStats(),
                IsExpansionEnabled()
            ]);
            
            if (this.statusElement) {
                this.statusElement.textContent = enabled ? 'Active' : 'Paused';
                this.statusElement.className = `status-value ${enabled ? 'active' : 'paused'}`;
            }
            
            if (this.bufferElement) {
                const buffer = stats?.buffer || '';
                this.bufferElement.textContent = buffer || '-';
                this.bufferElement.title = buffer ? `Current buffer: ${buffer}` : 'No current buffer';
            }
            
        } catch (error) {
            console.error('Error updating expansion status:', error);
        }
    }

    async toggleExpansion() {
        try {
            const currentState = await IsExpansionEnabled();
            await SetExpansionEnabled(!currentState);
            this.updateExpansionStatus();
            console.log('Expansion toggled to:', !currentState);
        } catch (error) {
            console.error('Error toggling expansion:', error);
        }
    }

    async clearExpansionBuffer() {
        try {
            await ClearExpansionBuffer();
            this.updateExpansionStatus();
            console.log('Expansion buffer cleared');
        } catch (error) {
            console.error('Error clearing expansion buffer:', error);
        }
    }
}
