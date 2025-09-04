package main

import (
	"context"
	"fmt"
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
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		engine:        core.NewEngine(),
		hotkeyManager: hotkey.NewManager(),
		isPaused:      false,
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
	case "openPalette":
		fmt.Printf("[APP] Opening palette via hotkey\n")
		runtime.EventsEmit(a.ctx, "hotkey:openPalette")

	case "togglePause":
		a.isPaused = !a.isPaused
		fmt.Printf("[APP] Toggling pause via hotkey: %t\n", a.isPaused)

		// Also pause/resume the expansion manager
		a.expansionManager.SetEnabled(!a.isPaused)

		runtime.EventsEmit(a.ctx, "hotkey:pauseToggled", map[string]interface{}{
			"paused": a.isPaused,
		})

	case "expandNow":
		fmt.Printf("[APP] Expand now via hotkey\n")
		runtime.EventsEmit(a.ctx, "hotkey:expandNow")

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
