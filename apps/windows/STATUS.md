# SnipQ Windows App - Current Status & Future Roadmap

*Last Updated: September 4, 2025*

## 📊 **Current Status Overview**

### ✅ **COMPLETED FEATURES**

#### **Core Architecture**
- **Wails v2 Framework**: Modern desktop app with Go backend + JavaScript/HTML frontend
- **System Tray Integration**: Runs in background, accessible via tray icon
- **Vault Management**: File-based snippet storage in `~/.snipq/vault/`
- **Settings Persistence**: App settings saved to `~/.snipq/app-settings.json`
- **Core Engine Integration**: Uses shared Go core library for snippet processing

#### **Text Expansion Engine**
- **Global Keyboard Hook**: Monitors typing across all Windows applications
- **Real-time Expansion**: Automatic snippet expansion as user types
- **Smart Prefix Detection**: Configurable trigger prefix (`:`, `;`, `#`, etc.)
- **Multiple Expand Keys**: Tab, Enter, Space support
- **Text Injection Methods**: Direct input simulation + clipboard fallback
- **Prefix Normalization**: Users can customize prefix but snippets stored with `:` internally

#### **Snippet Management (Read-Only)**
- **Group-based Organization**: Snippets organized in groups
- **Template Rendering**: Dynamic templates with date, UUID, counters
- **Preview Functionality**: Test snippet expansion before use
- **Sample Data Creation**: Auto-creates demo snippets on first run
- **Vault Information Display**: Shows vault location and statistics

#### **Smart Suggestions System**
- **Real-time Suggestions**: Shows matching snippets as user types
- **Auto-selection**: First suggestion automatically selected
- **Keyboard Navigation**: Arrow keys to navigate suggestions
- **Multi-key Acceptance**: Both Tab and Enter accept suggestions
- **Dynamic Prefix Support**: Works with any user-configured prefix
- **Visual Feedback**: Popup window with snippet previews

#### **Hotkeys & Controls**
- **Global Hotkey System**: Customizable system-wide shortcuts
- **Pause/Resume**: Toggle expansion on/off globally
- **Buffer Management**: Clear keyboard buffer manually
- **App Activation**: Hotkey to bring app to front
- **Status Monitoring**: Real-time expansion status display

#### **Settings Management**
- **Comprehensive Settings UI**: All settings configurable via UI
- **Change Tracking**: Visual indicators for unsaved changes
- **Export/Import**: Settings backup and restore
- **Validation**: Input validation with error messages
- **Real-time Updates**: Settings applied immediately without restart

#### **User Interface**
- **Two-Tab Layout**: Main (snippets) + Settings tabs
- **Responsive Design**: Clean, modern interface
- **Real-time Status**: Expansion status, buffer content, statistics
- **Group Navigation**: Sidebar for browsing snippet groups
- **Search/Filter**: Test trigger functionality with instant preview

### 🏗️ **Technical Architecture**

#### **Backend Components**
```
/apps/windows/
├── app.go                 # Main Wails app controller
├── tray.go               # System tray management
├── internal/
│   ├── expansion/        # Text expansion engine
│   ├── hook/            # Keyboard hook management  
│   ├── hotkey/          # Global hotkey registration
│   ├── suggestions/     # Smart suggestions system
│   ├── textinject/      # Text injection methods
│   └── clipboard/       # Clipboard operations
```

#### **Frontend Stack**
- **Vanilla JavaScript**: No framework dependencies
- **CSS3**: Modern styling with flexbox/grid
- **Wails JS Bindings**: Direct Go function calls from frontend
- **Event System**: Real-time updates from backend

#### **Key Components**

##### **Expansion Manager** (`internal/expansion/manager.go`)
- Orchestrates text expansion workflow
- Manages keyboard hook and suggestion system
- Handles trigger normalization for custom prefixes
- Coordinates between keyboard input and snippet rendering

##### **Keyboard Hook** (`internal/hook/keyboard.go`)
- Low-level Windows keyboard monitoring
- Buffer management for typed characters
- Smart trigger detection and expansion key handling
- Navigation support for suggestions

