package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"strings"
)

type View struct{}

var (
	mdflag = flag.Bool("md", false, "print as (extended) Markdown table")
)

func (View) Run(r io.Reader, w io.Writer, args ...string) error {
	setOSArgs(args...)
	flag.Parse()

	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	if *mdflag {
		printMarkdown(w, recs)
	} else {
		printSimple(w, recs)
	}
	return nil
}

func printSimple(w io.Writer, recs [][]string) {
	widths := maxWidths(recs)
	types := inferCols(recs[1:])

	sep := ""
	for i, x := range recs[0] {
		fmt.Fprintf(w, "%s%s", sep, pad(x, stringType, widths[i]))
		sep = ", "
	}
	fmt.Fprint(w, "\n")

	for _, rec := range recs[1:] {
		sep = ""
		for i, x := range rec {
			fmt.Fprintf(w, "%s%s", sep, pad(x, types[i], widths[i]))
			sep = ", "
		}
		fmt.Fprint(w, "\n")
	}
}

func printMarkdown(w io.Writer, recs [][]string) {
	widths := maxWidths(recs)
	types := inferCols(recs[1:])

	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", pad(x, stringType, widths[i]))
	}
	fmt.Fprint(w, "|\n")

	for i, t := range types {
		n := widths[i]
		var cell string
		switch t {
		case stringType:
			cell = strings.Repeat("-", n)
		default:
			cell = strings.Repeat("-", n-1) + ":"
		}
		fmt.Fprintf(w, "| %s ", cell)
	}
	fmt.Fprint(w, "|\n")

	for _, rec := range recs[1:] {
		for i := range rec {
			fmt.Fprintf(w, "| %s ", pad(rec[i], types[i], widths[i]))
		}
		fmt.Fprint(w, "|\n")
	}
}

func maxWidths(recs [][]string) []int {
	widths := make([]int, len(recs[0]))
	for _, rec := range recs {
		for i, x := range rec {
			if n := len(x); n > widths[i] {
				widths[i] = n
			}
		}
	}
	return widths
}

// pad pads x with n-number spaces; left-justify strings,
// right-justify all other inferredTypes.
func pad(x string, t inferredType, n int) string {
	if t == stringType {
		n *= -1
	}
	return fmt.Sprintf("%*s", n, x)
}
