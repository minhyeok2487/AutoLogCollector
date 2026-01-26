package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Server represents a Cisco device
type Server struct {
	IP       string
	Hostname string
}

// Credentials holds SSH login information
type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func main() {
	fmt.Println("=== Cisco Device Automation Tool ===")
	fmt.Println()

	// Load credentials
	creds, err := loadCredentials("credentials.json")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Credentials loaded")

	// Load servers
	servers, err := loadServers("config/servers.csv")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load servers: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] Loaded %d servers\n", len(servers))

	// Load commands
	commands, err := loadCommands("config/commands.txt")
	if err != nil {
		fmt.Printf("[ERROR] Failed to load commands: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] Loaded %d commands\n", len(commands))
	fmt.Println()

	// Create log directory
	logDir := filepath.Join("logs", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("[ERROR] Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	// Process each server
	successCount := 0
	failCount := 0

	for i, server := range servers {
		fmt.Printf("[%d/%d] Connecting to %s (%s)... ", i+1, len(servers), server.Hostname, server.IP)

		output, err := executeCommands(server, creds, commands)
		if err != nil {
			fmt.Printf("FAILED\n")
			fmt.Printf("        Error: %v\n", err)
			failCount++
			continue
		}

		// Save log
		logPath := filepath.Join(logDir, server.Hostname+".log")
		if err := saveLog(logPath, output); err != nil {
			fmt.Printf("FAILED (log save)\n")
			fmt.Printf("        Error: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("OK\n")
		successCount++
	}

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Failed:  %d\n", failCount)
	fmt.Printf("Total:   %d\n", len(servers))
	fmt.Printf("Logs saved to: %s\n", logDir)
}

// loadCredentials reads credentials from JSON file
func loadCredentials(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// loadServers reads server list from CSV file
func loadServers(path string) ([]Server, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var servers []Server

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) >= 2 && strings.TrimSpace(record[0]) != "" {
			servers = append(servers, Server{
				IP:       strings.TrimSpace(record[0]),
				Hostname: strings.TrimSpace(record[1]),
			})
		}
	}

	return servers, nil
}

// loadCommands reads commands from text file
func loadCommands(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commands = append(commands, line)
		}
	}

	return commands, scanner.Err()
}

// executeCommands connects to server and executes commands
func executeCommands(server Server, creds *Credentials, commands []string) (string, error) {
	// SSH config with legacy algorithm support for older Cisco devices
	config := &ssh.ClientConfig{
		User: creds.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(creds.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
		Config: ssh.Config{
			KeyExchanges: []string{
				"curve25519-sha256",
				"curve25519-sha256@libssh.org",
				"ecdh-sha2-nistp256",
				"ecdh-sha2-nistp384",
				"ecdh-sha2-nistp521",
				"diffie-hellman-group14-sha256",
				"diffie-hellman-group14-sha1",
				"diffie-hellman-group1-sha1", // Legacy support
			},
			Ciphers: []string{
				"aes128-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-cbc", // Legacy support
				"3des-cbc",   // Legacy support
			},
		},
	}

	// Connect to SSH
	addr := net.JoinHostPort(server.IP, "22")
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("session creation failed: %v", err)
	}
	defer session.Close()

	// Set up pseudo-terminal for interactive session
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("vt100", 80, 200, modes); err != nil {
		return "", fmt.Errorf("PTY request failed: %v", err)
	}

	// Get stdin/stdout pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("stdin pipe failed: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe failed: %v", err)
	}

	// Start shell
	if err := session.Shell(); err != nil {
		return "", fmt.Errorf("shell start failed: %v", err)
	}

	var output strings.Builder
	outputChan := make(chan string)
	doneChan := make(chan bool)

	// Read output in background
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				outputChan <- string(buf[:n])
			}
			if err != nil {
				doneChan <- true
				return
			}
		}
	}()

	// Helper to read with timeout
	readOutput := func(timeout time.Duration) string {
		var result strings.Builder
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		for {
			select {
			case data := <-outputChan:
				result.WriteString(data)
				timer.Reset(200 * time.Millisecond)
			case <-timer.C:
				return result.String()
			case <-doneChan:
				return result.String()
			}
		}
	}

	// Helper to send command
	sendCommand := func(cmd string) {
		fmt.Fprintln(stdin, cmd)
	}

	// Wait for initial prompt
	initialOutput := readOutput(2 * time.Second)
	output.WriteString(initialOutput)

	// Enter enable mode
	sendCommand("enable")
	time.Sleep(500 * time.Millisecond)
	enableOutput := readOutput(1 * time.Second)
	output.WriteString(enableOutput)

	// Send enable password (same as login password)
	sendCommand(creds.Password)
	time.Sleep(500 * time.Millisecond)
	passwordOutput := readOutput(1 * time.Second)
	output.WriteString(passwordOutput)

	// Execute commands
	for _, cmd := range commands {
		sendCommand(cmd)
		time.Sleep(300 * time.Millisecond)
		cmdOutput := readOutput(3 * time.Second)
		output.WriteString(cmdOutput)
	}

	return output.String(), nil
}

// saveLog writes output to log file
func saveLog(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