##### **Suggestions Manager** (`internal/suggestions/manager.go`)
- Real-time suggestion display
- Auto-selection of first suggestion
- Keyboard navigation (up/down arrows)
- Dynamic prefix conversion for display

##### **Settings System** (`app.go`)
- App-level settings management (separate from core vault settings)
- Persistence to `~/.snipq/app-settings.json`
- Real-time application of settings changes
- Settings validation and error handling

#### **Data Flow**
1. **Keyboard Hook** → detects typing → **Expansion Manager**
2. **Expansion Manager** → processes triggers → **Core Engine**
3. **Core Engine** → renders templates → **Text Injector**
4. **Suggestions Manager** ← real-time buffer updates → **Frontend UI**

---

## 🚧 **MISSING FEATURES - Future Roadmap**

### 📝 **Phase 1: CRUD Operations (High Priority)**

#### **Snippet Management**
- [ ] **Create New Snippets**: Form-based snippet creation
  - Basic fields: ID, Name, Trigger, Description, Template
  - Group assignment
  - Template syntax highlighting
  - Live preview during creation
  
- [ ] **Edit Existing Snippets**: Inline editing capabilities
  - Edit all snippet properties
  - Validation and conflict checking
  - Backup/restore functionality
  
- [ ] **Delete Snippets**: Safe deletion with confirmations
  - Dependency checking
  - Bulk delete operations
  
- [ ] **Snippet Organization**: 
  - Drag & drop reordering
  - Bulk operations (move, copy, delete)
  - Search and filter snippets

#### **Group Management**
- [ ] **Create New Groups**: Group creation interface
  - Group metadata (name, description, icon)
  - Sorting and priority settings
  
- [ ] **Edit Groups**: Modify group properties
  - Rename groups
  - Change descriptions and icons
  - Reorganize group structure
  
- [ ] **Delete Groups**: Safe group removal
  - Handle snippets in deleted groups
  - Merge/move options before deletion

#### **Enhanced UI for CRUD**
- [ ] **Rich Text Editor**: For snippet templates
  - Syntax highlighting for template functions
  - Auto-complete for functions and variables
  - Real-time template validation
  
- [ ] **Form Validation**: Comprehensive input validation
  - Duplicate trigger detection
  - Template syntax validation
  - Required field enforcement

#### **Implementation Notes for Phase 1**
- **Backend API**: Extend `app.go` with CRUD methods
- **Core Integration**: Use existing core engine CRUD operations
- **UI Framework**: Consider migration to React/Vue for complex forms
- **Validation**: Server-side validation with client-side feedback

---

### ⚙️ **Phase 2: Advanced Snippet Features (Medium Priority)**

#### **Query-Style Options Support**
Currently: Simple snippets only (`":hello"` → `"Hello World"`)
Needed: Query parameters (`":date?format=YYYY-MM-DD&tz=UTC"`)

- [ ] **Query Parser Integration**: Support parameter parsing
- [ ] **Parameter UI**: Form fields for snippet parameters
- [ ] **Dynamic Previews**: Show parameter effects in real-time
- [ ] **Parameter Validation**: Type checking and validation
- [ ] **Default Values**: Smart defaults for common parameters

#### **Template Enhancement**
- [ ] **Advanced Functions**: More built-in template functions
  - Math operations
  - String manipulation
  - Conditional logic
  - Loops and iterations
  
- [ ] **Custom Variables**: User-defined template variables
- [ ] **Script Integration**: JavaScript/Lua scripting support
- [ ] **External Data**: API calls within templates

#### **Implementation Notes for Phase 2**
- **Core Engine**: Extend core library for query parameter support
- **Parser Enhancement**: Upgrade query parsing in core
- **UI Components**: Build parameter input forms
- **Template Engine**: Enhance template function library

---

### 🤖 **Phase 3: Smart Snippet Builder (Medium Priority)**

