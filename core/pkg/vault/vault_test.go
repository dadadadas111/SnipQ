package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snipq/core/pkg/types"
)

func TestVaultValidation(t *testing.T) {
	// Test snippet validation
	tests := []struct {
		name    string
		snippet *types.Snippet
		wantErr bool
	}{
		{
			name: "valid snippet",
			snippet: &types.Snippet{
				ID:       "test1",
				Name:     "Test Snippet",
				Trigger:  "test",
				Template: "Hello {{.name}}",
				GroupID:  "group1",
			},
			wantErr: false,
		},
		{
			name:    "nil snippet",
			snippet: nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			snippet: &types.Snippet{
				Name:     "Test",
				Trigger:  "test",
				Template: "Hello",
				GroupID:  "group1",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			snippet: &types.Snippet{
				ID:       "test1",
				Trigger:  "test",
				Template: "Hello",
				GroupID:  "group1",
			},
			wantErr: true,
		},
		{
			name: "trigger with whitespace",
			snippet: &types.Snippet{
				ID:       "test1",
				Name:     "Test",
				Trigger:  "test trigger",
				Template: "Hello",
				GroupID:  "group1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSnippet(tt.snippet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSnippet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVaultGroupValidation(t *testing.T) {
	tests := []struct {
		name    string
		group   *types.Group
		wantErr bool
	}{
		{
			name: "valid group",
			group: &types.Group{
				ID:   "group1",
				Name: "Test Group",
			},
			wantErr: false,
		},
		{
			name:    "nil group",
			group:   nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			group: &types.Group{
				Name: "Test Group",
			},
			wantErr: true,
		},
		{
			name: "ID with invalid characters",
			group: &types.Group{
				ID:   "group/with/slash",
				Name: "Test Group",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroup(tt.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVaultCRUDOperations(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "vault-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create vault
	vault := NewVault()
	if err := vault.Load(tempDir); err != nil {
		t.Fatal(err)
	}

	// Test group creation
	group := &types.Group{
		ID:      "test-group",
		Name:    "Test Group",
		Enabled: true,
	}

	if err := vault.UpsertGroup(group); err != nil {
		t.Errorf("UpsertGroup() error = %v", err)
	}

	// Test snippet creation
	snippet := &types.Snippet{
		ID:       "test-snippet",
		Name:     "Test Snippet",
		Trigger:  "test",
		Template: "Hello {{.name}}",
		GroupID:  "test-group",
		Tags:     []string{"test", "sample"},
	}

	if err := vault.UpsertSnippet(snippet); err != nil {
		t.Errorf("UpsertSnippet() error = %v", err)
	}

	// Test snippet retrieval
	retrieved, err := vault.GetSnippet("test-snippet")
	if err != nil {
		t.Errorf("GetSnippet() error = %v", err)
	}
	if retrieved.Name != snippet.Name {
		t.Errorf("GetSnippet() name = %v, want %v", retrieved.Name, snippet.Name)
	}

	// Test duplicate trigger validation
	duplicateSnippet := &types.Snippet{
		ID:       "duplicate-snippet",
		Name:     "Duplicate Snippet",
		Trigger:  "test", // Same trigger as above
		Template: "Duplicate",
		GroupID:  "test-group",
	}

	if err := vault.UpsertSnippet(duplicateSnippet); err == nil {
		t.Error("UpsertSnippet() should have failed with duplicate trigger")
	}

	// Test search functionality
	results := vault.SearchSnippets("test")
	if len(results) != 1 {
		t.Errorf("SearchSnippets() found %d results, want 1", len(results))
	}

	// Test snippet deletion
	if err := vault.DeleteSnippet("test-snippet"); err != nil {
		t.Errorf("DeleteSnippet() error = %v", err)
	}

	// Verify deletion
	if _, err := vault.GetSnippet("test-snippet"); err == nil {
		t.Error("GetSnippet() should have failed after deletion")
	}
}

func TestVaultBackupRestore(t *testing.T) {
	// Create temporary directories
	vaultDir, err := os.MkdirTemp("", "vault-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	backupDir, err := os.MkdirTemp("", "backup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(backupDir)

	// Create vault with data
	vault := NewVault()
	if err := vault.Load(vaultDir); err != nil {
		t.Fatal(err)
	}

	// Add test data
	group := &types.Group{
		ID:      "test-group",
		Name:    "Test Group",
		Enabled: true,
	}
	if err := vault.UpsertGroup(group); err != nil {
		t.Fatal(err)
	}

	snippet := &types.Snippet{
		ID:       "test-snippet",
		Name:     "Test Snippet",
		Trigger:  "test",
		Template: "Hello {{.name}}",
		GroupID:  "test-group",
	}
	if err := vault.UpsertSnippet(snippet); err != nil {
		t.Fatal(err)
	}

	// Create backup
	if err := vault.BackupVault(backupDir); err != nil {
		t.Errorf("BackupVault() error = %v", err)
	}

	// Verify backup files exist
	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "snipq_backup_*"))
	if err != nil || len(backupFiles) == 0 {
		t.Error("Backup directory not created")
	}

	// Check manifest exists
	manifestPath := filepath.Join(backupFiles[0], "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Backup manifest not created")
	}

	// Test restore (create new vault)
	restoreVault := NewVault()
	if err := restoreVault.Load(vaultDir); err != nil {
		t.Fatal(err)
	}

	if err := restoreVault.RestoreVault(backupFiles[0]); err != nil {
		t.Errorf("RestoreVault() error = %v", err)
	}

	// Verify restored data
	restoredSnippet, err := restoreVault.GetSnippet("test-snippet")
	if err != nil {
		t.Errorf("GetSnippet() after restore error = %v", err)
	}
	if restoredSnippet.Name != snippet.Name {
		t.Errorf("Restored snippet name = %v, want %v", restoredSnippet.Name, snippet.Name)
	}
}

func TestVaultPathValidation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "valid relative path",
			path:    "./test",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/tmp/test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVaultPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVaultSettingsValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings *types.Settings
		wantErr  bool
	}{
		{
			name: "valid settings",
			settings: &types.Settings{
				Prefix:       ":",
				HistoryLimit: 100,
			},
			wantErr: false,
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
		},
		{
			name: "empty prefix",
			settings: &types.Settings{
				Prefix:       "",
				HistoryLimit: 100,
			},
			wantErr: true,
		},
		{
			name: "negative history limit",
			settings: &types.Settings{
				Prefix:       ":",
				HistoryLimit: -1,
			},
			wantErr: true,
		},
		{
			name: "excessive history limit",
			settings: &types.Settings{
				Prefix:       ":",
				HistoryLimit: 20000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
