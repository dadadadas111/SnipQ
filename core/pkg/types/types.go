package types

import (
	"time"
)

// TriggerInput represents a trigger with query parameters
type TriggerInput struct {
	RawTrigger string    // ":ty?lang=vi&tone=casual"
	AppID      string    // optional (per-app exclusions)
	Now        time.Time // testability
}

// Rendered represents the result of snippet expansion
type Rendered struct {
	Output       string         `json:"output"`
	CursorOffset int            `json:"cursorOffset"`
	UsedSnippet  string         `json:"usedSnippet"`
	UsedParams   map[string]any `json:"usedParams"`
}

// Group represents a snippet group
type Group struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Icon        string `yaml:"icon,omitempty" json:"icon,omitempty"`
	Order       int    `yaml:"order,omitempty" json:"order,omitempty"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
}

// Snippet represents a text snippet with template
type Snippet struct {
	ID          string         `yaml:"id" json:"id"`
	Name        string         `yaml:"name" json:"name"`
	Trigger     string         `yaml:"trigger" json:"trigger"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Strict      bool           `yaml:"strict,omitempty" json:"strict,omitempty"`
	Defaults    map[string]any `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Template    string         `yaml:"template" json:"template"`
	GroupID     string         `yaml:"-" json:"groupId"`
}

// Settings represents global vault settings
type Settings struct {
	Prefix            string   `yaml:"prefix" json:"prefix"`
	ExpandKey         string   `yaml:"expandKey" json:"expandKey"`
	StrictBoundaries  bool     `yaml:"strictBoundaries" json:"strictBoundaries"`
	ExcludedApps      []string `yaml:"excludedApps,omitempty" json:"excludedApps,omitempty"`
	Locale            string   `yaml:"locale" json:"locale"`
	DefaultDateFormat string   `yaml:"defaultDateFormat" json:"defaultDateFormat"`
	Timezone          string   `yaml:"timezone" json:"timezone"`
	HistoryEnabled    bool     `yaml:"historyEnabled" json:"historyEnabled"`
	HistoryLimit      int      `yaml:"historyLimit" json:"historyLimit"`
	PinForSensitive   bool     `yaml:"pinForSensitive" json:"pinForSensitive"`
}

// Counter represents a counter state
type Counter struct {
	Value     int       `json:"value"`
	Step      int       `json:"step"`
	Start     int       `json:"start"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CounterOpts represents options for counter operations
type CounterOpts struct {
	Pad  int `json:"pad,omitempty"`
	Step int `json:"step,omitempty"`
}

// HistoryEntry represents a snippet usage history entry
type HistoryEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	SnippetID  string         `json:"snippetId"`
	Output     string         `json:"output"`
	UsedParams map[string]any `json:"usedParams"`
	AppID      string         `json:"appId,omitempty"`
}
