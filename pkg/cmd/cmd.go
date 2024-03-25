package cmd

import (
	"fmt"
	"io"
	"strings"
)

type Command interface {
	Run(io.Reader, io.Writer, ...string) error
}

// rows wraps a set of records, for printing in test failures.
type rows [][]string

// String prints a pretty rectangle from rows.
func (rows rows) String() string {
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, x := range row {
			if n := len(x); n > widths[i] {
				widths[i] = n
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("[ ")
	pre := ""
	nl := ""
	for _, row := range rows {
		sb.WriteString(nl)
		sb.WriteString(pre)
		sep := ""
		for i, x := range row {
			if _, err := sb.WriteString(fmt.Sprintf("%s%*s", sep, widths[i], x)); err != nil {
				panic(err)
			}
			sep = ", "
		}
		pre = "  "
		nl = "\n"
	}
	sb.WriteString(" ]")
	return sb.String()
}
