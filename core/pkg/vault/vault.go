package vault

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/snipq/core/pkg/types"
)

// Vault manages the file-based snippet vault
type Vault struct {
	path     string
	groups   map[string]*types.Group
	snippets map[string]*types.Snippet
	settings *types.Settings
	counters map[string]*types.Counter
	history  []*types.HistoryEntry
}

// NewVault creates a new vault instance
func NewVault() *Vault {
	return &Vault{
		groups:   make(map[string]*types.Group),
		snippets: make(map[string]*types.Snippet),
		counters: make(map[string]*types.Counter),
		history:  make([]*types.HistoryEntry, 0),
	}
}

// Load loads the vault from the specified path
func (v *Vault) Load(path string) error {
	v.path = path

	// Ensure vault directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// Load settings
	if err := v.loadSettings(); err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Load counters
	if err := v.loadCounters(); err != nil {
		return fmt.Errorf("failed to load counters: %w", err)
	}

	// Load history
	if err := v.loadHistory(); err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	// Load groups and snippets
	if err := v.loadGroups(); err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}

	return nil
}

// Save saves the vault to disk
func (v *Vault) Save() error {
	if v.path == "" {
		return fmt.Errorf("vault path not set")
	}

	// Save settings
	if err := v.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Save counters
	if err := v.saveCounters(); err != nil {
		return fmt.Errorf("failed to save counters: %w", err)
	}

	// Save history
	if err := v.saveHistory(); err != nil {
		return fmt.Errorf("failed to save history: %w", err)
	}

	return nil
}

// GetSettings returns the vault settings
func (v *Vault) GetSettings() *types.Settings {
	if v.settings == nil {
		// Return default settings
		return &types.Settings{
			Prefix:            ":",
			ExpandKey:         "Tab",
			StrictBoundaries:  true,
			Locale:            "en-US",
			DefaultDateFormat: "2006-01-02",
			Timezone:          "Local",
			HistoryEnabled:    true,
			HistoryLimit:      200,
			PinForSensitive:   true,
		}
	}
	return v.settings
}

// SaveSettings saves the settings
func (v *Vault) SaveSettings(settings *types.Settings) error {
	v.settings = settings
	return v.saveSettings()
}

// ListGroups returns all groups sorted by order
func (v *Vault) ListGroups() []*types.Group {
	groups := make([]*types.Group, 0, len(v.groups))
	for _, group := range v.groups {
		groups = append(groups, group)
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Order != groups[j].Order {
			return groups[i].Order < groups[j].Order
		}
		return groups[i].Name < groups[j].Name
	})

	return groups
}

// ListSnippets returns all snippets for a group
func (v *Vault) ListSnippets(groupID string) []*types.Snippet {
	snippets := make([]*types.Snippet, 0)
	for _, snippet := range v.snippets {
		if snippet.GroupID == groupID {
			snippets = append(snippets, snippet)
		}
	}

	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].Name < snippets[j].Name
	})

	return snippets
}

// FindSnippetByTrigger finds a snippet by its trigger
func (v *Vault) FindSnippetByTrigger(trigger string) *types.Snippet {
	for _, snippet := range v.snippets {
		if snippet.Trigger == trigger {
			return snippet
		}
	}
	return nil
}

// UpsertSnippet adds or updates a snippet
func (v *Vault) UpsertSnippet(snippet *types.Snippet) error {
	if err := ValidateSnippet(snippet); err != nil {
		return err
	}

	// Check for duplicate trigger in the same group (excluding current snippet)
	if err := v.checkDuplicateTrigger(snippet.Trigger, snippet.GroupID, snippet.ID); err != nil {
		return err
	}

	// Ensure group exists
	if _, exists := v.groups[snippet.GroupID]; !exists {
		return fmt.Errorf("%w: group '%s' does not exist", ErrInvalidGroup, snippet.GroupID)
	}

	v.snippets[snippet.ID] = snippet

	// Save snippet to file
	return v.saveSnippet(snippet)
}

// DeleteSnippet deletes a snippet
func (v *Vault) DeleteSnippet(id string) error {
	snippet, exists := v.snippets[id]
	if !exists {
		return fmt.Errorf("snippet not found: %s", id)
	}

	// Delete from memory
	delete(v.snippets, id)

	// Delete file
	snippetPath := filepath.Join(v.path, "groups", snippet.GroupID, "snippets", snippet.ID+".yaml")
	return os.Remove(snippetPath)
}

