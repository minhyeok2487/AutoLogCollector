package cisco

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportToExcel exports execution results to an Excel file
// Format: Columns = Hostnames, Rows = Output lines
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

	// Write header row: empty + hostnames
	f.SetCellValue(sheetName, "A1", "Line")
	f.SetCellStyle(sheetName, "A1", "A1", headerStyle)

	for i, result := range results {
		col := getColumnName(i + 2) // B, C, D, ...
		f.SetCellValue(sheetName, col+"1", result.Server.Hostname)
		f.SetCellStyle(sheetName, col+"1", col+"1", headerStyle)
	}

	// Parse outputs into lines
	allOutputLines := make([][]string, len(results))
	maxLines := 0

	for i, result := range results {
		lines := strings.Split(result.Output, "\n")
		// Clean up lines (remove \r)
		cleanLines := make([]string, 0, len(lines))
		for _, line := range lines {
			cleanLine := strings.TrimRight(line, "\r")
			cleanLines = append(cleanLines, cleanLine)
		}
		allOutputLines[i] = cleanLines
		if len(cleanLines) > maxLines {
			maxLines = len(cleanLines)
		}
	}

	// Write data rows
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		row := lineIdx + 2 // Start from row 2

		// Line number
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), lineIdx+1)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), lineNumStyle)

		// Each server's line
		for serverIdx := range results {
			col := getColumnName(serverIdx + 2)
			cell := fmt.Sprintf("%s%d", col, row)

			if lineIdx < len(allOutputLines[serverIdx]) {
				f.SetCellValue(sheetName, cell, allOutputLines[serverIdx][lineIdx])
			} else {
				f.SetCellValue(sheetName, cell, "")
			}
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
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
