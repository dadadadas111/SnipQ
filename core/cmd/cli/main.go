package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snipq/core/pkg/core"
	"github.com/snipq/core/pkg/types"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "expand":
		if len(os.Args) < 3 {
			fmt.Println("Usage: snipq expand <trigger>")
			os.Exit(1)
		}
		handleExpand(os.Args[2])
	case "preview":
		if len(os.Args) < 3 {
			fmt.Println("Usage: snipq preview <trigger>")
			os.Exit(1)
		}
		handlePreview(os.Args[2])
	case "list":
		handleList()
	case "init":
		handleInit()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SnipQ CLI Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  snipq expand <trigger>   - Expand a trigger")
	fmt.Println("  snipq preview <trigger>  - Preview expansion")
	fmt.Println("  snipq list              - List all snippets")
	fmt.Println("  snipq init              - Initialize sample vault")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  snipq expand ':ty'")
	fmt.Println("  snipq expand ':ty?lang=vi&tone=casual'")
	fmt.Println("  snipq expand ':date?format=Mon, 02 Jan 2006'")
	fmt.Println("  snipq preview ':uuid?upper=true'")
}

func getVaultPath() string {
	vaultPath := os.Getenv("SNIPQ_VAULT")
	if vaultPath == "" {
		// Use testdata vault as default
		wd, _ := os.Getwd()
		vaultPath = filepath.Join(wd, "internal", "testdata", "vault")
	}
	return vaultPath
}

func initEngine() (*core.Engine, error) {
	engine := core.NewEngine()
	vaultPath := getVaultPath()

	err := engine.OpenVault(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open vault at %s: %w", vaultPath, err)
	}

	return engine, nil
}

func handleExpand(trigger string) {
	engine, err := initEngine()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	input := types.TriggerInput{
		RawTrigger: trigger,
		Now:        time.Now(),
	}

	result, err := engine.Expand(input)
	if err != nil {
		fmt.Printf("Error expanding '%s': %v\n", trigger, err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", trigger)
	fmt.Printf("Output: %s\n", result.Output)
	fmt.Printf("Snippet: %s\n", result.UsedSnippet)

	if len(result.UsedParams) > 0 {
		fmt.Println("Parameters:")
		for key, value := range result.UsedParams {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

func handlePreview(trigger string) {
	engine, err := initEngine()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	input := types.TriggerInput{
		RawTrigger: trigger,
		Now:        time.Now(),
	}

	result, err := engine.Preview(input)
	if err != nil {
		fmt.Printf("Error previewing '%s': %v\n", trigger, err)
		os.Exit(1)
	}

	fmt.Printf("Preview: %s\n", result)
}

func handleList() {
	engine, err := initEngine()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	groups, err := engine.ListGroups()
	if err != nil {
		fmt.Printf("Error listing groups: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available Snippets:")
	fmt.Println("")

	for _, group := range groups {
		fmt.Printf("📁 %s (%s)\n", group.Name, group.ID)

		snippets, err := engine.ListSnippets(group.ID)
		if err != nil {
			fmt.Printf("  Error listing snippets: %v\n", err)
			continue
		}

		for _, snippet := range snippets {
			fmt.Printf("  ✨ %s - %s\n", snippet.Trigger, snippet.Name)
			if snippet.Description != "" {
				fmt.Printf("     %s\n", snippet.Description)
			}
		}
		fmt.Println("")
	}
}

func handleInit() {
	vaultPath := "./vault"

	fmt.Printf("Initializing vault at: %s\n", vaultPath)

	engine := core.NewEngine()
	err := engine.OpenVault(vaultPath)
	if err != nil {
		fmt.Printf("Error initializing vault: %v\n", err)
		os.Exit(1)
	}

	// Create a sample snippet
	snippet := types.Snippet{
		ID:          "snp_hello",
		Name:        "Hello World",
		Trigger:     ":hello",
		Description: "Simple hello world snippet",
		Template:    "Hello, World! Generated at {{ date \"15:04:05\" \"Local\" }}",
		GroupID:     "10-personal",
	}

	err = engine.UpsertSnippet(snippet)
	if err != nil {
		fmt.Printf("Error creating sample snippet: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Vault initialized successfully!")
	fmt.Println("")
	fmt.Println("Try:")
	fmt.Printf("  export SNIPQ_VAULT=%s\n", vaultPath)
	fmt.Println("  snipq expand ':hello'")
}
