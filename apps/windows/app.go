package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"snipq-windows/internal/expansion"
	"snipq-windows/internal/hotkey"

	"github.com/snipq/core/pkg/core"
	"github.com/snipq/core/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	engine           *core.Engine
	trayManager      *TrayManager
	hotkeyManager    *hotkey.Manager
	expansionManager *expansion.ExpansionManager
	isPaused         bool
	currentSettings  AppSettings
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		engine:        core.NewEngine(),
		hotkeyManager: hotkey.NewManager(),
		isPaused:      false,
		currentSettings: AppSettings{
			TriggerPrefix:       ":",
			ExpandKey:           "Tab",
			TypingTimeout:       2000,
			ExpansionMethod:     "direct",
			SuggestionsEnabled:  true,
			MinQueryLength:      1,
			MaxSuggestions:      10,
			SuggestionDelay:     200,
			SuggestionHideDelay: 1000,
			StrictBoundaries:    true,
			PauseOnFailure:      false,
			BufferSize:          100,
			CaseSensitive:       false,
		},
	}
	app.trayManager = NewTrayManager(app)
	app.expansionManager = expansion.NewExpansionManager(app.engine)

	// Set up hotkey event handler
	app.hotkeyManager.SetEventHandler(app.handleHotkeyEvent)

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize tray manager
	err := a.trayManager.Initialize(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize tray manager: %v\n", err)
	}

	// Try multiple vault locations
	vaultPaths := []string{}

	// 1. User home directory vault
	if homeDir, err := os.UserHomeDir(); err == nil {
		vaultPaths = append(vaultPaths, filepath.Join(homeDir, ".snipq", "vault"))
	}

	// 2. Fallback to testdata vault for demo
	if wd, err := os.Getwd(); err == nil {
		testdataVault := filepath.Join(wd, "..", "..", "core", "internal", "testdata", "vault")
		vaultPaths = append(vaultPaths, testdataVault)
	}

	var vaultLoaded bool

	for _, vaultPath := range vaultPaths {
		fmt.Printf("Trying vault path: %s\n", vaultPath)
		err := a.engine.OpenVault(vaultPath)
		if err == nil {
			fmt.Printf("Successfully loaded vault from: %s\n", vaultPath)
			vaultLoaded = true
			break
		} else {
			fmt.Printf("Error opening vault at %s: %v\n", vaultPath, err)
		}
	}

	// Check if vault is empty (no groups)
	if vaultLoaded {
		groups, err := a.engine.ListGroups()
		if err != nil || len(groups) == 0 {
			fmt.Printf("Vault loaded but empty (groups: %d). Creating sample data...\n", len(groups))
			vaultLoaded = false // Force sample creation
		} else {
			fmt.Printf("Vault loaded with %d groups\n", len(groups))
		}
	}

	// If no vault loaded or vault is empty, create a new one with sample data
	if !vaultLoaded {
		if homeDir, err := os.UserHomeDir(); err == nil {
			vaultPath := filepath.Join(homeDir, ".snipq", "vault")
			fmt.Printf("Creating new vault at: %s\n", vaultPath)
			if err := a.createSampleVault(vaultPath); err != nil {
				fmt.Printf("Error creating sample vault: %v\n", err)
			} else {
				fmt.Printf("Successfully created sample vault\n")
			}
		}
	}

	// Initialize hotkeys with defaults if none exist (after vault is loaded)
	settings, err := a.engine.GetSettings()
	if err != nil || settings.Hotkeys == nil {
		// Initialize with default hotkeys
		if settings.Hotkeys == nil {
			settings.Hotkeys = hotkey.GetDefaultHotkeys()
			fmt.Printf("Initializing default hotkeys\n")

			// Save the default hotkeys to settings
			err = a.engine.SaveSettings(settings)
			if err != nil {
				fmt.Printf("Failed to save default hotkeys: %v\n", err)
			} else {
				// Save vault to disk
				err = a.engine.Save()
				if err != nil {
					fmt.Printf("Failed to save vault with default hotkeys: %v\n", err)
				} else {
					fmt.Printf("Default hotkeys saved to vault\n")
				}
			}
		}
	}

	// Register hotkeys (initial registration)
	err = a.hotkeyManager.InitialRegisterHotkeys(settings.Hotkeys)
	if err != nil {
		fmt.Printf("Failed to register hotkeys: %v\n", err)
	}

	// Load and apply saved app settings
	savedAppSettings, err := a.loadAppSettings()
	if err != nil {
		fmt.Printf("Failed to load app settings: %v\n", err)
	} else {
		// Apply the loaded settings
		a.currentSettings = savedAppSettings
		fmt.Printf("Loaded app settings: suggestionsEnabled=%t, prefix=%s\n",
			savedAppSettings.SuggestionsEnabled, savedAppSettings.TriggerPrefix)
	}

	// Start hotkey listening
	err = a.hotkeyManager.StartListening()
	if err != nil {
		fmt.Printf("Failed to start hotkey listening: %v\n", err)
	}

	// Start expansion manager for automatic snippet expansion
	err = a.expansionManager.Start()
	if err != nil {
		fmt.Printf("Failed to start expansion manager: %v\n", err)
	} else {
		fmt.Printf("Expansion manager started successfully\n")

		// Apply saved app settings to expansion manager
		err = a.UpdateSettings(a.currentSettings)
		if err != nil {
			fmt.Printf("Failed to apply saved app settings: %v\n", err)
		} else {
			fmt.Printf("Applied saved app settings successfully\n")
		}
	}
}

