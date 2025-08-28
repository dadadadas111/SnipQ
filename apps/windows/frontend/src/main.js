import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {GetGroups, GetSnippets, ExpandSnippet, PreviewSnippet, CreateSampleData, GetVaultInfo, ShowWindow, HideWindow, ToggleWindow, ExitApp} from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <div class="container">
        <header class="header">
            <img id="logo" class="logo" style="width: 32px; height: 32px;">
            <h1>SnipQ - Snippet Manager</h1>
            <div class="header-controls">
                <button class="btn-header" id="minimize-btn" title="Minimize to tray">−</button>
                <button class="btn-header" id="close-btn" title="Hide to tray">×</button>
            </div>
        </header>
        
        <div class="main-content">
            <div class="sidebar">
                <h3>Groups</h3>
                <div class="tray-controls">
                    <button class="btn small" id="hide-to-tray-btn">Hide to Tray</button>
                    <button class="btn small danger" id="exit-app-btn">Exit App</button>
                </div>
                <div id="vault-info" class="vault-info"></div>
                <div id="groups-list" class="groups-list"></div>
            </div>
            
            <div class="content">
                <div class="test-section">
                    <h3>Test Snippet Expansion</h3>
                    <div class="input-box">
                        <input class="input" id="trigger-input" type="text" placeholder="Enter trigger (e.g., :ty, :hello)" autocomplete="off" />
                        <button class="btn" id="expand-btn">Expand</button>
                        <button class="btn" id="preview-btn">Preview</button>
                        <button class="btn-clear" id="clear-btn" title="Clear field">✕</button>
                    </div>
                    <div class="result" id="result">Enter a trigger above to test expansion</div>
                </div>
                
                <div class="snippets-section">
                    <h3>Snippets</h3>
                    <div id="snippets-list" class="snippets-list"></div>
                </div>
            </div>
        </div>
    </div>
`;
document.getElementById('logo').src = logo;

let triggerInput = document.getElementById("trigger-input");
let resultElement = document.getElementById("result");
let groupsList = document.getElementById("groups-list");
let snippetsList = document.getElementById("snippets-list");
let vaultInfo = document.getElementById("vault-info");
let currentSelectedGroup = null;

// Focus on trigger input
triggerInput.focus();

// Add event listeners for buttons
document.getElementById('expand-btn').addEventListener('click', testExpansion);
document.getElementById('preview-btn').addEventListener('click', testPreview);
document.getElementById('clear-btn').addEventListener('click', clearField);

// Tray management event listeners
document.getElementById('minimize-btn').addEventListener('click', hideToTray);
document.getElementById('close-btn').addEventListener('click', hideToTray);
document.getElementById('hide-to-tray-btn').addEventListener('click', hideToTray);
document.getElementById('exit-app-btn').addEventListener('click', exitApp);

// Add Enter key support
triggerInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        testExpansion();
    }
});

// Load groups on startup
loadGroups();
checkVaultInfo();

// Tray management functions
function hideToTray() {
    console.log('Hiding to tray...');
    try {
        HideWindow()
            .then(() => {
                console.log('Window hidden to tray');
            })
            .catch((err) => {
                console.error('Error hiding to tray:', err);
            });
    } catch (err) {
        console.error('Exception hiding to tray:', err);
    }
}

function exitApp() {
    console.log('Exiting application...');
    if (confirm('Are you sure you want to exit SnipQ?')) {
        try {
            ExitApp()
                .then(() => {
                    console.log('Application exited');
                })
                .catch((err) => {
                    console.error('Error exiting app:', err);
                });
        } catch (err) {
            console.error('Exception exiting app:', err);
        }
    }
}

// Clear field function
function clearField() {
    triggerInput.value = '';
    resultElement.innerHTML = 'Enter a trigger above to test expansion';
    triggerInput.focus();
}

// Test expansion function
function testExpansion() {
    let trigger = triggerInput.value;
    if (trigger === "") {
        resultElement.innerText = "Please enter a trigger";
        return;
    }

    console.log('Testing expansion for:', trigger);
    try {
        ExpandSnippet(trigger)
            .then((result) => {
                console.log('Expansion result:', result);
                resultElement.innerHTML = `
                    <div><strong>Expanded:</strong> ${result.output}</div>
                    <div><strong>Snippet:</strong> ${result.usedSnippet}</div>
                    <div><strong>Parameters:</strong></div>
                    <pre>${JSON.stringify(result.usedParams, null, 2)}</pre>
                `;
            })
            .catch((err) => {
                console.error('Expansion error:', err);
                resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Expansion exception:', err);
        resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

function testPreview() {
    let trigger = triggerInput.value;
    if (trigger === "") {
        resultElement.innerText = "Please enter a trigger";
        return;
    }

    console.log('Testing preview for:', trigger);
    try {
        PreviewSnippet(trigger)
            .then((result) => {
                console.log('Preview result:', result);
                resultElement.innerHTML = `<strong>Preview:</strong> ${result}`;
            })
            .catch((err) => {
                console.error('Preview error:', err);
                resultElement.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Preview exception:', err);
        resultElement.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

function loadGroups() {
    console.log('Loading groups...');
    try {
        GetGroups()
            .then((groups) => {
                console.log('Groups loaded:', groups);
                groupsList.innerHTML = '';
                
                if (!groups || groups.length === 0) {
                    groupsList.innerHTML = '<div class="error">No groups found. The vault might be empty.</div>';
                    return;
                }
                
                groups.forEach(group => {
                    const groupElement = document.createElement('div');
                    groupElement.className = 'group-item';
                    groupElement.innerHTML = `
                        <div class="group-header" onclick="loadSnippets('${group.id}')" data-group-id="${group.id}">
                            <span>${group.icon || '📁'} ${group.name}</span>
                            <span class="group-id">${group.id}</span>
                        </div>
                    `;
                    groupsList.appendChild(groupElement);
                });
            })
            .catch((err) => {
                console.error('Error loading groups:', err);
                groupsList.innerHTML = `<div class="error">Error loading groups: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception loading groups:', err);
        groupsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}

