package cisco

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
)

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
