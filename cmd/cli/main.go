package main

import (
	"flag"
	"fmt"
	"os"

	"cisco-plink/internal/cisco"
)

func main() {
	// Parse command line arguments
	concurrent := flag.Int("c", 5, "Number of concurrent connections")
	flag.Parse()

	fmt.Println("=== Cisco Device Automation Tool ===")
	fmt.Printf("Concurrent connections: %d\n", *concurrent)
	fmt.Println()

	// Load credentials
	creds, err := cisco.LoadCredentials("credentials.json")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Credentials loaded")

	// Load servers
	servers, err := cisco.LoadServers("config/servers.csv")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load servers: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] Loaded %d servers\n", len(servers))

	// Load commands
	commands, err := cisco.LoadCommands("config/commands.txt")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load commands: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] Loaded %d commands\n", len(commands))
	fmt.Println()

	// Create runner with concurrent connections
	runner := cisco.NewRunner(servers, commands, creds, *concurrent)

	// Set progress callback
	runner.OnProgress = func(current, total int, server cisco.Server, status string) {
		if status == "connecting" {
			fmt.Printf("[%d/%d] Connecting to %s (%s)... ", current, total, server.Hostname, server.IP)
		}
	}

	// Set result callback
	runner.OnResult = func(result cisco.ExecutionResult) {
		if result.Success {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("FAILED\n")
			fmt.Printf("        Error: %s\n", result.Error)
		}
	}

	// Start execution
	if err := runner.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start: %v\n", err)
		os.Exit(1)
	}

	// Wait for completion
	for runner.IsRunning() {
		// Busy wait (in CLI, this is acceptable)
	}

	// Print summary
	success, fail, total := runner.GetSummary()
	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed:  %d\n", fail)
	fmt.Printf("Total:   %d\n", total)
	fmt.Printf("Logs saved to: %s\n", runner.LogDir)
}
