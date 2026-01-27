package cisco

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportToExcel exports execution results to an Excel file
// Format: Columns = Hostnames, Rows = Output lines
// Commands are highlighted with yellow background, bold text, and preceded by empty row
func ExportToExcel(results []ExecutionResult, commands []string, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Results"
	f.SetSheetName("Sheet1", sheetName)

	// Set header style
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

	// Set cell style
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set line number style
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

	// Command highlight style - yellow background + bold
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

	// Command line number style
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

	// Write header row: empty + hostnames
	f.SetCellValue(sheetName, "A1", "Line")
	f.SetCellStyle(sheetName, "A1", "A1", headerStyle)

	for i, result := range results {
		col := getColumnName(i + 2) // B, C, D, ...
		f.SetCellValue(sheetName, col+"1", result.Server.Hostname)
		f.SetCellStyle(sheetName, col+"1", col+"1", headerStyle)
	}

	// Parse outputs into lines and mark command lines
	allOutputLines := make([][]string, len(results))
	allIsCommand := make([][]bool, len(results))
	maxLines := 0

	for i, result := range results {
		lines := strings.Split(result.Output, "\n")
		cleanLines := make([]string, 0, len(lines))
		isCommand := make([]bool, 0, len(lines))

		for _, line := range lines {
			cleanLine := strings.TrimRight(line, "\r")
			cleanLines = append(cleanLines, cleanLine)
			isCommand = append(isCommand, containsCommand(cleanLine, commands))
		}

		allOutputLines[i] = cleanLines
		allIsCommand[i] = isCommand

		if len(cleanLines) > maxLines {
			maxLines = len(cleanLines)
		}
	}

	// Build rows with empty rows before commands
	type rowData struct {
		lineNum   int
		isCommand bool
		isEmpty   bool
	}

	rows := make([]rowData, 0)
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		// Check if any server has a command on this line
		anyCommand := false
		for serverIdx := range results {
			if lineIdx < len(allIsCommand[serverIdx]) && allIsCommand[serverIdx][lineIdx] {
				anyCommand = true
				break
			}
		}

		// Add empty row before command (except for first line)
		if anyCommand && lineIdx > 0 {
			rows = append(rows, rowData{lineNum: -1, isCommand: false, isEmpty: true})
		}

		rows = append(rows, rowData{lineNum: lineIdx, isCommand: anyCommand, isEmpty: false})
	}

	// Write data rows
	for rowIdx, row := range rows {
		excelRow := rowIdx + 2 // Start from row 2

		if row.isEmpty {
			// Empty row
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", excelRow), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), lineNumStyle)
			for serverIdx := range results {
				col := getColumnName(serverIdx + 2)
				cell := fmt.Sprintf("%s%d", col, excelRow)
				f.SetCellValue(sheetName, cell, "")
				f.SetCellStyle(sheetName, cell, cell, cellStyle)
			}
			continue
		}

		lineIdx := row.lineNum

		// Line number
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", excelRow), lineIdx+1)
		if row.isCommand {
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), cmdLineNumStyle)
		} else {
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), lineNumStyle)
		}

		// Each server's line
		for serverIdx := range results {
			col := getColumnName(serverIdx + 2)
			cell := fmt.Sprintf("%s%d", col, excelRow)

			if lineIdx < len(allOutputLines[serverIdx]) {
				f.SetCellValue(sheetName, cell, allOutputLines[serverIdx][lineIdx])
			} else {
				f.SetCellValue(sheetName, cell, "")
			}

			// Apply command style if this line contains a command
			if row.isCommand {
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

	return f.SaveAs(outputPath)
}

// containsCommand checks if a line contains any of the commands
func containsCommand(line string, commands []string) bool {
	for _, cmd := range commands {
		if strings.Contains(line, cmd) {
			return true
		}
	}
	return false
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
