package expansion

import (
	"fmt"
	"log"
	"strings"
	"time"

	"snipq-windows/internal/clipboard"
	"snipq-windows/internal/hook"
	"snipq-windows/internal/suggestions"
	"snipq-windows/internal/textinject"

	"github.com/snipq/core/pkg/core"
	"github.com/snipq/core/pkg/types"
)

// ExpansionMethod represents different methods of text injection
type ExpansionMethod int

const (
	MethodDirect ExpansionMethod = iota
	MethodClipboard
)

// ExpansionManager manages snippet expansion through keyboard hooks
type ExpansionManager struct {
	engine            *core.Engine
	keyboardHook      *hook.KeyboardHook
	textInjector      *textinject.TextInjector
	clipboardManager  *clipboard.ClipboardManager
	suggestionManager *suggestions.SuggestionManager

	// Settings
	enabled         bool
	expansionMethod ExpansionMethod
	pauseOnFailure  bool
	currentSettings types.Settings // Store current settings

	// State
	lastTrigger     string
	lastTriggerTime time.Time
}

// NewExpansionManager creates a new expansion manager
func NewExpansionManager(engine *core.Engine) *ExpansionManager {
	em := &ExpansionManager{
		engine:            engine,
		keyboardHook:      hook.NewKeyboardHook(),
		textInjector:      textinject.NewTextInjector(),
		clipboardManager:  clipboard.NewClipboardManager(),
		suggestionManager: suggestions.NewSuggestionManager(engine),
		enabled:           true,
		expansionMethod:   MethodDirect,
		pauseOnFailure:    false,
	}

	// Set up trigger handler
	em.keyboardHook.SetTriggerHandler(em.handleTrigger)

	// Set up buffer handler for suggestions
	em.keyboardHook.SetBufferHandler(em.handleBufferUpdate)

	// Set up navigation handler for suggestion control
	em.keyboardHook.SetNavigationHandler(em.handleNavigation)

	// Set up suggestion acceptance handler for Tab key
	em.keyboardHook.SetSuggestionAcceptHandler(em.handleSuggestionAccept)

	// Set up suggestion selection handler
	em.suggestionManager.SetSelectionHandler(em.handleSuggestionSelection)

	return em
}

// Start starts the expansion manager
func (em *ExpansionManager) Start() error {
	log.Println("[EXPANSION] Starting expansion manager")

	// Load settings from engine
	settings, err := em.engine.GetSettings()
	if err != nil {
		log.Printf("[EXPANSION] Warning: Failed to load settings: %v", err)
		// Continue with defaults
	} else {
		em.keyboardHook.UpdateSettings(settings)
	}

	// Start keyboard hook
	err = em.keyboardHook.Start()
	if err != nil {
		return fmt.Errorf("failed to start keyboard hook: %w", err)
	}

	log.Println("[EXPANSION] Expansion manager started successfully")
	return nil
}

// handleBufferUpdate handles buffer changes from the keyboard hook
func (em *ExpansionManager) handleBufferUpdate(buffer string) {
	// Update suggestions based on buffer content
	em.suggestionManager.UpdateQuery(buffer)
}

// handleSuggestionSelection handles when a suggestion is selected
func (em *ExpansionManager) handleSuggestionSelection(suggestion suggestions.SuggestionItem) {
	log.Printf("[EXPANSION] Suggestion selected: %s", suggestion.Trigger)

	// Get current buffer to determine what user actually typed
	currentBuffer := em.GetCurrentBuffer()

	// Expand the selected trigger with smart deletion
	err := em.expandSuggestion(suggestion.Trigger, currentBuffer)
	if err != nil {
		log.Printf("[EXPANSION] Failed to expand selected suggestion '%s': %v", suggestion.Trigger, err)
	}
}

// handleNavigation handles navigation events from the keyboard hook
func (em *ExpansionManager) handleNavigation(direction string) bool {
	// Only handle navigation if suggestions are enabled and visible
	if !em.suggestionManager.IsEnabled() {
		return false
	}

	switch direction {
	case "up":
		em.suggestionManager.SelectPrevious()
		return true
	case "down":
		em.suggestionManager.SelectNext()
		return true
	case "accept":
		return em.suggestionManager.AcceptSelected()
	}

	return false
}

