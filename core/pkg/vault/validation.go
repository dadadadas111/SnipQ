package vault

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/snipq/core/pkg/types"
)

// ValidateSnippet validates a snippet before saving
func ValidateSnippet(snippet *types.Snippet) error {
	if snippet == nil {
		return ErrInvalidSnippet
	}

	if strings.TrimSpace(snippet.ID) == "" {
		return fmt.Errorf("%w: ID cannot be empty", ErrInvalidSnippet)
	}

	if strings.TrimSpace(snippet.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidSnippet)
	}

	if strings.TrimSpace(snippet.Trigger) == "" {
		return fmt.Errorf("%w: trigger cannot be empty", ErrInvalidSnippet)
	}

	if strings.TrimSpace(snippet.Template) == "" {
		return fmt.Errorf("%w: template cannot be empty", ErrInvalidSnippet)
	}

	if strings.TrimSpace(snippet.GroupID) == "" {
		return fmt.Errorf("%w: group ID cannot be empty", ErrInvalidSnippet)
	}

	// Validate trigger format
	if strings.ContainsAny(snippet.Trigger, " \t\n\r") {
		return fmt.Errorf("%w: trigger cannot contain whitespace", ErrInvalidSnippet)
	}

	return nil
}

// ValidateGroup validates a group before saving
func ValidateGroup(group *types.Group) error {
	if group == nil {
		return ErrInvalidGroup
	}

	if strings.TrimSpace(group.ID) == "" {
		return fmt.Errorf("%w: ID cannot be empty", ErrInvalidGroup)
	}

	if strings.TrimSpace(group.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidGroup)
	}

	// Validate ID format (no special characters that could cause file system issues)
	if strings.ContainsAny(group.ID, " \t\n\r/\\:*?\"<>|") {
		return fmt.Errorf("%w: ID contains invalid characters", ErrInvalidGroup)
	}

	return nil
}

// ValidateSettings validates settings before saving
func ValidateSettings(settings *types.Settings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	if strings.TrimSpace(settings.Prefix) == "" {
		return fmt.Errorf("prefix cannot be empty")
	}

	if settings.HistoryLimit < 0 {
		return fmt.Errorf("history limit cannot be negative")
	}

	if settings.HistoryLimit > MaxHistoryLimit {
		return fmt.Errorf("history limit cannot exceed %d", MaxHistoryLimit)
	}

	return nil
}

// ValidateVaultPath validates a vault path
func ValidateVaultPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrInvalidPath
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}

	// Check if path is valid
	if !filepath.IsAbs(absPath) {
		return fmt.Errorf("%w: path must be absolute", ErrInvalidPath)
	}

	return nil
}
