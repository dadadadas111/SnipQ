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
                        const groupElement = document.createElement('div');
                        groupElement.className = 'group-item';
                        groupElement.innerHTML = `
                            <div class="group-header" onclick="window.groupsManager.loadSnippets('${group.id}')" data-group-id="${group.id}">
                                <span>${group.icon || '📁'} ${group.name}</span>
                                <span class="group-id">${group.id}</span>
                            </div>
                        `;
                        this.groupsList.appendChild(groupElement);
                    });
                })
                .catch((err) => {
                    console.error('Error loading groups:', err);
                    this.groupsList.innerHTML = `<div class="error">Error loading groups: ${err}</div>`;
                });
        } catch (err) {
            console.error('Exception loading groups:', err);
            this.groupsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
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
                    this.snippetsList.innerHTML = `<div class="error">Error loading snippets: ${err}</div>`;
                });
        } catch (err) {
            console.error('Exception loading snippets:', err);
            this.snippetsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
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
                    this.vaultInfo.innerHTML = `<div class="error">Error: ${err}</div>`;
                });
        } catch (err) {
            console.error('Exception getting vault info:', err);
            this.vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
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
                    this.vaultInfo.innerHTML = `<div class="error">Error creating sample data: ${err}</div>`;
                });
        } catch (err) {
            console.error('Exception creating sample data:', err);
            this.vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
        }
    }
}
