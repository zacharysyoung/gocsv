package subcmd

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type SubCommander interface {
	CheckConfig() error
	Run(io.Reader, io.Writer) error
}

// rows wraps a set of records, for printing in test failures.
type rows [][]string

// errNoData reports an empty CSV
var errNoData = errors.New("no data")

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

// ColGroup holds a single column index or range of column indexes.
// All indexes (single or in a range) are presumed to be valid,
// but the subcmd using a group will determine if the indexes
// actually fit within the header of the CSV being processed.
//
// Indexes are 1-based, and all ranges are inclusive. A group can
// be a single index, a closed range, or an open range, for
// example:
// - [1]: the first column
// - [1, 4]: columns 1 to 4
// - [4, 1]: columns 4 to 1
// - [-1, 4]: the first column to column 4 (columns 1 to 4)
// - [4, -1]: column 4 to the last column
type ColGroup []int

var errBadRange = errors.New("bad open-ended range")

// FinalizeCols expands groups into a final, flat list of 1-based
// indexes, checking header to make sure all specified indexes are
// in bounds, e.g.,
//
// or with (N-length) header,
//
//	[[4,6],[8],[-1,3],[9,-1],[7]] → [4,5,6,8,1,2,3,9,N,7]
func FinalizeCols(groups []ColGroup, header []string) (newCols []int, err error) {
	if len(groups) == 0 {
		for i := range header {
			newCols = append(newCols, i+1) // i+1, 1-based
		}
		return
	}

	n := len(header) // 1-based
	for _, x := range groups {
		var a, b int
		switch len(x) {
		case 1:
			a, b = x[0], x[0]
		case 2:
			a, b = x[0], x[1]
			if a == -1 {
				a = 1
			}
			if b == -1 {
				b = n
			}
		default:
			panic(fmt.Errorf("%w: %v", errBadRange, x))
		}

		if a < 1 || b < a {
			panic(fmt.Errorf("group %v fails constraint 1 <= %d < %d", x, a, b))
		}

		if b > n {
			err = fmt.Errorf("%w: %d-%d with %d column(s)", errBadRange, a, b, n)
			return
		}
		for i := a; i <= b; i++ {
			newCols = append(newCols, i)
		}
	}
	return
}
