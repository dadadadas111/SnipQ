package expansion

import (
	"fmt"
	"log"
	"time"

	"snipq-windows/internal/clipboard"
	"snipq-windows/internal/hook"
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
	engine           *core.Engine
	keyboardHook     *hook.KeyboardHook
	textInjector     *textinject.TextInjector
	clipboardManager *clipboard.ClipboardManager

	// Settings
	enabled         bool
	expansionMethod ExpansionMethod
	pauseOnFailure  bool

	// State
	lastTrigger     string
	lastTriggerTime time.Time
}

// NewExpansionManager creates a new expansion manager
func NewExpansionManager(engine *core.Engine) *ExpansionManager {
	em := &ExpansionManager{
		engine:           engine,
		keyboardHook:     hook.NewKeyboardHook(),
		textInjector:     textinject.NewTextInjector(),
		clipboardManager: clipboard.NewClipboardManager(),
		enabled:          true,
		expansionMethod:  MethodDirect,
		pauseOnFailure:   false,
	}

	// Set up trigger handler
	em.keyboardHook.SetTriggerHandler(em.handleTrigger)

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
	// Create trigger input
	input := types.TriggerInput{
		RawTrigger: trigger,
		Now:        time.Now(),
		AppID:      "", // TODO: Detect current application
	}

	// Expand using the core engine
	rendered, err := em.engine.Expand(input)
	if err != nil {
		return fmt.Errorf("failed to expand trigger: %w", err)
	}

	log.Printf("[EXPANSION] Expanded '%s' to '%s'", trigger, rendered.Output)

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
	return map[string]interface{}{
		"enabled":         em.enabled,
		"expansionMethod": em.expansionMethod,
		"hookRunning":     em.keyboardHook.IsRunning(),
		"lastTrigger":     em.lastTrigger,
		"lastTriggerTime": em.lastTriggerTime,
		"currentBuffer":   em.GetCurrentBuffer(),
	}
}