// createSampleVault creates a sample vault with demo snippets
func (a *App) createSampleVault(vaultPath string) error {
	fmt.Printf("createSampleVault: Creating vault directory: %s\n", vaultPath)
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// First, create the group directory structure manually since our vault expects it
	groupPath := filepath.Join(vaultPath, "groups", "personal")
	fmt.Printf("createSampleVault: Creating group directory: %s\n", groupPath)
	if err := os.MkdirAll(filepath.Join(groupPath, "snippets"), 0755); err != nil {
		return fmt.Errorf("failed to create group directory: %w", err)
	}

	// Create group.yaml file
	groupYAML := `id: "personal"
name: "Personal"
description: "Personal snippets"
icon: "👤"
order: 10
enabled: true`

	groupYAMLPath := filepath.Join(groupPath, "group.yaml")
	fmt.Printf("createSampleVault: Creating group.yaml: %s\n", groupYAMLPath)
	if err := os.WriteFile(groupYAMLPath, []byte(groupYAML), 0644); err != nil {
		return fmt.Errorf("failed to create group.yaml: %w", err)
	}

	// Create settings.yaml
	settingsYAML := `prefix: ":"
expandKey: "Tab"
strictBoundaries: true
excludedApps: []
locale: "en-US"
defaultDateFormat: "2006-01-02"
timezone: "Local"
historyEnabled: true
historyLimit: 200
pinForSensitive: false`

	settingsPath := filepath.Join(vaultPath, "settings.yaml")
	fmt.Printf("createSampleVault: Creating settings.yaml: %s\n", settingsPath)
	if err := os.WriteFile(settingsPath, []byte(settingsYAML), 0644); err != nil {
		return fmt.Errorf("failed to create settings.yaml: %w", err)
	}

	// Create empty counters.json
	countersPath := filepath.Join(vaultPath, "counters.json")
	fmt.Printf("createSampleVault: Creating counters.json: %s\n", countersPath)
	if err := os.WriteFile(countersPath, []byte("{}"), 0644); err != nil {
		return fmt.Errorf("failed to create counters.json: %w", err)
	}

	// Now reload the vault to pick up the group structure
	fmt.Printf("createSampleVault: Reloading vault\n")
	if err := a.engine.OpenVault(vaultPath); err != nil {
		return fmt.Errorf("failed to reload vault: %w", err)
	}

	// Create sample snippets
	snippets := []types.Snippet{
		{
			ID:          "hello",
			Name:        "Hello World",
			Trigger:     ":hello",
			Description: "Simple hello world greeting",
			Template:    "Hello, World! 👋",
			GroupID:     "personal",
		},
		{
			ID:          "ty",
			Name:        "Thank You",
			Trigger:     ":ty",
			Description: "Thank you message",
			Template:    "Thank you! 😊",
			GroupID:     "personal",
		},
		{
			ID:          "date_today",
			Name:        "Current Date",
			Trigger:     ":today",
			Description: "Insert current date",
			Template:    "{{ date \"Monday, January 2, 2006\" \"Local\" }}",
			GroupID:     "personal",
		},
		{
			ID:          "today_test",
			Name:        "Today Test",
			Trigger:     ":todaytest",
			Description: "Test snippet for today prefix",
			Template:    "This is a test snippet for today: {{ date \"2006-01-02\" \"Local\" }}",
			GroupID:     "personal",
		},
		{
			ID:          "todo_item",
			Name:        "Todo Item",
			Trigger:     ":todo",
			Description: "Create a todo item",
			Template:    "- [ ] TODO: Add your task here",
			GroupID:     "personal",
		},
		{
			ID:          "email_sig",
			Name:        "Email Signature",
			Trigger:     ":sig",
			Description: "Professional email signature",
			Template:    "Best regards,\nSnipQ User\nsnipq@example.com",
			GroupID:     "personal",
		},
	}

	fmt.Printf("createSampleVault: Creating %d snippets\n", len(snippets))
	for _, snippet := range snippets {
		if err := a.engine.UpsertSnippet(snippet); err != nil {
			fmt.Printf("Error creating snippet %s: %v\n", snippet.ID, err)
			return fmt.Errorf("failed to create snippet %s: %w", snippet.ID, err)
		} else {
			fmt.Printf("Created snippet: %s (%s)\n", snippet.Trigger, snippet.Name)
		}
	}

	fmt.Printf("createSampleVault: Saving vault\n")
	if err := a.engine.Save(); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	fmt.Printf("createSampleVault: Successfully created sample vault with %d snippets\n", len(snippets))
	return nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetGroups returns all snippet groups
func (a *App) GetGroups() ([]types.Group, error) {
	return a.engine.ListGroups()
}

// GetSnippets returns all snippets for a group
func (a *App) GetSnippets(groupID string) ([]types.Snippet, error) {
	return a.engine.ListSnippets(groupID)
}

// ExpandSnippet expands a trigger and returns the result
func (a *App) ExpandSnippet(trigger string) (types.Rendered, error) {
	input := types.TriggerInput{
		RawTrigger: trigger,
	}
	return a.engine.Expand(input)
}

// PreviewSnippet previews a trigger expansion
func (a *App) PreviewSnippet(trigger string) (string, error) {
	input := types.TriggerInput{
		RawTrigger: trigger,
	}
	return a.engine.Preview(input)
}

// GetSettings returns the current settings
func (a *App) GetSettings() (types.Settings, error) {
	return a.engine.GetSettings()
}

// SaveSnippet saves a snippet
func (a *App) SaveSnippet(snippet types.Snippet) error {
	return a.engine.UpsertSnippet(snippet)
}

// DeleteSnippet deletes a snippet
func (a *App) DeleteSnippet(id string) error {
	return a.engine.DeleteSnippet(id)
}

// CreateSampleData creates sample data for testing
func (a *App) CreateSampleData() error {
	// Get vault path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	vaultPath := filepath.Join(homeDir, ".snipq", "vault")

	fmt.Printf("CreateSampleData: Removing existing vault at %s\n", vaultPath)
	// Remove existing vault
	os.RemoveAll(vaultPath)

	fmt.Printf("CreateSampleData: Creating fresh vault\n")
	// Create fresh vault with sample data
	err = a.createSampleVault(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create sample vault: %w", err)
	}

	fmt.Printf("CreateSampleData: Sample vault created successfully\n")
	return nil
}

// GetVaultInfo returns debug info about the vault
func (a *App) GetVaultInfo() map[string]interface{} {
	homeDir, _ := os.UserHomeDir()
	vaultPath := filepath.Join(homeDir, ".snipq", "vault")

	info := map[string]interface{}{
		"vaultPath": vaultPath,
		"exists":    false,
		"groups":    0,
		"snippets":  0,
	}

	if _, err := os.Stat(vaultPath); err == nil {
		info["exists"] = true
	}

	// Count groups and snippets
	groups, err := a.engine.ListGroups()
	if err == nil {
		info["groups"] = len(groups)

		totalSnippets := 0
		for _, group := range groups {
			snippets, err := a.engine.ListSnippets(group.ID)
			if err == nil {
				totalSnippets += len(snippets)
			}
		}
		info["snippets"] = totalSnippets
	}

	return info
}

// Tray Management API Methods

// ShowWindow shows the main window (callable from frontend)
func (a *App) ShowWindow() {
	a.trayManager.ShowWindow()
}

// HideWindow hides the main window (callable from frontend)
func (a *App) HideWindow() {
	a.trayManager.HideWindow()
}

// ToggleWindow toggles window visibility (callable from frontend)
func (a *App) ToggleWindow() {
	a.trayManager.ToggleWindow()
}

// ExitApp exits the application (callable from frontend)
func (a *App) ExitApp() {
	a.hotkeyManager.StopListening()
	a.hotkeyManager.UnregisterAll()
	a.expansionManager.Stop()
	a.trayManager.ExitApp()
}

// Hotkey Management API Methods

// handleHotkeyEvent handles hotkey activation events
func (a *App) handleHotkeyEvent(action string) {
	fmt.Printf("[APP] Hotkey event received: %s\n", action)

	switch action {
	case "toggleWindow":
		fmt.Printf("[APP] Toggling window visibility via hotkey\n")
		a.ToggleWindow()
		runtime.EventsEmit(a.ctx, "hotkey:toggleWindow")

	case "toggleExpansion":
		a.isPaused = !a.isPaused
		fmt.Printf("[APP] Toggling expansion via hotkey: paused=%t\n", a.isPaused)

		// Also pause/resume the expansion manager
		a.expansionManager.SetEnabled(!a.isPaused)

		runtime.EventsEmit(a.ctx, "hotkey:expansionToggled", map[string]interface{}{
			"paused": a.isPaused,
		})

	case "showSuggestions":
		fmt.Printf("[APP] Triggering suggestions via hotkey\n")
		runtime.EventsEmit(a.ctx, "hotkey:showSuggestions")

	case "quickExpand":
		fmt.Printf("[APP] Quick expand via hotkey\n")
		runtime.EventsEmit(a.ctx, "hotkey:quickExpand")

	case "focusSearch":
		fmt.Printf("[APP] Focus search via hotkey\n")
		a.ShowWindow() // Ensure window is visible first
		runtime.EventsEmit(a.ctx, "hotkey:focusSearch")

	default:
		fmt.Printf("[APP] Unknown hotkey action: %s\n", action)
	}
}

// GetHotkeyStatuses returns the current registration status of all hotkeys
func (a *App) GetHotkeyStatuses() map[string]*types.HotkeyStatus {
	return a.hotkeyManager.GetStatuses()
}

// GetHotkeys returns the current hotkey settings
func (a *App) GetHotkeys() map[string]*types.Hotkey {
	settings, err := a.engine.GetSettings()
	if err != nil || settings.Hotkeys == nil {
		return hotkey.GetDefaultHotkeys()
	}
	return settings.Hotkeys
}

// UpdateHotkeys updates hotkey settings and re-registers them
func (a *App) UpdateHotkeys(hotkeys map[string]*types.Hotkey) error {
	fmt.Printf("[APP] Updating hotkeys: %+v\n", hotkeys)

	// Get current settings
	settings, err := a.engine.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	// Update hotkeys in settings
	settings.Hotkeys = hotkeys

	// Save settings to vault
	err = a.engine.SaveSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Save vault to disk
	err = a.engine.Save()
	if err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	// Re-register hotkeys
	err = a.hotkeyManager.RegisterHotkeys(hotkeys)
	if err != nil {
		fmt.Printf("[APP] Failed to register hotkeys: %v\n", err)
		// Continue anyway since settings are saved
	}

	fmt.Printf("[APP] Hotkeys updated and saved successfully\n")
	return nil
}

// ValidateHotkey validates a hotkey configuration
func (a *App) ValidateHotkey(hk types.Hotkey) error {
	return hotkey.ValidateHotkey(hk)
}

// RefreshHotkeyStatuses refreshes the registration status of all hotkeys
func (a *App) RefreshHotkeyStatuses() map[string]*types.HotkeyStatus {
	fmt.Printf("[APP] Refreshing hotkey statuses\n")

	// Get current hotkeys
	hotkeys := a.GetHotkeys()

	// Re-register to get fresh status
	err := a.hotkeyManager.RegisterHotkeys(hotkeys)
	if err != nil {
		fmt.Printf("[APP] Error refreshing hotkeys: %v\n", err)
	}

	return a.hotkeyManager.GetStatuses()
}

// GetDefaultHotkeys returns the default hotkey configuration
func (a *App) GetDefaultHotkeys() map[string]*types.Hotkey {
	return hotkey.GetDefaultHotkeys()
}

// GetPauseState returns the current pause state
func (a *App) GetPauseState() bool {
	return a.isPaused
}

// TogglePause toggles the pause state manually (callable from UI)
func (a *App) TogglePause() bool {
	a.isPaused = !a.isPaused
	fmt.Printf("[APP] Toggling pause via UI: %t\n", a.isPaused)

	// Also pause/resume the expansion manager
	a.expansionManager.SetEnabled(!a.isPaused)

	runtime.EventsEmit(a.ctx, "hotkey:pauseToggled", map[string]interface{}{
		"paused": a.isPaused,
	})
	return a.isPaused
}

// Expansion Management API Methods

// GetExpansionStats returns statistics about the expansion manager
func (a *App) GetExpansionStats() map[string]interface{} {
	return a.expansionManager.GetStats()
}

// TestExpansion manually tests snippet expansion
func (a *App) TestExpansion(trigger string) error {
	return a.expansionManager.TestExpansion(trigger)
}

// GetExpansionBuffer returns the current keyboard buffer
func (a *App) GetExpansionBuffer() string {
	return a.expansionManager.GetCurrentBuffer()
}

// ClearExpansionBuffer clears the keyboard buffer
func (a *App) ClearExpansionBuffer() {
	a.expansionManager.ClearBuffer()
}

// SetExpansionEnabled enables or disables automatic expansion
func (a *App) SetExpansionEnabled(enabled bool) {
	a.expansionManager.SetEnabled(enabled)
}

// IsExpansionEnabled returns whether automatic expansion is enabled
func (a *App) IsExpansionEnabled() bool {
	return a.expansionManager.IsEnabled()
}

// Suggestion Management API Methods

// IsSuggestionsEnabled returns whether suggestions are enabled
func (a *App) IsSuggestionsEnabled() bool {
	return a.expansionManager.IsSuggestionsEnabled()
}

// SetSuggestionsEnabled enables or disables suggestions
func (a *App) SetSuggestionsEnabled(enabled bool) {
	a.expansionManager.SetSuggestionsEnabled(enabled)
}

// SelectNextSuggestion selects the next suggestion
func (a *App) SelectNextSuggestion() {
	a.expansionManager.SelectNextSuggestion()
}

// SelectPreviousSuggestion selects the previous suggestion
func (a *App) SelectPreviousSuggestion() {
	a.expansionManager.SelectPreviousSuggestion()
}

// AcceptSelectedSuggestion accepts the currently selected suggestion
func (a *App) AcceptSelectedSuggestion() bool {
	return a.expansionManager.AcceptSelectedSuggestion()
}

// Settings Management API Methods

// AppSettings represents all application settings
type AppSettings struct {
	// Basic expansion settings
	TriggerPrefix   string `json:"triggerPrefix"`
	ExpandKey       string `json:"expandKey"`
	TypingTimeout   int    `json:"typingTimeout"`   // milliseconds, 0 = no timeout
	ExpansionMethod string `json:"expansionMethod"` // "direct" or "clipboard"

	// Suggestions settings
	SuggestionsEnabled  bool `json:"suggestionsEnabled"`
	MinQueryLength      int  `json:"minQueryLength"`
	MaxSuggestions      int  `json:"maxSuggestions"`
	SuggestionDelay     int  `json:"suggestionDelay"`     // milliseconds
	SuggestionHideDelay int  `json:"suggestionHideDelay"` // milliseconds

	// Advanced settings
	StrictBoundaries bool `json:"strictBoundaries"`
	PauseOnFailure   bool `json:"pauseOnFailure"`
	BufferSize       int  `json:"bufferSize"`
	CaseSensitive    bool `json:"caseSensitive"`

	// Hotkeys (optional, can be updated separately)
	Hotkeys map[string]interface{} `json:"hotkeys,omitempty"`
}

// GetAppSettings returns the current application settings
func (a *App) GetAppSettings() AppSettings {
	return a.currentSettings
}

// UpdateSettings updates the application settings
func (a *App) UpdateSettings(settings AppSettings) error {
	log.Printf("Updating application settings: %+v", settings)

	// Validate settings first
	if err := a.validateSettings(settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Apply hotkeys if provided
	if len(settings.Hotkeys) > 0 {
		// Convert interface{} map to proper hotkey format
		hotkeys := make(map[string]*types.Hotkey)
		for action, hotkeyData := range settings.Hotkeys {
			if hotkeyMap, ok := hotkeyData.(map[string]interface{}); ok {
				hotkey := &types.Hotkey{}

				if modifiers, ok := hotkeyMap["modifiers"].([]interface{}); ok {
					for _, mod := range modifiers {
						if modStr, ok := mod.(string); ok {
							hotkey.Modifiers = append(hotkey.Modifiers, modStr)
						}
					}
				}

				if key, ok := hotkeyMap["key"].(string); ok {
					hotkey.Key = key
				}

				if enabled, ok := hotkeyMap["enabled"].(bool); ok {
					hotkey.Enabled = enabled
				}

				hotkeys[action] = hotkey
			}
		}

		// Update hotkeys
		err := a.UpdateHotkeys(hotkeys)
		if err != nil {
			log.Printf("Failed to update hotkeys: %v", err)
			return fmt.Errorf("failed to update hotkeys: %w", err)
		}
	}

	// Now update expansion manager settings by creating core settings
	if a.expansionManager != nil {
		// Create core settings with the new values
		coreSettings := types.Settings{
			Prefix:           settings.TriggerPrefix,
			ExpandKey:        settings.ExpandKey,
			StrictBoundaries: settings.StrictBoundaries,
			// Default values for other fields
			Locale:            "en-US",
			DefaultDateFormat: "2006-01-02",
			Timezone:          "Local",
			HistoryEnabled:    true,
			HistoryLimit:      200,
			PinForSensitive:   false,
		}

		// Update the expansion manager using existing UpdateSettings method
		// This will update the keyboard hook with the new prefix, expand key, etc.
		err := a.expansionManager.UpdateSettingsWithValues(coreSettings)
		if err != nil {
			log.Printf("Failed to update expansion manager settings: %v", err)
			return fmt.Errorf("failed to update expansion manager: %w", err)
		}

		// Update expansion method
		if settings.ExpansionMethod == "clipboard" {
			a.expansionManager.SetExpansionMethod(expansion.MethodClipboard)
		} else {
			a.expansionManager.SetExpansionMethod(expansion.MethodDirect)
		}

		log.Printf("Applied settings: prefix=%s, expandKey=%s, method=%s, strictBoundaries=%t",
			settings.TriggerPrefix, settings.ExpandKey, settings.ExpansionMethod, settings.StrictBoundaries)
	}

	// Apply suggestions settings
	a.SetSuggestionsEnabled(settings.SuggestionsEnabled)
	log.Printf("Applied suggestions enabled: %t", settings.SuggestionsEnabled)

	// Store the current settings
	a.currentSettings = settings

	// Save AppSettings to disk
	err := a.saveAppSettings(settings)
	if err != nil {
		log.Printf("Failed to save app settings: %v", err)
		return fmt.Errorf("failed to save app settings: %w", err)
	}

	log.Println("Settings updated and applied successfully")
	return nil
}

// GetDefaultSettings returns the default application settings
func (a *App) GetDefaultSettings() AppSettings {
	return AppSettings{
		TriggerPrefix:       ":",
		ExpandKey:           "Tab",
		TypingTimeout:       2000,
		ExpansionMethod:     "direct",
		SuggestionsEnabled:  true,
		MinQueryLength:      1,
		MaxSuggestions:      10,
		SuggestionDelay:     200,
		SuggestionHideDelay: 1000,
		StrictBoundaries:    true,
		PauseOnFailure:      false,
		BufferSize:          100,
		CaseSensitive:       false,
	}
}

// ExportSettings exports all settings as JSON
func (a *App) ExportSettings() (string, error) {
	log.Println("Exporting settings")

	settings := a.GetAppSettings()
	hotkeys := a.GetHotkeys()

	// Convert hotkeys to interface{} for JSON
	hotkeyMap := make(map[string]interface{})
	for action, hotkey := range hotkeys {
		hotkeyMap[action] = map[string]interface{}{
			"modifiers": hotkey.Modifiers,
			"key":       hotkey.Key,
			"enabled":   hotkey.Enabled,
		}
	}
	settings.Hotkeys = hotkeyMap

	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}

	return string(jsonData), nil
}

// ImportSettings imports settings from JSON
func (a *App) ImportSettings(jsonData string) error {
	log.Println("Importing settings")

	var settings AppSettings
	err := json.Unmarshal([]byte(jsonData), &settings)
	if err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// Validate settings
	err = a.validateSettings(settings)
	if err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Apply settings
	return a.UpdateSettings(settings)
}

// validateSettings validates the settings values
func (a *App) validateSettings(settings AppSettings) error {
	// Validate trigger prefix
	if len(settings.TriggerPrefix) != 1 {
		return fmt.Errorf("trigger prefix must be a single character")
	}

	// Validate expand key
	validExpandKeys := []string{"Tab", "Enter", "Space"}
	valid := false
	for _, key := range validExpandKeys {
		if settings.ExpandKey == key {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("expand key must be one of: %v", validExpandKeys)
	}

	// Validate ranges
	if settings.TypingTimeout < 0 || settings.TypingTimeout > 30000 {
		return fmt.Errorf("typing timeout must be between 0 and 30000 ms")
	}

	if settings.MinQueryLength < 1 || settings.MinQueryLength > 10 {
		return fmt.Errorf("minimum query length must be between 1 and 10")
	}

	if settings.MaxSuggestions < 1 || settings.MaxSuggestions > 50 {
		return fmt.Errorf("maximum suggestions must be between 1 and 50")
	}

	if settings.SuggestionDelay < 0 || settings.SuggestionDelay > 5000 {
		return fmt.Errorf("suggestion delay must be between 0 and 5000 ms")
	}

	if settings.BufferSize < 10 || settings.BufferSize > 1000 {
		return fmt.Errorf("buffer size must be between 10 and 1000")
	}

	return nil
}

// saveAppSettings saves the application settings to a file
func (a *App) saveAppSettings(settings AppSettings) error {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create snipq config directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".snipq")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save settings as JSON
	settingsPath := filepath.Join(configDir, "app-settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	log.Printf("App settings saved to: %s", settingsPath)
	return nil
}

// loadAppSettings loads the application settings from a file
func (a *App) loadAppSettings() (AppSettings, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return a.GetDefaultSettings(), fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Load settings from JSON file
	settingsPath := filepath.Join(homeDir, ".snipq", "app-settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("App settings file not found, using defaults")
			return a.GetDefaultSettings(), nil
		}
		return a.GetDefaultSettings(), fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		log.Printf("Failed to parse settings file, using defaults: %v", err)
		return a.GetDefaultSettings(), nil
	}

	log.Printf("App settings loaded from: %s", settingsPath)
	return settings, nil
}
