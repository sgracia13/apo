// Azure Prod Ops CLI - A TUI for Azure DevOps
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/user/apo/internal/agent"
	"github.com/user/apo/internal/api"
	"github.com/user/apo/internal/config"
	"github.com/user/apo/internal/ui"
)

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		runTUI()
		return
	}

	switch os.Args[1] {
	case "ui", "tui":
		runTUI()
	case "config":
		runConfig()
	case "ask":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: apo ask <question>")
			os.Exit(1)
		}
		runAsk(strings.Join(os.Args[2:], " "))
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Printf("apo v%s - Azure Prod Ops CLI\n", version)
	default:
		runAsk(strings.Join(os.Args[1:], " "))
	}
}

func runTUI() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	app, err := ui.NewApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'apo config' to configure your connection.")
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runConfig() {
	cfg, _ := config.Load()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println()
	fmt.Println("🔧 Azure DevOps Configuration")
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("Organization [%s]: ", cfg.Organization)
	scanner.Scan()
	if input := strings.TrimSpace(scanner.Text()); input != "" {
		cfg.Organization = input
	}

	fmt.Printf("Project [%s]: ", cfg.Project)
	scanner.Scan()
	if input := strings.TrimSpace(scanner.Text()); input != "" {
		cfg.Project = input
	}

	fmt.Print("Personal Access Token (PAT): ")
	scanner.Scan()
	if input := strings.TrimSpace(scanner.Text()); input != "" {
		cfg.PAT = input
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Configuration saved to %s\n", config.GetConfigPath())

	fmt.Print("\nTesting connection... ")
	client := api.NewClient(cfg)
	projects, err := client.ListProjects()
	if err != nil {
		fmt.Printf("❌\n   %v\n", err)
		return
	}
	fmt.Printf("✅ Connected! Found %d project(s).\n", len(projects))
	fmt.Println("\nRun 'apo' to launch the TUI!")
}

func runAsk(query string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.ValidateWithProject(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'apo config' to configure your connection.")
		os.Exit(1)
	}

	client := api.NewClient(cfg)
	ag := agent.New(client)
	result := ag.Ask(query)

	if !result.Success {
		fmt.Printf("❌ %s\n", result.Message)
	} else {
		fmt.Printf("\n%s\n", result.Message)
	}

	if len(result.Suggestions) > 0 {
		fmt.Println()
		for _, s := range result.Suggestions {
			fmt.Printf("  💡 %s\n", s)
		}
	}

	if result.Data != nil {
		fmt.Println()
		fmt.Print(agent.FormatResult(result.Data))
	}
	fmt.Println()
}

func printHelp() {
	fmt.Printf(`
╔═══════════════════════════════════════════════════════════════╗
║              Azure Prod Ops CLI v%s                        ║
║              TUI for Azure DevOps                             ║
╚═══════════════════════════════════════════════════════════════╝

Usage:
  apo                   Launch the TUI (default)
  apo ask <question>    Ask a natural language question
  apo <question>        Ask a natural language question (shortcut)
  apo config            Configure Azure DevOps connection
  apo help              Show this help
  apo version           Show version

TUI Navigation:
  [1-5]       Switch between tabs
  [/]         Open Copilot mode
  [↑↓/jk]     Navigate items
  [g/G]       Go to top/bottom
  [Enter]     Open detail view
  [f]         Filter current list
  [Tab]       Cycle tabs
  [r]         Refresh data
  [Esc]       Back / Cancel
  [q]         Quit

Examples:
  apo "what work items are assigned to me?"
  apo "show failed builds"
  apo ask "list all pipelines"

Configuration:
  Config file: ~/.config/apo/config.json

  Environment variables (override config file):
    AZURE_DEVOPS_ORG
    AZURE_DEVOPS_PROJECT  
    AZURE_DEVOPS_PAT
`, version)
}
