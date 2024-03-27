package cmd

import (
	"fmt"
	"io"
	"strings"
)

type Command interface {
	Run(io.Reader, io.Writer, ...string) error
}

var Commands = map[string]Command{
	"view": View{},
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
