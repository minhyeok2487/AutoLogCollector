package cisco

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CommandBlock represents a command and its output
type CommandBlock struct {
	Command string
	Lines   []string // Including the command line itself
}

// ExportToExcel exports execution results to an Excel file
// Format: One sheet per command, Columns = Hostnames, Rows = Output lines
func ExportToExcel(results []ExecutionResult, commands []string, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Define styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1a73e8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	lineNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "888888"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"f5f5f5"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cmdStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF9C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cmdLineNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "888888"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF9C4"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Parse outputs by command for each server
	serverCommandBlocks := make([][]CommandBlock, len(results))
	for i, result := range results {
		serverCommandBlocks[i] = splitOutputByCommands(result.Output, commands)
	}

	// Create a sheet for each command
	for cmdIdx, cmd := range commands {
		sheetName := sanitizeSheetName(cmd, cmdIdx+1)

		// Create sheet (first one replaces Sheet1)
		if cmdIdx == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}

		// Write header row
		f.SetCellValue(sheetName, "A1", "Line")
		f.SetCellStyle(sheetName, "A1", "A1", headerStyle)

		for i, result := range results {
			col := getColumnName(i + 2)
			f.SetCellValue(sheetName, col+"1", result.Server.Hostname)
			f.SetCellStyle(sheetName, col+"1", col+"1", headerStyle)
		}

		// Get command blocks for this command from each server
		commandOutputs := make([][]string, len(results))
		maxLines := 0

		for serverIdx := range results {
			blocks := serverCommandBlocks[serverIdx]
			for _, block := range blocks {
				if block.Command == cmd {
					commandOutputs[serverIdx] = block.Lines
					if len(block.Lines) > maxLines {
						maxLines = len(block.Lines)
					}
					break
				}
			}
		}

		// Write data rows
		for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
			excelRow := lineIdx + 2

			// Line number
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", excelRow), lineIdx+1)

			// First line (command line) gets special style
			if lineIdx == 0 {
				f.SetCellStyle(sheetName, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), cmdLineNumStyle)
			} else {
				f.SetCellStyle(sheetName, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), lineNumStyle)
			}

			// Each server's line
			for serverIdx := range results {
				col := getColumnName(serverIdx + 2)
				cell := fmt.Sprintf("%s%d", col, excelRow)

				if lineIdx < len(commandOutputs[serverIdx]) {
					f.SetCellValue(sheetName, cell, commandOutputs[serverIdx][lineIdx])
				} else {
					f.SetCellValue(sheetName, cell, "")
				}

				// First line gets command style
				if lineIdx == 0 {
					f.SetCellStyle(sheetName, cell, cell, cmdStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, cellStyle)
				}
			}
		}

		// Set column widths
		f.SetColWidth(sheetName, "A", "A", 8)
		for i := range results {
			col := getColumnName(i + 2)
			f.SetColWidth(sheetName, col, col, 60)
		}
	}

	return f.SaveAs(outputPath)
}

// splitOutputByCommands parses output and splits it by command blocks
func splitOutputByCommands(output string, commands []string) []CommandBlock {
	lines := strings.Split(output, "\n")
	blocks := make([]CommandBlock, 0)

	var currentBlock *CommandBlock

	for _, line := range lines {
		cleanLine := strings.TrimRight(line, "\r")

		// Check if this line is a command prompt line (e.g., "Router#show run")
		// Command line must have prompt character (#, >) followed by the command
		matchedCmd := ""
		for _, cmd := range commands {
			if isCommandLine(cleanLine, cmd) {
				matchedCmd = cmd
				break
			}
		}

		if matchedCmd != "" {
			// Start a new block
			if currentBlock != nil {
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CommandBlock{
				Command: matchedCmd,
				Lines:   []string{cleanLine},
			}
		} else if currentBlock != nil {
			// Add to current block
			currentBlock.Lines = append(currentBlock.Lines, cleanLine)
		}
	}

	// Don't forget the last block
	if currentBlock != nil {
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// isCommandLine checks if the line is a command prompt line
// Cisco prompts look like: "hostname#command" or "hostname>command"
func isCommandLine(line, command string) bool {
	// Find prompt character position
	hashPos := strings.LastIndex(line, "#")
	gtPos := strings.LastIndex(line, ">")

	promptPos := hashPos
	if gtPos > promptPos {
		promptPos = gtPos
	}

	if promptPos == -1 {
		return false
	}

	// Check if command appears right after the prompt character
	afterPrompt := line[promptPos+1:]
	afterPrompt = strings.TrimLeft(afterPrompt, " ") // Remove leading spaces

	return strings.HasPrefix(afterPrompt, command)
}

// sanitizeSheetName creates a valid Excel sheet name from a command
func sanitizeSheetName(cmd string, index int) string {
	// Remove invalid characters for Excel sheet names
	invalid := []string{":", "\\", "/", "?", "*", "[", "]"}
	name := cmd
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Excel sheet name max length is 31 characters
	if len(name) > 28 {
		name = name[:28]
	}

	// Add index prefix to ensure uniqueness
	return fmt.Sprintf("%d_%s", index, name)
}

// getColumnName converts column index (1-based) to Excel column name
func getColumnName(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}
