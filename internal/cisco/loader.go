package cisco

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
)

// LoadCredentials reads credentials from JSON file
func LoadCredentials(path string) (*Credentials, error) {
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

// LoadServers reads server list from CSV file
func LoadServers(path string) ([]Server, error) {
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

// LoadCommands reads commands from text file
func LoadCommands(path string) ([]string, error) {
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