// GetCounter returns a counter value
func (v *Vault) GetCounter(name string) *types.Counter {
	return v.counters[name]
}

// UpdateCounter updates a counter value
func (v *Vault) UpdateCounter(name string, counter *types.Counter) error {
	v.counters[name] = counter
	return v.saveCounters()
}

// AddHistoryEntry adds an entry to the history
func (v *Vault) AddHistoryEntry(entry *types.HistoryEntry) error {
	if !v.GetSettings().HistoryEnabled {
		return nil
	}

	v.history = append(v.history, entry)

	// Trim history if it exceeds the limit
	limit := v.GetSettings().HistoryLimit
	if len(v.history) > limit {
		v.history = v.history[len(v.history)-limit:]
	}

	return v.saveHistory()
}

// Private methods

func (v *Vault) loadSettings() error {
	settingsPath := filepath.Join(v.path, "settings.yaml")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Use default settings
			v.settings = v.GetSettings()
			return nil
		}
		return err
	}

	var settings types.Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return err
	}

	v.settings = &settings
	return nil
}

func (v *Vault) saveSettings() error {
	settingsPath := filepath.Join(v.path, "settings.yaml")

	data, err := yaml.Marshal(v.settings)
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0600)
}

func (v *Vault) loadCounters() error {
	countersPath := filepath.Join(v.path, "counters.json")

	data, err := os.ReadFile(countersPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No counters file yet
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &v.counters)
}

func (v *Vault) saveCounters() error {
	countersPath := filepath.Join(v.path, "counters.json")

	data, err := json.MarshalIndent(v.counters, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(countersPath, data, 0600)
}

func (v *Vault) loadHistory() error {
	historyPath := filepath.Join(v.path, "history.jsonl")

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No history file yet
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry types.HistoryEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid entries
		}

		v.history = append(v.history, &entry)
	}

	return nil
}

func (v *Vault) saveHistory() error {
	historyPath := filepath.Join(v.path, "history.jsonl")

	var lines []string
	for _, entry := range v.history {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		lines = append(lines, string(data))
	}

	return os.WriteFile(historyPath, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

func (v *Vault) loadGroups() error {
	groupsDir := filepath.Join(v.path, "groups")

	if err := os.MkdirAll(groupsDir, 0755); err != nil {
		return err
	}

	return filepath.WalkDir(groupsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() != "groups" {
			return v.loadGroup(d.Name())
		}

		return nil
	})
}

func (v *Vault) loadGroup(groupID string) error {
	groupPath := filepath.Join(v.path, "groups", groupID, "group.yaml")

	data, err := os.ReadFile(groupPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default group
			group := &types.Group{
				ID:      groupID,
				Name:    groupID,
				Enabled: true,
			}
			v.groups[groupID] = group
			return v.saveGroup(group)
		}
		return err
	}

	var group types.Group
	if err := yaml.Unmarshal(data, &group); err != nil {
		return err
	}

	group.ID = groupID
	v.groups[groupID] = &group

	// Load snippets for this group
	return v.loadSnippetsForGroup(groupID)
}

func (v *Vault) saveGroup(group *types.Group) error {
	groupDir := filepath.Join(v.path, "groups", group.ID)
	if err := os.MkdirAll(groupDir, 0755); err != nil {
		return err
	}

	groupPath := filepath.Join(groupDir, "group.yaml")
	data, err := yaml.Marshal(group)
	if err != nil {
		return err
	}

	return os.WriteFile(groupPath, data, 0600)
}

func (v *Vault) loadSnippetsForGroup(groupID string) error {
	snippetsDir := filepath.Join(v.path, "groups", groupID, "snippets")

	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		return err
	}

	return filepath.WalkDir(snippetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".yaml") {
			return v.loadSnippet(path, groupID)
		}

		return nil
	})
}

func (v *Vault) loadSnippet(path, groupID string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var snippet types.Snippet
	if err := yaml.Unmarshal(data, &snippet); err != nil {
		return err
	}

	snippet.GroupID = groupID
	v.snippets[snippet.ID] = &snippet

	return nil
}

