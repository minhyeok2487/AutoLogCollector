package cisco

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Cisco prompt pattern: hostname# or hostname> or hostname(config)#
var promptPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+(\([^)]+\))?[#>]\s*$`)

// newSSHConfig creates SSH client config with legacy algorithm support for older Cisco devices
func newSSHConfig(creds *Credentials) *ssh.ClientConfig {
	return &ssh.ClientConfig{
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
				"diffie-hellman-group1-sha1",
			},
			Ciphers: []string{
				"aes128-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-cbc",
				"3des-cbc",
			},
		},
	}
}

// ExecuteCommands connects to server and executes commands with real-time log callback
func ExecuteCommands(server Server, creds *Credentials, commands []string, chunkTimeoutSec int, enableMode, disablePaging bool, onLog func(line string)) (string, error) {
	config := newSSHConfig(creds)

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
	var lineBuffer strings.Builder

	// Read output in background with line-based callback
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				outputChan <- chunk

				// Line-based parsing for callback
				if onLog != nil {
					lineBuffer.WriteString(chunk)
					content := lineBuffer.String()
					lines := strings.Split(content, "\n")

					// Call callback for complete lines (except the last incomplete one)
					for i := 0; i < len(lines)-1; i++ {
						line := strings.TrimRight(lines[i], "\r")
						if line != "" {
							onLog(line)
						}
					}

					// Keep the last incomplete line in buffer
					lineBuffer.Reset()
					lineBuffer.WriteString(lines[len(lines)-1])
				}
			}
			if err != nil {
				// Flush remaining buffer
				if onLog != nil && lineBuffer.Len() > 0 {
					remaining := strings.TrimRight(lineBuffer.String(), "\r")
					if remaining != "" {
						onLog(remaining)
					}
				}
				doneChan <- true
				return
			}
		}
	}()

	// Helper to check if line is a Cisco prompt
	isPrompt := func(line string) bool {
		line = strings.TrimSpace(line)
		if line == "" {
			return false
		}
		return promptPattern.MatchString(line)
	}

	// Chunk timeout duration (user configurable)
	chunkTimeout := time.Duration(chunkTimeoutSec) * time.Second

	// Helper to read with timeout, prompt detection, and --More-- handling
	readOutput := func(timeout time.Duration, detectPrompt bool) string {
		var result strings.Builder
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		for {
			select {
			case data := <-outputChan:
				result.WriteString(data)

				// Check for --More-- prompt and send space to continue
				if strings.Contains(data, "--More--") || strings.Contains(data, " --More-- ") {
					fmt.Fprint(stdin, " ") // Send space without newline to continue
					timer.Reset(chunkTimeout)
					continue
				}

				// Check for prompt if detection is enabled
				if detectPrompt {
					content := result.String()
					// Get the last non-empty line
					lines := strings.Split(content, "\n")
					for i := len(lines) - 1; i >= 0; i-- {
						line := strings.TrimRight(lines[i], "\r\n ")
						if line != "" {
							if isPrompt(line) {
								// Prompt detected - wait a bit for any trailing data
								time.Sleep(100 * time.Millisecond)
								return result.String()
							}
							break
						}
					}
				}

				// Reset timer - wait for more data (use configurable chunk timeout)
				timer.Reset(chunkTimeout)
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
	initialOutput := readOutput(3*time.Second, false)
	output.WriteString(initialOutput)

	// Enter enable mode (optional)
	if enableMode {
		sendCommand("enable")
		time.Sleep(500 * time.Millisecond)
		enableOutput := readOutput(2*time.Second, false)
		output.WriteString(enableOutput)

		// Send enable password
		sendCommand(creds.EnablePassword)
		time.Sleep(500 * time.Millisecond)
		passwordOutput := readOutput(2*time.Second, false)
		output.WriteString(passwordOutput)
	}

	// Disable paging to get full output (optional)
	if disablePaging {
		sendCommand("terminal length 0")
		time.Sleep(300 * time.Millisecond)
		readOutput(2*time.Second, false)
	}

	// Execute commands - wait for timeout (prompt detection disabled for reliability)
	for _, cmd := range commands {
		sendCommand(cmd)
		time.Sleep(300 * time.Millisecond)
		cmdOutput := readOutput(120*time.Second, false)
		output.WriteString(cmdOutput)
	}

	// Restore terminal length to default (only if paging was disabled)
	if disablePaging {
		sendCommand("terminal length 24")
		time.Sleep(300 * time.Millisecond)
		termOutput := readOutput(2*time.Second, false)
		output.WriteString(termOutput)
	}

	// Exit gracefully
	sendCommand("exit")
	time.Sleep(500 * time.Millisecond)

	// Drain any remaining output
	drainOutput := readOutput(2*time.Second, false)
	output.WriteString(drainOutput)

	return output.String(), nil
}

// SaveLog writes output to log file
func SaveLog(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