// handleSuggestionAccept handles suggestion acceptance from Tab key
func (em *ExpansionManager) handleSuggestionAccept() bool {
	// Only handle if suggestions are enabled and visible
	if !em.suggestionManager.IsEnabled() {
		return false
	}

	// Check if there's a current selection to accept
	selectedSuggestion := em.suggestionManager.SelectCurrent()
	if selectedSuggestion == nil {
		return false
	}

	// Accept the selected suggestion
	accepted := em.suggestionManager.AcceptSelected()
	if accepted {
		log.Printf("[EXPANSION] Tab key accepted suggestion: %s", selectedSuggestion.Trigger)
	}

	return accepted
}

// Stop stops the expansion manager
func (em *ExpansionManager) Stop() {
	log.Println("[EXPANSION] Stopping expansion manager")

	em.keyboardHook.Stop()

	log.Println("[EXPANSION] Expansion manager stopped")
}

// IsRunning returns whether the expansion manager is running
func (em *ExpansionManager) IsRunning() bool {
	return em.keyboardHook.IsRunning()
}

// SetEnabled enables or disables expansion
func (em *ExpansionManager) SetEnabled(enabled bool) {
	em.enabled = enabled

	// Also enable/disable suggestions
	em.suggestionManager.SetEnabled(enabled)

	log.Printf("[EXPANSION] Expansion %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// IsEnabled returns whether expansion is enabled
func (em *ExpansionManager) IsEnabled() bool {
	return em.enabled
}

// SetExpansionMethod sets the text injection method
func (em *ExpansionManager) SetExpansionMethod(method ExpansionMethod) {
	em.expansionMethod = method
	log.Printf("[EXPANSION] Expansion method set to: %v", method)
}

// GetExpansionMethod returns the current expansion method
func (em *ExpansionManager) GetExpansionMethod() ExpansionMethod {
	return em.expansionMethod
}

// UpdateSettings updates the expansion settings
func (em *ExpansionManager) UpdateSettings() error {
	settings, err := em.engine.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	em.keyboardHook.UpdateSettings(settings)
	log.Println("[EXPANSION] Settings updated")
	return nil
}

// UpdateSettingsWithValues updates the expansion settings with provided values
func (em *ExpansionManager) UpdateSettingsWithValues(settings types.Settings) error {
	em.currentSettings = settings // Store settings
	em.keyboardHook.UpdateSettings(settings)
	em.suggestionManager.UpdateSettings(settings) // Update suggestion manager settings
	log.Printf("[EXPANSION] Settings updated with values: prefix=%s, expandKey=%s, strictBoundaries=%t",
		settings.Prefix, settings.ExpandKey, settings.StrictBoundaries)
	return nil
}

// handleTrigger handles trigger detection from the keyboard hook
func (em *ExpansionManager) handleTrigger(trigger string) {
	log.Printf("[EXPANSION] Trigger detected: %s", trigger)

	// Check if expansion is enabled
	if !em.enabled {
		log.Printf("[EXPANSION] Expansion disabled, ignoring trigger: %s", trigger)
		return
	}

	// Store trigger info for debugging
	em.lastTrigger = trigger
	em.lastTriggerTime = time.Now()

	// Expand the trigger
	err := em.expandTrigger(trigger)
	if err != nil {
		log.Printf("[EXPANSION] Failed to expand trigger '%s': %v", trigger, err)

		if em.pauseOnFailure {
			em.SetEnabled(false)
			log.Println("[EXPANSION] Paused due to expansion failure")
		}
	}
}

// expandTrigger expands a trigger and injects the result
func (em *ExpansionManager) expandTrigger(trigger string) error {
	// Normalize trigger: convert user's prefix to standard ":" prefix for lookup
	normalizedTrigger := em.normalizeTrigger(trigger)

	// Create trigger input with normalized trigger
	input := types.TriggerInput{
		RawTrigger: normalizedTrigger,
		Now:        time.Now(),
		AppID:      "", // TODO: Detect current application
	}

	// Expand using the core engine
	rendered, err := em.engine.Expand(input)
	if err != nil {
		return fmt.Errorf("failed to expand trigger: %w", err)
	}

	log.Printf("[EXPANSION] Expanded '%s' (normalized: '%s') to '%s'", trigger, normalizedTrigger, rendered.Output)

	// Calculate how many characters to delete (trigger length)
	deleteCount := len(trigger)

	// Inject the expanded text
	switch em.expansionMethod {
	case MethodDirect:
		err = em.textInjector.ReplaceText(deleteCount, rendered.Output)
		if err != nil {
			return fmt.Errorf("failed to inject text directly: %w", err)
		}

	case MethodClipboard:
		err = em.injectViaClipboard(deleteCount, rendered.Output)
		if err != nil {
			return fmt.Errorf("failed to inject text via clipboard: %w", err)
		}

	default:
		return fmt.Errorf("unknown expansion method: %v", em.expansionMethod)
	}

	log.Printf("[EXPANSION] Successfully expanded and injected: %s -> %s", trigger, rendered.Output)
	return nil
}

// expandSuggestion expands a suggestion trigger with smart deletion based on what user actually typed
func (em *ExpansionManager) expandSuggestion(trigger string, currentBuffer string) error {
	// Normalize trigger: convert user's prefix to standard ":" prefix for lookup
	normalizedTrigger := em.normalizeTrigger(trigger)

	// Create trigger input with normalized trigger
	input := types.TriggerInput{
		RawTrigger: normalizedTrigger,
		Now:        time.Now(),
		AppID:      "", // TODO: Detect current application
	}

	// Expand using the core engine
	rendered, err := em.engine.Expand(input)
	if err != nil {
		return fmt.Errorf("failed to expand trigger: %w", err)
	}

	log.Printf("[EXPANSION] Expanded suggestion '%s' (normalized: '%s') to '%s'", trigger, normalizedTrigger, rendered.Output)

	// Calculate smart delete count based on what user actually typed
	deleteCount := em.calculateSmartDeleteCount(trigger, currentBuffer)

	log.Printf("[EXPANSION] Smart delete count: %d (trigger: '%s', buffer: '%s')", deleteCount, trigger, currentBuffer)

	// Inject the expanded text
	switch em.expansionMethod {
	case MethodDirect:
		err = em.textInjector.ReplaceText(deleteCount, rendered.Output)
		if err != nil {
			return fmt.Errorf("failed to inject text directly: %w", err)
		}

	case MethodClipboard:
		err = em.injectViaClipboard(deleteCount, rendered.Output)
		if err != nil {
			return fmt.Errorf("failed to inject text via clipboard: %w", err)
		}

	default:
		return fmt.Errorf("unknown expansion method: %v", em.expansionMethod)
	}

	log.Printf("[EXPANSION] Successfully expanded and injected suggestion: %s -> %s", trigger, rendered.Output)
	return nil
}

// calculateSmartDeleteCount calculates how many characters to delete based on what user actually typed
func (em *ExpansionManager) calculateSmartDeleteCount(fullTrigger string, currentBuffer string) int {
	// If buffer is empty, nothing to delete
	if currentBuffer == "" {
		return 0
	}

	// Find the last occurrence of the trigger prefix using current settings
	prefix := em.currentSettings.Prefix
	lastPrefixIndex := strings.LastIndex(currentBuffer, prefix)
	if lastPrefixIndex == -1 {
		// No prefix found, delete the entire buffer
		return len(currentBuffer)
	}

	// Extract what user actually typed from the last prefix to the end
	userTyped := currentBuffer[lastPrefixIndex:]

	// Ensure the user typed string is a prefix of the full trigger
	if strings.HasPrefix(fullTrigger, userTyped) {
		return len(userTyped)
	}

	// Fallback: if no match, try to find common prefix
	for i := 1; i <= len(userTyped) && i <= len(fullTrigger); i++ {
		if userTyped[len(userTyped)-i:] == fullTrigger[:i] {
			return len(userTyped)
		}
	}

	// Last resort: delete only the prefix if it exists
	if strings.HasSuffix(currentBuffer, prefix) {
		return len(prefix)
	}

	// No intelligent deletion possible, delete what user typed
	return len(userTyped)
}

// injectViaClipboard injects text using clipboard and Ctrl+V
func (em *ExpansionManager) injectViaClipboard(deleteCount int, text string) error {
	// Save current clipboard content
	originalClipboard, err := em.clipboardManager.GetText()
	if err != nil {
		log.Printf("[EXPANSION] Warning: Failed to save original clipboard: %v", err)
		originalClipboard = ""
	}

	// Set new text to clipboard
	err = em.clipboardManager.SetText(text)
	if err != nil {
		return fmt.Errorf("failed to set clipboard: %w", err)
	}

	// Delete the trigger characters
	if deleteCount > 0 {
		err = em.textInjector.SendBackspaces(deleteCount)
		if err != nil {
			return fmt.Errorf("failed to delete trigger: %w", err)
		}
	}

	// Paste the content
	err = em.textInjector.SendClipboardPaste()
	if err != nil {
		return fmt.Errorf("failed to paste: %w", err)
	}

	// Restore original clipboard after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		if originalClipboard != "" {
			err := em.clipboardManager.SetText(originalClipboard)
			if err != nil {
				log.Printf("[EXPANSION] Warning: Failed to restore clipboard: %v", err)
			}
		}
	}()

	return nil
}