window.loadSnippets = function(groupId) {
    console.log('Loading snippets for group:', groupId);
    
    // Update group selection state
    currentSelectedGroup = groupId;
    document.querySelectorAll('.group-header').forEach(header => {
        header.classList.remove('active');
    });
    document.querySelector(`[data-group-id="${groupId}"]`).classList.add('active');
    
    try {
        GetSnippets(groupId)
            .then((snippets) => {
                console.log('Snippets loaded:', snippets);
                snippetsList.innerHTML = `<h4>Snippets in ${groupId}</h4>`;
                
                if (!snippets || snippets.length === 0) {
                    snippetsList.innerHTML += '<div class="error">No snippets found in this group.</div>';
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
                        <div class="snippet-template"><pre>${snippet.template}</pre></div>
                        <div class="snippet-actions">
                            <button class="btn small" onclick="testTrigger('${snippet.trigger}')">Test This</button>
                        </div>
                    `;
                    snippetsList.appendChild(snippetElement);
                });
            })
            .catch((err) => {
                console.error('Error loading snippets:', err);
                snippetsList.innerHTML = `<div class="error">Error loading snippets: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception loading snippets:', err);
        snippetsList.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
};

// Helper function to test a specific trigger
window.testTrigger = function(trigger) {
    triggerInput.value = trigger;
    testExpansion();
    // Scroll to top to view the test section
    document.querySelector('.test-section').scrollIntoView({ 
        behavior: 'smooth', 
        block: 'start' 
    });
};

function checkVaultInfo() {
    try {
        GetVaultInfo()
            .then((info) => {
                console.log('Vault info:', info);
                vaultInfo.innerHTML = `
                    <div class="vault-info-content">
                        <small>Groups: ${info.groups} | Snippets: ${info.snippets}</small>
                    </div>
                `;
            })
            .catch((err) => {
                console.error('Error getting vault info:', err);
                vaultInfo.innerHTML = `<div class="error">Error: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception getting vault info:', err);
        vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
};

function createSampleData() {
    try {
        CreateSampleData()
            .then(() => {
                console.log('Sample data created');
                vaultInfo.innerHTML = '<div class="success">Sample data created!</div>';
                // Refresh the UI
                loadGroups();
                checkVaultInfo();
            })
            .catch((err) => {
                console.error('Error creating sample data:', err);
                vaultInfo.innerHTML = `<div class="error">Error creating sample data: ${err}</div>`;
            });
    } catch (err) {
        console.error('Exception creating sample data:', err);
        vaultInfo.innerHTML = `<div class="error">Exception: ${err}</div>`;
    }
}
