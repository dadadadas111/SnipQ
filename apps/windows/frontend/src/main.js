import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {GetGroups, GetSnippets, ExpandSnippet, PreviewSnippet} from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <div class="container">
        <header class="header">
            <img id="logo" class="logo" style="width: 32px; height: 32px;">
            <h1>SnipQ - Snippet Manager</h1>
        </header>
        
        <div class="main-content">
            <div class="sidebar">
                <h3>Groups</h3>
                <div id="groups-list" class="groups-list"></div>
            </div>
            
            <div class="content">
                <div class="test-section">
                    <h3>Test Snippet Expansion</h3>
                    <div class="input-box">
                        <input class="input" id="trigger-input" type="text" placeholder="Enter trigger (e.g., :ty?lang=vi)" autocomplete="off" />
                        <button class="btn" onclick="testExpansion()">Expand</button>
                        <button class="btn" onclick="testPreview()">Preview</button>
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

// Focus on trigger input
triggerInput.focus();

// Load groups on startup
loadGroups();

// Setup the test functions
window.testExpansion = function () {
    let trigger = triggerInput.value;
    if (trigger === "") return;

    try {
        ExpandSnippet(trigger)
            .then((result) => {
                resultElement.innerHTML = `
                    <strong>Expanded:</strong> ${result.output}<br>
                    <strong>Snippet:</strong> ${result.usedSnippet}<br>
                    <strong>Parameters:</strong> ${JSON.stringify(result.usedParams, null, 2)}
                `;
            })
            .catch((err) => {
                resultElement.innerText = `Error: ${err}`;
                console.error(err);
            });
    } catch (err) {
        resultElement.innerText = `Error: ${err}`;
        console.error(err);
    }
};

window.testPreview = function () {
    let trigger = triggerInput.value;
    if (trigger === "") return;

    try {
        PreviewSnippet(trigger)
            .then((result) => {
                resultElement.innerHTML = `<strong>Preview:</strong> ${result}`;
            })
            .catch((err) => {
                resultElement.innerText = `Error: ${err}`;
                console.error(err);
            });
    } catch (err) {
        resultElement.innerText = `Error: ${err}`;
        console.error(err);
    }
};

function loadGroups() {
    try {
        GetGroups()
            .then((groups) => {
                groupsList.innerHTML = '';
                groups.forEach(group => {
                    const groupElement = document.createElement('div');
                    groupElement.className = 'group-item';
                    groupElement.innerHTML = `
                        <div class="group-header" onclick="loadSnippets('${group.id}')">
                            <span>${group.icon || '📁'} ${group.name}</span>
                            <span class="group-id">${group.id}</span>
                        </div>
                    `;
                    groupsList.appendChild(groupElement);
                });
            })
            .catch((err) => {
                groupsList.innerHTML = `<div class="error">Error loading groups: ${err}</div>`;
                console.error(err);
            });
    } catch (err) {
        groupsList.innerHTML = `<div class="error">Error: ${err}</div>`;
        console.error(err);
    }
}

window.loadSnippets = function(groupId) {
    try {
        GetSnippets(groupId)
            .then((snippets) => {
                snippetsList.innerHTML = '<h4>Snippets in ' + groupId + '</h4>';
                snippets.forEach(snippet => {
                    const snippetElement = document.createElement('div');
                    snippetElement.className = 'snippet-item';
                    snippetElement.innerHTML = `
                        <div class="snippet-header">
                            <strong>${snippet.trigger}</strong> - ${snippet.name}
                        </div>
                        <div class="snippet-description">${snippet.description || ''}</div>
                        <div class="snippet-template"><pre>${snippet.template}</pre></div>
                    `;
                    snippetsList.appendChild(snippetElement);
                });
            })
            .catch((err) => {
                snippetsList.innerHTML = `<div class="error">Error loading snippets: ${err}</div>`;
                console.error(err);
            });
    } catch (err) {
        snippetsList.innerHTML = `<div class="error">Error: ${err}</div>`;
        console.error(err);
    }
};
