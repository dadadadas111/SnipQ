package vault

import (
	"errors"
)

// Common errors
var (
	ErrVaultNotLoaded   = errors.New("vault not loaded")
	ErrInvalidPath      = errors.New("invalid vault path")
	ErrSnippetNotFound  = errors.New("snippet not found")
	ErrGroupNotFound    = errors.New("group not found")
	ErrDuplicateTrigger = errors.New("duplicate trigger")
	ErrDuplicateSnippet = errors.New("duplicate snippet")
	ErrDuplicateGroup   = errors.New("duplicate group")
	ErrInvalidSnippet   = errors.New("invalid snippet")
	ErrInvalidGroup     = errors.New("invalid group")
)

// Vault constants
const (
	SettingsFileName = "settings.yaml"
	CountersFileName = "counters.json"
	HistoryFileName  = "history.jsonl"
	GroupsDir        = "groups"
	SnippetsDir      = "snippets"
	GroupFileName    = "group.yaml"

	DefaultHistoryLimit = 200
	MaxHistoryLimit     = 10000
)