func (v *Vault) saveSnippet(snippet *types.Snippet) error {
	snippetDir := filepath.Join(v.path, "groups", snippet.GroupID, "snippets")
	if err := os.MkdirAll(snippetDir, 0755); err != nil {
		return err
	}

	snippetPath := filepath.Join(snippetDir, snippet.ID+".yaml")
	data, err := yaml.Marshal(snippet)
	if err != nil {
		return err
	}

	return os.WriteFile(snippetPath, data, 0600)
}

// checkDuplicateTrigger checks if a trigger already exists in the group
func (v *Vault) checkDuplicateTrigger(trigger, groupID, excludeSnippetID string) error {
	for _, snippet := range v.snippets {
		if snippet.GroupID == groupID && snippet.Trigger == trigger && snippet.ID != excludeSnippetID {
			return fmt.Errorf("%w: trigger '%s' already exists in group '%s' (snippet: %s)",
				ErrDuplicateTrigger, trigger, groupID, snippet.ID)
		}
	}
	return nil
}

// CreateGroup creates a new group
func (v *Vault) CreateGroup(group *types.Group) error {
	if err := ValidateGroup(group); err != nil {
		return err
	}

	// Check for duplicate ID
	if _, exists := v.groups[group.ID]; exists {
		return fmt.Errorf("%w: group with ID '%s' already exists", ErrDuplicateGroup, group.ID)
	}

	v.groups[group.ID] = group
	return v.saveGroup(group)
}

// UpsertGroup adds or updates a group
func (v *Vault) UpsertGroup(group *types.Group) error {
	if err := ValidateGroup(group); err != nil {
		return err
	}

	v.groups[group.ID] = group
	return v.saveGroup(group)
}

// DeleteGroup deletes a group and all its snippets
func (v *Vault) DeleteGroup(groupID string) error {
	if _, exists := v.groups[groupID]; !exists {
		return fmt.Errorf("%w: group '%s' not found", ErrInvalidGroup, groupID)
	}

	// Delete all snippets in the group
	for snippetID, snippet := range v.snippets {
		if snippet.GroupID == groupID {
			delete(v.snippets, snippetID)
		}
	}

	// Delete from memory
	delete(v.groups, groupID)

	// Delete directory
	groupDir := filepath.Join(v.path, "groups", groupID)
	return os.RemoveAll(groupDir)
}

// GetSnippet returns a snippet by ID
func (v *Vault) GetSnippet(id string) (*types.Snippet, error) {
	snippet, exists := v.snippets[id]
	if !exists {
		return nil, fmt.Errorf("snippet not found: %s", id)
	}
	return snippet, nil
}

// GetGroup returns a group by ID
func (v *Vault) GetGroup(id string) (*types.Group, error) {
	group, exists := v.groups[id]
	if !exists {
		return nil, fmt.Errorf("group not found: %s", id)
	}
	return group, nil
}

// ListAllSnippets returns all snippets across all groups
func (v *Vault) ListAllSnippets() []*types.Snippet {
	snippets := make([]*types.Snippet, 0, len(v.snippets))
	for _, snippet := range v.snippets {
		snippets = append(snippets, snippet)
	}

	sort.Slice(snippets, func(i, j int) bool {
		if snippets[i].GroupID != snippets[j].GroupID {
			return snippets[i].GroupID < snippets[j].GroupID
		}
		return snippets[i].Name < snippets[j].Name
	})

	return snippets
}

// SearchSnippets searches for snippets by name, trigger, or tags
func (v *Vault) SearchSnippets(query string) []*types.Snippet {
	query = strings.ToLower(query)
	var results []*types.Snippet

	for _, snippet := range v.snippets {
		// Search in name, trigger, and tags
		if strings.Contains(strings.ToLower(snippet.Name), query) ||
			strings.Contains(strings.ToLower(snippet.Trigger), query) ||
			v.searchInTags(snippet.Tags, query) {
			results = append(results, snippet)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// searchInTags checks if query matches any tag
func (v *Vault) searchInTags(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetHistory returns the expansion history
func (v *Vault) GetHistory() []*types.HistoryEntry {
	return v.history
}

// ClearHistory clears the expansion history
func (v *Vault) ClearHistory() error {
	v.history = make([]*types.HistoryEntry, 0)
	return v.saveHistory()
}
