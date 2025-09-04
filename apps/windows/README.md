# SnipQ Windows App

A powerful snippet manager for Windows that allows you to expand text snippets using customizable hotkeys and triggers.

## Features

- **Automatic Text Expansion**: Type triggers like `:hello` and expand them to full text snippets
- **Customizable Hotkeys**: Configure global hotkeys for opening the snippet palette, toggling pause, etc.
- **Real-time Expansion**: Snippets expand as you type in any application
- **Template Support**: Use dynamic templates with dates, variables, and more
- **Tray Integration**: Runs in the system tray for quick access
- **Cross-Application**: Works in any Windows application (Word, browser, IDE, etc.)

## Quick Start

### 1. First Launch
- Run `snipq-windows.exe`
- The app will create a sample vault with demo snippets
- The app will start in the system tray

### 2. Try Your First Snippet
- Open any text editor (Notepad, Word, browser, etc.)
- Type `:hello` and press **Tab**
- The trigger will be replaced with "Hello, World! 👋"

### 3. Available Sample Snippets
- `:hello` → "Hello, World! 👋"
- `:ty` → "Thank you! 😊"
- `:today` → Current date (e.g., "Wednesday, September 4, 2025")
- `:sig` → Email signature template

## How to Use

### Expanding Snippets
1. **Type a trigger**: Start with `:` followed by the snippet name (e.g., `:hello`)
2. **Press the expand key**: Default is **Tab** key
3. **Text replaces**: The trigger is replaced with the expanded snippet

### Managing Snippets
1. **Open the app**: Double-click the tray icon or right-click → "Show Window"
2. **Browse snippets**: Click on groups in the sidebar to see available snippets
3. **Test expansion**: Use the test box to try snippets before using them in other apps

### Hotkeys (Global)
- **Ctrl+Alt+Space**: Open snippet palette (brings app to front)
- **Ctrl+Alt+Period**: Toggle pause/resume automatic expansion
- **Ctrl+Alt+Enter**: Expand current trigger immediately (optional)

### Customizing Hotkeys
1. Open the app window
2. Go to the "Hotkeys" tab
3. Click "Change" next to any action to record a new hotkey
4. Click "Save Changes" to apply

### Pause/Resume
- **Pause**: Temporarily disable automatic expansion (hotkeys still work)
- **Resume**: Re-enable automatic expansion
- **Toggle via**: Hotkey (Ctrl+Alt+Period) or app UI

## Configuration

### Expansion Settings
The expansion engine monitors your typing and looks for trigger patterns. Key settings:

- **Prefix**: Default `:` (can be changed in settings)
- **Expand Key**: Default `Tab` (triggers expansion)
- **Strict Boundaries**: Ensures triggers only expand at word boundaries

### Text Injection Methods
The app uses two methods to inject expanded text:

1. **Direct Input** (Default): Simulates keystrokes directly
2. **Clipboard**: Uses clipboard + Ctrl+V (more compatible with some apps)

## Development

### Live Development
To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes.

### Building
To build a redistributable, production mode package, use `wails build`.

## Troubleshooting

### Snippets Not Expanding
1. **Check if paused**: Look for "Paused" status in the app or tray tooltip
2. **Check permissions**: Some applications may block input simulation
3. **Try clipboard method**: Change expansion method in settings
4. **Check trigger format**: Must start with `:` and match exactly

### Hotkeys Not Working
1. **Check registration status**: Go to Hotkeys tab to see registration status
2. **Try different combinations**: Some hotkeys may conflict with other applications
3. **Run as administrator**: May be needed for some system-level applications

### Application Compatibility
- **Works best with**: Most text editors, browsers, Office applications
- **May have issues with**: Some games, elevated applications, secure input fields
- **Alternative**: Use the app's test interface to copy text manually

## File Locations

- **Vault**: `%USERPROFILE%\.snipq\vault\`
- **Settings**: `%USERPROFILE%\.snipq\vault\settings.yaml`
- **Snippets**: `%USERPROFILE%\.snipq\vault\groups\[group-name]\snippets\`

## Known Limitations

1. **Admin applications**: May not work in applications running as administrator
2. **Secure input fields**: Password fields and secure forms may block expansion
3. **Some games**: Full-screen games may interfere with hotkeys
4. **Rich text**: Complex formatting may not be preserved
5. **Very long snippets**: Extremely long text may have timing issues

---

**Version**: 1.0.0  
**Platform**: Windows 10/11 (64-bit)
