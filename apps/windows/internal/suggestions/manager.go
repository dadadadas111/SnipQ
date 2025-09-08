package suggestions

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/snipq/core/pkg/core"
	"github.com/snipq/core/pkg/types"
)

// SuggestionItem represents a single suggestion
type SuggestionItem struct {
	Trigger     string `json:"trigger"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GroupID     string `json:"groupId"`
	Preview     string `json:"preview"`
}

// SuggestionHandler is called when suggestions should be updated
type SuggestionHandler func(suggestions []SuggestionItem, query string)

// SelectionHandler is called when a suggestion is selected
type SelectionHandler func(suggestion SuggestionItem)

// SuggestionManager manages snippet suggestions
type SuggestionManager struct {
	engine *core.Engine
	popup  *SuggestionPopup
	mu     sync.RWMutex

	// Handlers
	suggestionHandler SuggestionHandler
	selectionHandler  SelectionHandler

	// External components for state updates
	keyboardHook interface {
		SetSuggestionsVisible(bool)
	}

	// Settings
	enabled         bool
	minQueryLength  int
	maxSuggestions  int
	showDelay       time.Duration
	hideDelay       time.Duration
	currentSettings types.Settings // Store current settings

	// State
	currentQuery       string
	currentSuggestions []SuggestionItem
	isVisible          bool
	selectedIndex      int
	lastUpdateTime     time.Time
	hideTimer          *time.Timer
}

// NewSuggestionManager creates a new suggestion manager
func NewSuggestionManager(engine *core.Engine) *SuggestionManager {
	sm := &SuggestionManager{
		engine:         engine,
		enabled:        true,
		minQueryLength: 1,
		maxSuggestions: 10,
		showDelay:      200 * time.Millisecond,
		hideDelay:      1000 * time.Millisecond,
		selectedIndex:  -1,
		// Initialize with default settings
		currentSettings: types.Settings{
			Prefix:           ":",
			ExpandKey:        "Tab",
			StrictBoundaries: true,
		},
	}

	// Create popup window
	sm.popup = NewSuggestionPopup()
	sm.popup.SetSelectionHandler(sm.handleSelection)

	return sm
}

// SetSuggestionHandler sets the handler for suggestion updates
func (sm *SuggestionManager) SetSuggestionHandler(handler SuggestionHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.suggestionHandler = handler
}

// SetSelectionHandler sets the handler for suggestion selection
func (sm *SuggestionManager) SetSelectionHandler(handler SelectionHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selectionHandler = handler
}

// SetKeyboardHook sets the keyboard hook for state synchronization
func (sm *SuggestionManager) SetKeyboardHook(hook interface{ SetSuggestionsVisible(bool) }) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.keyboardHook = hook
}

// SetEnabled enables or disables suggestions
func (sm *SuggestionManager) SetEnabled(enabled bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.enabled = enabled
	if !enabled {
		sm.hideSuggestions()
	}

	log.Printf("[SUGGESTIONS] Suggestions %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// UpdateSettings updates the suggestion settings with provided values
func (sm *SuggestionManager) UpdateSettings(settings types.Settings) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentSettings = settings
	log.Printf("[SUGGESTIONS] Settings updated with prefix=%s", settings.Prefix)
}

// IsEnabled returns whether suggestions are enabled
func (sm *SuggestionManager) IsEnabled() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.enabled
}

// UpdateQuery updates the current query and triggers suggestion updates
func (sm *SuggestionManager) UpdateQuery(query string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.enabled {
		return
	}

	// Clean the query - extract potential trigger
	cleanQuery := sm.extractTrigger(query)

	// Check if query changed
	if cleanQuery == sm.currentQuery {
		return
	}

	sm.currentQuery = cleanQuery
	sm.lastUpdateTime = time.Now()

	log.Printf("[SUGGESTIONS] Query updated: '%s' (from buffer: '%s')", cleanQuery, query)

	// Cancel existing hide timer
	if sm.hideTimer != nil {
		sm.hideTimer.Stop()
		sm.hideTimer = nil
	}

	// If query is too short, hide suggestions
	if len(cleanQuery) < sm.minQueryLength {
		sm.hideSuggestions()
		return
	}

	// Update suggestions after a delay to avoid flickering
	go func() {
		time.Sleep(sm.showDelay)
		sm.updateSuggestions(cleanQuery)
	}()
}

// extractTrigger extracts the trigger part from the buffer
func (sm *SuggestionManager) extractTrigger(buffer string) string {
	if buffer == "" {
		return ""
	}

	// Look for the last occurrence of the user's configured prefix
	lastPrefixIndex := strings.LastIndex(buffer, sm.currentSettings.Prefix)
	if lastPrefixIndex == -1 {
		return ""
	}

	// Extract everything from the last prefix to the end
	trigger := buffer[lastPrefixIndex:]

	// Check if this looks like a valid trigger (no spaces after prefix)
	if len(trigger) > 1 && strings.Contains(trigger[1:], " ") {
		return ""
	}

	return trigger
}

// updateSuggestions fetches and updates suggestions for the given query
func (sm *SuggestionManager) updateSuggestions(query string) {
	// Double-check that query is still current
	sm.mu.RLock()
	if query != sm.currentQuery || !sm.enabled {
		sm.mu.RUnlock()
		return
	}
	sm.mu.RUnlock()

	// Fetch suggestions
	suggestions := sm.fetchSuggestions(query)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check again in case query changed while fetching
	if query != sm.currentQuery {
		return
	}

	sm.currentSuggestions = suggestions

	// Auto-select first suggestion if available, otherwise no selection
	if len(suggestions) > 0 {
		sm.selectedIndex = 0 // Auto-select first item
		sm.showSuggestions(suggestions, query)
	} else {
		sm.selectedIndex = -1 // No selection when no suggestions
		sm.hideSuggestions()
	}
}

// fetchSuggestions retrieves matching snippets for the query
func (sm *SuggestionManager) fetchSuggestions(query string) []SuggestionItem {
	// Remove the prefix for matching using current settings
	searchQuery := strings.TrimPrefix(query, sm.currentSettings.Prefix)
	searchQuery = strings.ToLower(searchQuery)

	var suggestions []SuggestionItem

	// Get all groups
	groups, err := sm.engine.ListGroups()
	if err != nil {
		log.Printf("[SUGGESTIONS] Error listing groups: %v", err)
		return suggestions
	}

	// Search through all snippets
	for _, group := range groups {
		snippets, err := sm.engine.ListSnippets(group.ID)
		if err != nil {
			log.Printf("[SUGGESTIONS] Error listing snippets for group %s: %v", group.ID, err)
			continue
		}

		for _, snippet := range snippets {
			// Check if snippet matches query
			if sm.matchesQuery(snippet, searchQuery) {
				// Convert vault trigger to user's preferred prefix for display
				displayTrigger := sm.convertTriggerToUserPrefix(snippet.Trigger)

				// Generate preview
				preview, err := sm.generatePreview(snippet)
				if err != nil {
					preview = snippet.Template
				}

				suggestion := SuggestionItem{
					Trigger:     displayTrigger, // Show with user's preferred prefix
					Name:        snippet.Name,
					Description: snippet.Description,
					GroupID:     snippet.GroupID,
					Preview:     preview,
				}

				suggestions = append(suggestions, suggestion)

				// Limit number of suggestions
				if len(suggestions) >= sm.maxSuggestions {
					break
				}
			}
		}

		if len(suggestions) >= sm.maxSuggestions {
			break
		}
	}

	log.Printf("[SUGGESTIONS] Found %d suggestions for query '%s'", len(suggestions), query)
	return suggestions
}

// matchesQuery checks if a snippet matches the search query
func (sm *SuggestionManager) matchesQuery(snippet types.Snippet, query string) bool {
	if query == "" {
		return true
	}

	// Check trigger (normalize both snippet trigger and query - always use ":" for vault lookup)
	// Snippet triggers in vault use ":" prefix, so normalize them by removing ":"
	normalizedSnippetTrigger := strings.TrimPrefix(strings.ToLower(snippet.Trigger), ":")
	if strings.HasPrefix(normalizedSnippetTrigger, query) {
		return true
	}

	// Check name
	if strings.Contains(strings.ToLower(snippet.Name), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(snippet.Description), query) {
		return true
	}

	return false
}

// generatePreview generates a preview of the snippet
func (sm *SuggestionManager) generatePreview(snippet types.Snippet) (string, error) {
	input := types.TriggerInput{
		RawTrigger: snippet.Trigger,
		Now:        time.Now(),
		AppID:      "",
	}

	preview, err := sm.engine.Preview(input)
	if err != nil {
		return snippet.Template, err
	}

	// Limit preview length
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}

	return preview, nil
}

// showSuggestions displays the suggestions popup
func (sm *SuggestionManager) showSuggestions(suggestions []SuggestionItem, query string) {
	if !sm.isVisible {
		sm.isVisible = true
		sm.popup.Show(suggestions, query)

		// Notify keyboard hook that suggestions are now visible
		if sm.keyboardHook != nil {
			sm.keyboardHook.SetSuggestionsVisible(true)
		}

		// Set the current selection (if any)
		if sm.selectedIndex >= 0 && sm.selectedIndex < len(suggestions) {
			sm.popup.SetSelection(sm.selectedIndex)
		}

		// Call handler if set
		if sm.suggestionHandler != nil {
			go sm.suggestionHandler(suggestions, query)
		}
	} else {
		sm.popup.Update(suggestions, query)

		// Set the current selection (if any)
		if sm.selectedIndex >= 0 && sm.selectedIndex < len(suggestions) {
			sm.popup.SetSelection(sm.selectedIndex)
		}

		// Call handler if set
		if sm.suggestionHandler != nil {
			go sm.suggestionHandler(suggestions, query)
		}
	}
}

// hideSuggestions hides the suggestions popup
func (sm *SuggestionManager) hideSuggestions() {
	if sm.isVisible {
		sm.isVisible = false
		sm.popup.Hide()
		sm.currentSuggestions = nil
		sm.selectedIndex = -1

		// Notify keyboard hook that suggestions are now hidden
		if sm.keyboardHook != nil {
			sm.keyboardHook.SetSuggestionsVisible(false)
		}

		// Call handler if set
		if sm.suggestionHandler != nil {
			go sm.suggestionHandler(nil, "")
		}
	}
}

// SelectNext selects the next suggestion
func (sm *SuggestionManager) SelectNext() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isVisible || len(sm.currentSuggestions) == 0 {
		return
	}

	sm.selectedIndex++
	if sm.selectedIndex >= len(sm.currentSuggestions) {
		sm.selectedIndex = 0
	}

	sm.popup.SetSelection(sm.selectedIndex)
}

// SelectPrevious selects the previous suggestion
func (sm *SuggestionManager) SelectPrevious() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isVisible || len(sm.currentSuggestions) == 0 {
		return
	}

	sm.selectedIndex--
	if sm.selectedIndex < 0 {
		sm.selectedIndex = len(sm.currentSuggestions) - 1
	}

	sm.popup.SetSelection(sm.selectedIndex)
}

// SelectCurrent returns the currently selected suggestion
func (sm *SuggestionManager) SelectCurrent() *SuggestionItem {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.isVisible || sm.selectedIndex < 0 || sm.selectedIndex >= len(sm.currentSuggestions) {
		return nil
	}

	return &sm.currentSuggestions[sm.selectedIndex]
}

// AcceptSelected accepts the currently selected suggestion
func (sm *SuggestionManager) AcceptSelected() bool {
	suggestion := sm.SelectCurrent()
	if suggestion == nil {
		return false
	}

	sm.handleSelection(*suggestion)
	return true
}

// handleSelection handles when a suggestion is selected
func (sm *SuggestionManager) handleSelection(suggestion SuggestionItem) {
	log.Printf("[SUGGESTIONS] Selected suggestion: %s", suggestion.Trigger)

	sm.hideSuggestions()

	// Call handler if set
	sm.mu.RLock()
	handler := sm.selectionHandler
	sm.mu.RUnlock()

	if handler != nil {
		go handler(suggestion)
	}
}

// ScheduleHide schedules hiding the suggestions after a delay
func (sm *SuggestionManager) ScheduleHide() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Cancel existing timer
	if sm.hideTimer != nil {
		sm.hideTimer.Stop()
	}

	// Schedule hide
	sm.hideTimer = time.AfterFunc(sm.hideDelay, func() {
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.hideSuggestions()
	})
}

// CancelHide cancels any scheduled hide operation
func (sm *SuggestionManager) CancelHide() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.hideTimer != nil {
		sm.hideTimer.Stop()
		sm.hideTimer = nil
	}
}

// GetStats returns suggestion statistics
func (sm *SuggestionManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"enabled":         sm.enabled,
		"isVisible":       sm.isVisible,
		"currentQuery":    sm.currentQuery,
		"suggestionCount": len(sm.currentSuggestions),
		"selectedIndex":   sm.selectedIndex,
		"lastUpdateTime":  sm.lastUpdateTime,
	}
}

// Destroy cleans up the suggestion manager
func (sm *SuggestionManager) Destroy() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.hideTimer != nil {
		sm.hideTimer.Stop()
		sm.hideTimer = nil
	}

	sm.hideSuggestions()

	if sm.popup != nil {
		sm.popup.Destroy()
	}
}

// convertTriggerToUserPrefix converts vault trigger (with ":") to user's preferred prefix
func (sm *SuggestionManager) convertTriggerToUserPrefix(vaultTrigger string) string {
	if sm.currentSettings.Prefix == ":" {
		// Already using standard prefix, no conversion needed
		return vaultTrigger
	}

	// Replace ":" prefix with user's preferred prefix
	if strings.HasPrefix(vaultTrigger, ":") {
		return sm.currentSettings.Prefix + strings.TrimPrefix(vaultTrigger, ":")
	}

	// If trigger doesn't start with ":", return as-is
	return vaultTrigger
}
