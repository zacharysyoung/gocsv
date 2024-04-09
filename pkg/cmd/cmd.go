package cmd

import (
	"fmt"
	"io"
	"strings"
)

type SubCommander interface {
	CheckConfig() error
	Run(io.Reader, io.Writer) error
}

type testSubCommander interface {
	SubCommander
	fromJSON([]byte) error
}

// rows wraps a set of records, for printing in test failures.
type rows [][]string

// String prints a pretty rectangle from rows.
func (recs rows) String() string {
	widths := getColWidths(recs)

	var sb strings.Builder
	sb.WriteString("[ ")
	pre := ""
	nl := ""
	for i := range recs {
		sb.WriteString(nl)
		sb.WriteString(pre)
		sep := ""
		for j := range recs[i] {
			sb.WriteString(fmt.Sprintf("%s%*s", sep, widths[j], recs[i][j]))
			sep = ", "
		}
		pre = "  "
		nl = "\n"
	}
	sb.WriteString(" ]")
	return sb.String()
}

// getColWidths returns a slice of the widths of the widest
// cell in each column of recs.
func getColWidths(recs [][]string) []int {
	widths := make([]int, len(recs[0]))
	for i := range recs {
		for j := range recs[i] {
			if n := len([]rune(recs[i][j])); n > widths[j] {
				widths[j] = n
			}
		}
	}
	return widths
}

// Base1Cols returns friendly 1-based indexes for the columns
// in header.
func Base1Cols(header []string) (newCols []int) {
	newCols = make([]int, len(header))
	for i := range header {
		newCols[i] = i + 1
	}
	return
}

// Base0Cols turns friendly 1-based indexes to 0-based indexes.
func Base0Cols(cols []int) (newCols []int) {
	for _, x := range cols {
		newCols = append(newCols, x-1)
	}
	return
}