#### **AI-Assisted Creation**
- [ ] **Template Suggestions**: AI-powered template recommendations
- [ ] **Pattern Recognition**: Detect common snippet patterns
- [ ] **Smart Triggers**: Suggest optimal trigger names
- [ ] **Usage Analytics**: Recommend improvements based on usage

#### **Visual Template Builder**
- [ ] **Drag & Drop Interface**: Visual template construction
- [ ] **Component Library**: Pre-built template components
- [ ] **Logic Builder**: Visual if/then/else conditions
- [ ] **Preview Modes**: Multiple preview scenarios

#### **Import/Export Tools**
- [ ] **Snippet Packs**: Import/export snippet collections
- [ ] **Format Converters**: Import from other snippet tools
- [ ] **Sharing Templates**: Share templates with community

#### **Implementation Notes for Phase 3**
- **AI Integration**: Partner with AI providers (OpenAI, Claude)
- **Visual Editor**: Web-based drag-drop editor
- **Analytics Engine**: Track usage patterns locally
- **Community Platform**: Web platform for sharing

---

### ☁️ **Phase 4: Cloud Sync Integration (High Priority)**

#### **Authentication System**
- [ ] **Account Management**: Link to web account
  - User registration/login flow
  - API key generation and management
  - Account settings synchronization

#### **Sync Infrastructure**
- [ ] **Vault Synchronization**: 
  - Upload/download entire vault
  - Incremental sync (changed files only)
  - Conflict resolution strategies
  - Offline-first with sync when available

- [ ] **Settings Sync**:
  - Sync app settings across devices
  - Device-specific vs shared settings
  - Settings versioning and rollback

- [ ] **Multi-device Support**:
  - Device registration and management
  - Per-device customizations
  - Selective sync (groups/snippets)

#### **Cloud Features**
- [ ] **Backup & Restore**: Automatic cloud backups
- [ ] **Version History**: Track changes over time
- [ ] **Cross-platform Access**: Sync with browser extension, mobile
- [ ] **Sharing & Collaboration**: Share snippets/groups with others

#### **API Integration**
- [ ] **Sync API Client**: Integration with sync-api service
- [ ] **Authentication Flow**: OAuth-style API key workflow
- [ ] **Conflict Resolution**: Handle sync conflicts intelligently
- [ ] **Network Handling**: Offline support, retry logic, error handling

#### **Implementation Notes for Phase 4**
- **Sync Protocol**: Design for eventual real-time collaboration
- **Security**: Encrypted vault storage, API key scoping
- **Conflict Resolution**: Three-way merge strategies
- **User Experience**: Seamless background sync

#### **Cloud Sync Workflow**
1. **User Journey**: User visits website → creates account → generates API key → enters key in Windows app
2. **Technical Flow**: API key validation → device registration → vault upload → ongoing sync
3. **Security**: API keys with limited scope, encrypted vault storage
4. **Offline Support**: Full functionality without internet, sync when available

---

### 🔧 **Phase 5: Advanced Features (Lower Priority)**

#### **Enhanced User Experience**
- [ ] **Themes & Customization**: UI themes and layout options
- [ ] **Keyboard Shortcuts**: More granular hotkey controls
- [ ] **Accessibility**: Screen reader support, high contrast
- [ ] **Multi-language**: Localization support

#### **Integration Features**
- [ ] **App-specific Settings**: Per-application expansion rules
- [ ] **Context Awareness**: Location/time-based snippets
- [ ] **External Tool Integration**: IDE plugins, email clients
- [ ] **Automation**: Trigger snippets via external events

#### **Analytics & Optimization**
- [ ] **Usage Statistics**: Track snippet usage patterns
- [ ] **Performance Monitoring**: Expansion timing and reliability
- [ ] **Smart Suggestions**: ML-powered snippet recommendations
- [ ] **Adaptive Learning**: Learn from user behavior patterns

---

## 🏆 **Recent Achievements**