// GetLastTrigger returns information about the last trigger
func (em *ExpansionManager) GetLastTrigger() (string, time.Time) {
	return em.lastTrigger, em.lastTriggerTime
}

// GetCurrentBuffer returns the current keyboard buffer (for debugging)
func (em *ExpansionManager) GetCurrentBuffer() string {
	return em.keyboardHook.GetCurrentBuffer()
}

// ClearBuffer clears the keyboard buffer
func (em *ExpansionManager) ClearBuffer() {
	em.keyboardHook.ClearBuffer()
}

// TestExpansion manually tests expansion without keyboard input
func (em *ExpansionManager) TestExpansion(trigger string) error {
	if !em.enabled {
		return fmt.Errorf("expansion is disabled")
	}

	return em.expandTrigger(trigger)
}

// PreviewExpansion previews what a trigger would expand to
func (em *ExpansionManager) PreviewExpansion(trigger string) (string, error) {
	input := types.TriggerInput{
		RawTrigger: trigger,
		Now:        time.Now(),
		AppID:      "",
	}

	return em.engine.Preview(input)
}

// GetStats returns expansion statistics
func (em *ExpansionManager) GetStats() map[string]interface{} {
	suggestionStats := em.suggestionManager.GetStats()

	return map[string]interface{}{
		"enabled":         em.enabled,
		"expansionMethod": em.expansionMethod,
		"hookRunning":     em.keyboardHook.IsRunning(),
		"lastTrigger":     em.lastTrigger,
		"lastTriggerTime": em.lastTriggerTime,
		"currentBuffer":   em.GetCurrentBuffer(),
		"suggestions":     suggestionStats,
	}
}

