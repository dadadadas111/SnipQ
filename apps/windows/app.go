package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snipq/core/pkg/core"
	"github.com/snipq/core/pkg/types"
)

// App struct
type App struct {
	ctx    context.Context
	engine *core.Engine
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		engine: core.NewEngine(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize the vault - use a default location for now
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	vaultPath := filepath.Join(homeDir, ".snipq", "vault")
	err = a.engine.OpenVault(vaultPath)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		// Try creating a new vault with sample data
		if err := a.createSampleVault(vaultPath); err != nil {
			fmt.Printf("Error creating sample vault: %v\n", err)
		}
	}
}

// createSampleVault creates a sample vault with demo snippets
func (a *App) createSampleVault(vaultPath string) error {
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		return err
	}

	if err := a.engine.OpenVault(vaultPath); err != nil {
		return err
	}

	// Create sample snippets - the vault will create group structure automatically
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

	for _, snippet := range snippets {
		if err := a.engine.UpsertSnippet(snippet); err != nil {
			fmt.Printf("Error creating snippet %s: %v\n", snippet.ID, err)
		}
	}

	return a.engine.Save()
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