### **Smart Prefix System** (Completed)
- **Problem**: Users wanted custom prefixes (`;`, `#`) but vault stores snippets with `:`
- **Solution**: Prefix normalization - users type with custom prefix, system searches with `:`
- **Implementation**: 
  - `normalizeTrigger()` in expansion manager
  - `convertTriggerToUserPrefix()` in suggestions manager
  - Dynamic prefix detection in keyboard hook

### **Enhanced Suggestion System** (Completed)
- **Auto-selection**: First suggestion automatically highlighted
- **Smart Tab Handling**: Tab accepts suggestions OR expands complete triggers
- **Multi-key Support**: Both Tab and Enter accept suggestions
- **Real-time Updates**: Suggestions update as user types

### **Robust Settings Management** (Completed)
- **Persistent Storage**: App settings saved separately from vault settings
- **Change Tracking**: Visual indicators for unsaved changes
- **Immediate Application**: Settings applied without restart
- **Validation**: Input validation with user feedback

---

## 📊 **Technical Metrics**

### **Codebase Stats**
- **Total Files**: ~30 Go files, ~5 frontend files
- **Backend Components**: 6 major internal packages
- **Frontend**: ~1500 lines of vanilla JavaScript
- **Configuration**: Wails v2, Go 1.23

### **Performance Characteristics**
- **Startup Time**: < 2 seconds
- **Memory Usage**: ~50MB baseline
- **Keyboard Hook Latency**: < 5ms
- **Suggestion Display**: < 100ms

### **Supported Features**
- **Trigger Prefixes**: Any single character
- **Expand Keys**: Tab, Enter, Space
- **Text Injection**: Direct input + clipboard fallback
- **Template Functions**: Date, UUID, counters, clipboard
- **File Formats**: YAML vault, JSON settings

---

## 📋 **Development Priorities**

### **Immediate (Next 2-4 weeks)**
1. **CRUD Operations**: Essential for user productivity
2. **Form Validation**: Ensure data integrity
3. **Error Handling**: Improve user experience

### **Short Term (1-3 months)**
1. **Cloud Sync**: Critical for multi-device users
2. **Query Parameters**: Unlock advanced snippet capabilities
3. **UI Enhancement**: Modern, responsive interface

### **Long Term (3-6 months)**
1. **Smart Builder**: AI-assisted snippet creation
2. **Analytics**: Usage insights and optimization
3. **Community Features**: Sharing and collaboration

---

## 🔧 **Architecture Considerations**

### **Current Strengths**
- **Modular Design**: Clear separation of concerns
- **Event-Driven**: Reactive to user input
- **Cross-App Compatibility**: Works in any Windows application
- **Robust Settings**: Comprehensive configuration options

### **Areas for Improvement**
- **Database**: Consider SQLite for complex queries/indexing
- **UI Framework**: React/Vue for complex CRUD operations
- **Testing**: Unit and integration test coverage
- **Documentation**: API documentation and user guides

### **Scalability Considerations**
- **Large Vaults**: Optimize for thousands of snippets
- **Multi-User**: Prepare for team/organization features
- **Plugin System**: Architecture for third-party extensions
- **Performance**: Memory and CPU optimization

---

## 🎯 **Success Metrics**

### **User Experience**
- **Expansion Speed**: < 50ms from trigger to text
- **Suggestion Accuracy**: > 90% relevant first suggestions
- **Reliability**: > 99.5% successful expansions
- **Usability**: < 5 minutes to create first custom snippet

### **Technical Performance**
- **Memory Efficiency**: < 100MB with 1000+ snippets
- **Startup Performance**: < 3 seconds cold start
- **Sync Speed**: < 30 seconds full vault sync
- **Error Rate**: < 0.1% expansion failures

### **Adoption Goals**
- **User Retention**: > 80% monthly active users
- **Feature Usage**: > 60% using custom snippets
- **Sync Adoption**: > 40% using cloud sync
- **Community Engagement**: > 100 shared snippet packs

---

*This document serves as the definitive reference for SnipQ Windows app development status and planning. It will be updated as features are completed and priorities evolve.*
