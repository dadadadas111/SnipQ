package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/snipq/core/pkg/parser"
	"github.com/snipq/core/pkg/template"
	"github.com/snipq/core/pkg/types"
	"github.com/snipq/core/pkg/vault"
)

// Engine implements the Core interface
type Engine struct {
	vault    *vault.Vault
	template *template.Engine
}

// NewEngine creates a new core engine
func NewEngine() *Engine {
	return &Engine{
		vault:    vault.NewVault(),
		template: template.NewEngine(),
	}
}

// OpenVault opens a vault at the specified path
func (e *Engine) OpenVault(path string) error {
	return e.vault.Load(path)
}

// Reload reloads the vault from disk
func (e *Engine) Reload() error {
	// For now, just reload the vault
	// In the future, we might need to track the vault path
	return nil
}

// Save saves the vault to disk
func (e *Engine) Save() error {
	return e.vault.Save()
}

// Expand expands a trigger input into rendered output
func (e *Engine) Expand(input types.TriggerInput) (types.Rendered, error) {
	// Set default timestamp if not provided
	if input.Now.IsZero() {
		input.Now = time.Now()
	}

	// Parse the trigger
	parsed, err := parser.ParseTrigger(input.RawTrigger)
	if err != nil {
		return types.Rendered{}, fmt.Errorf("failed to parse trigger: %w", err)
	}

	// Find the snippet
	snippet := e.vault.FindSnippetByTrigger(parsed.Trigger)
	if snippet == nil {
		return types.Rendered{}, fmt.Errorf("snippet not found: %s", parsed.Trigger)
	}

	// Check if app is excluded
	settings := e.vault.GetSettings()
	if input.AppID != "" && e.isAppExcluded(input.AppID, settings.ExcludedApps) {
		return types.Rendered{}, fmt.Errorf("app excluded: %s", input.AppID)
	}

	// Merge parameters
	globalDefaults := e.getGlobalDefaults(settings)
	mergedParams := parser.MergeParams(parsed.Params, snippet.Defaults, globalDefaults)

	// Add special variables
	mergedParams["now"] = input.Now
	mergedParams["timestamp"] = input.Now.Unix()

	// Handle counters if the template uses them
	err = e.handleCounters(snippet.Template, mergedParams)
	if err != nil {
		return types.Rendered{}, fmt.Errorf("failed to handle counters: %w", err)
	}

	// Render the template
	output, err := e.template.Render(snippet.Template, mergedParams)
	if err != nil {
		return types.Rendered{}, fmt.Errorf("failed to render template: %w", err)
	}

	// Create rendered result
	rendered := types.Rendered{
		Output:       output,
		CursorOffset: 0, // TODO: Calculate cursor position from template
		UsedSnippet:  snippet.ID,
		UsedParams:   mergedParams,
	}

	// Add to history
	if settings.HistoryEnabled {
		historyEntry := &types.HistoryEntry{
			Timestamp:  input.Now,
			SnippetID:  snippet.ID,
			Output:     output,
			UsedParams: mergedParams,
			AppID:      input.AppID,
		}
		// Ignore history error in expansion context
		_ = e.vault.AddHistoryEntry(historyEntry)
	}

	return rendered, nil
}

// Preview previews a trigger expansion without side effects
func (e *Engine) Preview(input types.TriggerInput) (string, error) {
	// Set default timestamp if not provided
	if input.Now.IsZero() {
		input.Now = time.Now()
	}

	// Parse the trigger
	parsed, err := parser.ParseTrigger(input.RawTrigger)
	if err != nil {
		return "", fmt.Errorf("failed to parse trigger: %w", err)
	}

	// Find the snippet
	snippet := e.vault.FindSnippetByTrigger(parsed.Trigger)
	if snippet == nil {
		return "", fmt.Errorf("snippet not found: %s", parsed.Trigger)
	}

	// Merge parameters
	settings := e.vault.GetSettings()
	globalDefaults := e.getGlobalDefaults(settings)
	mergedParams := parser.MergeParams(parsed.Params, snippet.Defaults, globalDefaults)

	// Add special variables
	mergedParams["now"] = input.Now
	mergedParams["timestamp"] = input.Now.Unix()

	// Render the template (without side effects like updating counters)
	return e.template.Render(snippet.Template, mergedParams)
}

// ListGroups returns all groups
func (e *Engine) ListGroups() ([]types.Group, error) {
	vaultGroups := e.vault.ListGroups()
	groups := make([]types.Group, len(vaultGroups))

	for i, vg := range vaultGroups {
		groups[i] = *vg
	}

	return groups, nil
}

// ListSnippets returns all snippets for a group
func (e *Engine) ListSnippets(groupID string) ([]types.Snippet, error) {
	vaultSnippets := e.vault.ListSnippets(groupID)
	snippets := make([]types.Snippet, len(vaultSnippets))

	for i, vs := range vaultSnippets {
		snippets[i] = *vs
	}

	return snippets, nil
}

// UpsertSnippet adds or updates a snippet
func (e *Engine) UpsertSnippet(s types.Snippet) error {
	return e.vault.UpsertSnippet(&s)
}

// DeleteSnippet deletes a snippet
func (e *Engine) DeleteSnippet(id string) error {
	return e.vault.DeleteSnippet(id)
}

// GetSettings returns the vault settings
func (e *Engine) GetSettings() (types.Settings, error) {
	settings := e.vault.GetSettings()
	return *settings, nil
}

// SaveSettings saves the vault settings
func (e *Engine) SaveSettings(settings types.Settings) error {
	return e.vault.SaveSettings(&settings)
}

// NextCounter increments and returns the next counter value
func (e *Engine) NextCounter(name string, opts types.CounterOpts) (string, error) {
	counter := e.vault.GetCounter(name)
	if counter == nil {
		// Create new counter
		counter = &types.Counter{
			Value:     1,
			Step:      1,
			Start:     1,
			UpdatedAt: time.Now(),
		}
	}

	// Apply options
	step := counter.Step
	if opts.Step > 0 {
		step = opts.Step
	}

	// Update counter
	counter.Value += step
	counter.UpdatedAt = time.Now()

	// Save counter
	err := e.vault.UpdateCounter(name, counter)
	if err != nil {
		return "", err
	}

	// Format with padding if specified
	if opts.Pad > 0 {
		return fmt.Sprintf("%0*d", opts.Pad, counter.Value), nil
	}

	return strconv.Itoa(counter.Value), nil
}

// Private helper methods

func (e *Engine) isAppExcluded(appID string, excludedApps []string) bool {
	for _, excluded := range excludedApps {
		if excluded == appID {
			return true
		}
	}
	return false
}

func (e *Engine) getGlobalDefaults(settings *types.Settings) map[string]any {
	return map[string]any{
		"dateFormat": settings.DefaultDateFormat,
		"timezone":   settings.Timezone,
		"locale":     settings.Locale,
	}
}

func (e *Engine) handleCounters(templateText string, params map[string]any) error {
	// This is a simplified implementation
	// In a real implementation, we would parse the template to find counter calls
	// and pre-populate their values to avoid multiple increments during rendering
	return nil
}
