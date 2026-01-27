package cisco

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportToExcel exports execution results to an Excel file
func ExportToExcel(results []ExecutionResult, commands []string, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Results"
	f.SetSheetName("Sheet1", sheetName)

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1a73e8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set cell style
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Write header row
	f.SetCellValue(sheetName, "A1", "Hostname")
	f.SetCellStyle(sheetName, "A1", "A1", headerStyle)

	for i, cmd := range commands {
		col := getColumnName(i + 2) // B, C, D, ...
		f.SetCellValue(sheetName, col+"1", cmd)
		f.SetCellStyle(sheetName, col+"1", col+"1", headerStyle)
	}

	// Write data rows
	for rowIdx, result := range results {
		row := rowIdx + 2 // Start from row 2

		// Hostname
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), result.Server.Hostname)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), cellStyle)

		// Parse output by commands
		cmdOutputs := parseCommandOutputs(result.Output, commands)

		for i, output := range cmdOutputs {
			col := getColumnName(i + 2)
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue(sheetName, cell, output)
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 20)
	for i := range commands {
		col := getColumnName(i + 2)
		f.SetColWidth(sheetName, col, col, 50)
	}

	// Set row heights for better readability
	for i := 2; i <= len(results)+1; i++ {
		f.SetRowHeight(sheetName, i, 100)
	}

	return f.SaveAs(outputPath)
}

// getColumnName converts column index (1-based) to Excel column name
func getColumnName(n int) string {
	result := ""
	for n > 0 {
		n-- // Make it 0-based
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}

// parseCommandOutputs extracts output for each command from the full output
func parseCommandOutputs(fullOutput string, commands []string) []string {
	outputs := make([]string, len(commands))

	if fullOutput == "" {
		return outputs
	}

	lines := strings.Split(fullOutput, "\n")

	// Find command positions
	cmdPositions := make([]int, len(commands))
	for i := range cmdPositions {
		cmdPositions[i] = -1
	}

	for lineIdx, line := range lines {
		for cmdIdx, cmd := range commands {
			// Look for command in line (command might appear after prompt like "Router#show version")
			if strings.Contains(line, cmd) {
				if cmdPositions[cmdIdx] == -1 {
					cmdPositions[cmdIdx] = lineIdx
				}
			}
		}
	}

	// Extract output between commands
	for i := 0; i < len(commands); i++ {
		startLine := cmdPositions[i]
		if startLine == -1 {
			continue
		}

		// Find end line (next command position or end of output)
		endLine := len(lines)
		for j := i + 1; j < len(commands); j++ {
			if cmdPositions[j] != -1 {
				endLine = cmdPositions[j]
				break
			}
		}

		// Extract lines between start and end
		if startLine+1 < endLine {
			outputLines := lines[startLine+1 : endLine]
			outputs[i] = strings.TrimSpace(strings.Join(outputLines, "\n"))
		}
	}

	return outputs
}
