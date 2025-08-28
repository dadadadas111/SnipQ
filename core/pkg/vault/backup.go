package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/snipq/core/pkg/types"
)

// BackupVault creates a backup of the vault data
func (v *Vault) BackupVault(backupDir string) error {
	if err := ValidateVaultPath(backupDir); err != nil {
		return fmt.Errorf("invalid backup directory: %w", err)
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("snipq_backup_%s", timestamp))

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create timestamped backup directory: %w", err)
	}

	// Backup snippets
	if err := v.backupSnippets(backupPath); err != nil {
		return fmt.Errorf("failed to backup snippets: %w", err)
	}

	// Backup groups
	if err := v.backupGroups(backupPath); err != nil {
		return fmt.Errorf("failed to backup groups: %w", err)
	}

	// Backup settings
	if err := v.backupSettings(backupPath); err != nil {
		return fmt.Errorf("failed to backup settings: %w", err)
	}

	// Backup counters
	if err := v.backupCounters(backupPath); err != nil {
		return fmt.Errorf("failed to backup counters: %w", err)
	}

	// Create backup manifest
	manifest := map[string]interface{}{
		"created_at":  time.Now().Format(time.RFC3339),
		"source_path": v.path,
		"backup_path": backupPath,
		"version":     "1.0",
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create backup manifest: %w", err)
	}

	manifestPath := filepath.Join(backupPath, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0600); err != nil {
		return fmt.Errorf("failed to write backup manifest: %w", err)
	}

	return nil
}

// RestoreVault restores vault data from a backup
func (v *Vault) RestoreVault(backupPath string) error {
	if err := ValidateVaultPath(backupPath); err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	// Check if backup exists
	manifestPath := filepath.Join(backupPath, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("backup manifest not found: %s", manifestPath)
	}

	// Read manifest
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read backup manifest: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse backup manifest: %w", err)
	}

	// Create backup of current state before restore
	backupDir := filepath.Join(v.path, "backups", "pre_restore")
	if err := v.BackupVault(backupDir); err != nil {
		return fmt.Errorf("failed to create pre-restore backup: %w", err)
	}

	// Restore files
	if err := v.restoreSnippets(backupPath); err != nil {
		return fmt.Errorf("failed to restore snippets: %w", err)
	}

	if err := v.restoreGroups(backupPath); err != nil {
		return fmt.Errorf("failed to restore groups: %w", err)
	}

	if err := v.restoreSettings(backupPath); err != nil {
		return fmt.Errorf("failed to restore settings: %w", err)
	}

	if err := v.restoreCounters(backupPath); err != nil {
		return fmt.Errorf("failed to restore counters: %w", err)
	}

	// Reload vault data
	return v.Load(v.path)
}

func (v *Vault) backupSnippets(backupPath string) error {
	snippetsPath := filepath.Join(backupPath, "snippets.json")
	snippetsData, err := json.MarshalIndent(v.snippets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snippetsPath, snippetsData, 0600)
}

func (v *Vault) backupGroups(backupPath string) error {
	groupsPath := filepath.Join(backupPath, "groups.json")
	groupsData, err := json.MarshalIndent(v.groups, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(groupsPath, groupsData, 0600)
}

func (v *Vault) backupSettings(backupPath string) error {
	settingsPath := filepath.Join(backupPath, "settings.yaml")
	settingsData, err := yaml.Marshal(v.settings)
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, settingsData, 0600)
}

func (v *Vault) backupCounters(backupPath string) error {
	countersPath := filepath.Join(backupPath, "counters.json")
	countersData, err := json.MarshalIndent(v.counters, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(countersPath, countersData, 0600)
}

func (v *Vault) restoreSnippets(backupPath string) error {
	snippetsPath := filepath.Join(backupPath, "snippets.json")
	data, err := os.ReadFile(snippetsPath)
	if err != nil {
		return err
	}

	var snippets map[string]*types.Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return err
	}

	v.snippets = snippets

	// Save individual snippets
	for _, snippet := range snippets {
		if err := v.saveSnippet(snippet); err != nil {
			return fmt.Errorf("failed to save snippet %s: %w", snippet.ID, err)
		}
	}

	return nil
}

func (v *Vault) restoreGroups(backupPath string) error {
	groupsPath := filepath.Join(backupPath, "groups.json")
	data, err := os.ReadFile(groupsPath)
	if err != nil {
		return err
	}

	var groups map[string]*types.Group
	if err := json.Unmarshal(data, &groups); err != nil {
		return err
	}

	v.groups = groups

	// Save individual groups
	for _, group := range groups {
		if err := v.saveGroup(group); err != nil {
			return fmt.Errorf("failed to save group %s: %w", group.ID, err)
		}
	}

	return nil
}

func (v *Vault) restoreSettings(backupPath string) error {
	settingsPath := filepath.Join(backupPath, "settings.yaml")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var settings types.Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return err
	}

	v.settings = &settings
	return v.saveSettings()
}

func (v *Vault) restoreCounters(backupPath string) error {
	countersPath := filepath.Join(backupPath, "counters.json")
	data, err := os.ReadFile(countersPath)
	if err != nil {
		return err
	}

	var counters map[string]*types.Counter
	if err := json.Unmarshal(data, &counters); err != nil {
		return err
	}

	v.counters = counters
	return v.saveCounters()
}
