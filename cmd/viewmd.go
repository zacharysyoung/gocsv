package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type ViewMDSubcommand struct {
	maxWidth int
	maxLines int
	maxRows  int
}

func (sub *ViewMDSubcommand) Name() string {
	return "viewmd"
}
func (sub *ViewMDSubcommand) Aliases() []string {
	return []string{}
}
func (sub *ViewMDSubcommand) Description() string {
	return "Display a CSV in Markdown (MD) tabular format."
}
func (sub *ViewMDSubcommand) SetFlags(fs *flag.FlagSet) {
	fs.IntVar(&sub.maxWidth, "max-width", 0, "Maximum width per column")
	fs.IntVar(&sub.maxWidth, "w", 0, "Maximum width per column (shorthand)")
	fs.IntVar(&sub.maxLines, "max-lines", 0, "Maximum number of lines per cell")
	fs.IntVar(&sub.maxLines, "l", 0, "Maximum number of lines per cell (shorthand)")
	fs.IntVar(&sub.maxRows, "n", 0, "Number of rows to display")
}

func (sub *ViewMDSubcommand) Run(args []string) {
	if sub.maxWidth < 0 {
		fmt.Fprintln(os.Stderr, "Invalid argument --max-width")
		os.Exit(1)
	}
	if sub.maxLines < 0 {
		fmt.Fprintln(os.Stderr, "Invalid argument --max-lines")
		os.Exit(1)
	}
	if sub.maxRows < 0 {
		sub.maxRows = 0
	}

	inputCsvs := GetInputCsvsOrPanic(args, 1)
	ViewMD(inputCsvs[0], sub.maxWidth, sub.maxLines, sub.maxRows)
}

func ViewMD(inputCsv *InputCsv, maxWidth, maxLines, maxRows int) {

	imc := NewInMemoryCsvFromInputCsv(inputCsv)

	// Default to 0
	columnWidths := make([]int, imc.NumColumns())
	for j, cell := range imc.header {
		cellLength := getCellWidth(cell, maxLines)
		if cellLength > columnWidths[j] {
			if maxWidth > 0 && cellLength > maxWidth {
				columnWidths[j] = maxWidth
			} else {
				columnWidths[j] = cellLength
			}
		}
	}

	// Get the actual number of rows to display
	numRowsToView := imc.NumRows()
	if maxRows > 0 && maxRows < numRowsToView {
		numRowsToView = maxRows
	}

	for i := 0; i < numRowsToView; i++ {
		row := imc.rows[i]
		for j, cell := range row {
			if columnWidths[j] == maxWidth {
				continue
			}
			cellLength := getCellWidth(cell, maxLines)
			if cellLength > columnWidths[j] {
				if maxWidth > 0 && cellLength > maxWidth {
					columnWidths[j] = maxWidth
				} else {
					columnWidths[j] = cellLength
				}
			}
		}
	}

	rowSeparator := getHeaderSeparator(columnWidths)

	// Print header
	printRow(imc.header, columnWidths, maxLines)
	fmt.Println(rowSeparator)

	// Print rows
	for i := 0; i < numRowsToView; i++ {
		row := imc.rows[i]
		printRow(row, columnWidths, maxLines)
	}
}

func getHeaderSeparator(widths []int) string {
	cells := make([]string, len(widths))
	for i, width := range widths {
		cells[i] = strings.Repeat("-", width)
	}
	return fmt.Sprintf("|-%s-|", strings.Join(cells, "-|-"))
}
