// Groups and snippets management functionality
import { GetGroups, GetSnippets, GetVaultInfo, CreateSampleData } from '../../wailsjs/go/main/App';

export class GroupsManager {
    constructor() {
        this.groupsList = null;
        this.snippetsList = null;
        this.vaultInfo = null;
        this.currentSelectedGroup = null;
        this.onTestTrigger = null; // Callback for testing triggers
    }

    init() {
        this.groupsList = document.getElementById("groups-list");
        this.snippetsList = document.getElementById("snippets-list");
        this.vaultInfo = document.getElementById("vault-info");
        
        // Load initial data
        this.loadGroups();
        this.checkVaultInfo();
    }

    setTestTriggerCallback(callback) {
        this.onTestTrigger = callback;
    }

    loadGroups() {
        console.log('Loading groups...');
        try {
            GetGroups()
                .then((groups) => {
                    console.log('Groups loaded:', groups);
                    this.groupsList.innerHTML = '';
                    
                    if (!groups || groups.length === 0) {
                        this.groupsList.innerHTML = `
                            <div class="error">
                                No groups found. The vault might be empty.
                                <div style="margin-top: 1rem;">
                                    <button class="btn small" onclick="window.groupsManager.createSampleData()">Create Sample Data</button>
                                </div>
                            </div>
                        `;
                        return;
                    }
                    
                    groups.forEach(group => {
                        // Create container div with class 'group-item'
                        const groupElement = document.createElement('div');
                        groupElement.className = 'group-item';
                        
                        // Create 'group-header' div
                        const groupHeader = document.createElement('div');
                        groupHeader.className = 'group-header';
                        groupHeader.dataset.groupId = group.id;
                        
                        // Create and safely set icon and name span
                        const nameSpan = document.createElement('span');
                        nameSpan.textContent = `${group.icon || '📁'} ${group.name}`;
                        
                        // Create and safely set ID span
                        const idSpan = document.createElement('span');
                        idSpan.className = 'group-id';
                        idSpan.textContent = group.id;
                        
                        // Append spans to header
                        groupHeader.appendChild(nameSpan);
                        groupHeader.appendChild(idSpan);
                        
                        // Attach click handler via addEventListener
                        groupHeader.addEventListener('click', () => this.loadSnippets(group.id));
                        
                        // Append header to container
                        groupElement.appendChild(groupHeader);
                        
                        // Append container to groups list
                        this.groupsList.appendChild(groupElement);
                    });
                })
                .catch((err) => {
                    console.error('Error loading groups:', err);
                    const errorDiv = document.createElement('div');
                    errorDiv.className = 'error';
                    errorDiv.textContent = 'Error loading groups: ' + String(err);
                    this.groupsList.innerHTML = '';
                    this.groupsList.appendChild(errorDiv);
                });
        } catch (err) {
            console.error('Exception loading groups:', err);
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.textContent = 'Exception: ' + String(err);
            this.groupsList.innerHTML = '';
            this.groupsList.appendChild(errorDiv);
        }
    }

    loadSnippets(groupId) {
        console.log('Loading snippets for group:', groupId);
        
        // Update group selection state
        this.currentSelectedGroup = groupId;
        document.querySelectorAll('.group-header').forEach(header => {
            header.classList.remove('active');
        });
        document.querySelector(`[data-group-id="${groupId}"]`).classList.add('active');
        
        try {
            GetSnippets(groupId)
                .then((snippets) => {
                    console.log('Snippets loaded:', snippets);
                    this.snippetsList.innerHTML = `<h4>Snippets in ${groupId}</h4>`;
                    
                    if (!snippets || snippets.length === 0) {
                        this.snippetsList.innerHTML += '<div class="error">No snippets found in this group.</div>';
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
                            <div class="snippet-actions">
                                <button class="btn small" onclick="window.groupsManager.testTrigger('${snippet.trigger}')">Test</button>
                            </div>
                        `;
                        this.snippetsList.appendChild(snippetElement);
                    });
                })
                .catch((err) => {
                    console.error('Error loading snippets:', err);
                    const errorDiv = document.createElement('div');
                    errorDiv.className = 'error';
                    errorDiv.textContent = 'Error loading snippets: ' + String(err);
                    this.snippetsList.innerHTML = '';
                    this.snippetsList.appendChild(errorDiv);
                });
        } catch (err) {
            console.error('Exception loading snippets:', err);
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.textContent = 'Exception: ' + String(err);
            this.snippetsList.innerHTML = '';
            this.snippetsList.appendChild(errorDiv);
        }
    }

    testTrigger(trigger) {
        if (this.onTestTrigger) {
            this.onTestTrigger(trigger);
        }
    }

    checkVaultInfo() {
        try {
            GetVaultInfo()
                .then((info) => {
                    console.log('Vault info:', info);
                    this.vaultInfo.innerHTML = `
                        <div class="vault-info-content">
                            <small>Groups: ${info.groups} | Snippets: ${info.snippets}</small>
                        </div>
                    `;
                })
                .catch((err) => {
                    console.error('Error getting vault info:', err);
                    const errorDiv = document.createElement('div');
                    errorDiv.className = 'error';
                    errorDiv.textContent = 'Error: ' + String(err);
                    this.vaultInfo.innerHTML = '';
                    this.vaultInfo.appendChild(errorDiv);
                });
        } catch (err) {
            console.error('Exception getting vault info:', err);
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.textContent = 'Exception: ' + String(err);
            this.vaultInfo.innerHTML = '';
            this.vaultInfo.appendChild(errorDiv);
        }
    }

    createSampleData() {
        try {
            CreateSampleData()
                .then(() => {
                    console.log('Sample data created');
                    this.vaultInfo.innerHTML = '<div class="success">Sample data created!</div>';
                    // Refresh the UI
                    this.loadGroups();
                    this.checkVaultInfo();
                })
                .catch((err) => {
                    console.error('Error creating sample data:', err);
                    const errorDiv = document.createElement('div');
                    errorDiv.className = 'error';
                    errorDiv.textContent = 'Error creating sample data: ' + String(err);
                    this.vaultInfo.innerHTML = '';
                    this.vaultInfo.appendChild(errorDiv);
                });
        } catch (err) {
            console.error('Exception creating sample data:', err);
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error';
            errorDiv.textContent = 'Exception: ' + String(err);
            this.vaultInfo.innerHTML = '';
            this.vaultInfo.appendChild(errorDiv);
        }
    }
}