// Suggestion Control Methods

// IsSuggestionsEnabled returns whether suggestions are enabled
func (em *ExpansionManager) IsSuggestionsEnabled() bool {
	return em.suggestionManager.IsEnabled()
}

// SetSuggestionsEnabled enables or disables suggestions
func (em *ExpansionManager) SetSuggestionsEnabled(enabled bool) {
	em.suggestionManager.SetEnabled(enabled)
}

// SelectNextSuggestion selects the next suggestion
func (em *ExpansionManager) SelectNextSuggestion() {
	em.suggestionManager.SelectNext()
}

// SelectPreviousSuggestion selects the previous suggestion
func (em *ExpansionManager) SelectPreviousSuggestion() {
	em.suggestionManager.SelectPrevious()
}

// AcceptSelectedSuggestion accepts the currently selected suggestion
func (em *ExpansionManager) AcceptSelectedSuggestion() bool {
	return em.suggestionManager.AcceptSelected()
}

// normalizeTrigger converts user's custom prefix to standard ":" prefix for vault lookup
func (em *ExpansionManager) normalizeTrigger(trigger string) string {
	if em.currentSettings.Prefix == ":" {
		// Already using standard prefix, no normalization needed
		return trigger
	}

	// Replace user's prefix with standard ":" prefix
	if strings.HasPrefix(trigger, em.currentSettings.Prefix) {
		return ":" + strings.TrimPrefix(trigger, em.currentSettings.Prefix)
	}

	// If trigger doesn't start with current prefix, return as-is
	return trigger
}
