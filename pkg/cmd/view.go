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
		return printMarkdown(w, recs)
	} else {
		return printSimple(w, recs)
	}
}

func printSimple(w io.Writer, recs [][]string) error {
	widths := maxWidths(recs)
	types := inferCols(recs[1:])

	// header row
	sep := ""
	for i, x := range recs[0] {
		_, err := fmt.Fprintf(w, "%s%s", sep, paddedCell(x, stringType, widths[i]))
		if err != nil {
			return err
		}
		sep = ", "
	}
	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	// body rows
	for _, rec := range recs[1:] {
		sep = ""
		for i, x := range rec {
			_, err := fmt.Fprintf(w, "%s%s", sep, paddedCell(x, types[i], widths[i]))
			if err != nil {
				return err
			}
			sep = ", "
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}

	return nil
}

func printMarkdown(w io.Writer, recs [][]string) error {
	widths := maxWidths(recs)
	types := inferCols(recs[1:])

	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", paddedCell(x, stringType, widths[i]))
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
			fmt.Fprintf(w, "| %s ", paddedCell(rec[i], types[i], widths[i]))
		}
		fmt.Fprint(w, "|\n")
	}

	return nil
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

func paddedCell(x string, t inferredType, w int) string {
	n := w - len(x)
	pad := strings.Repeat(" ", n)
	switch t {
	case stringType:
		return x + pad
	default:
		return pad + x

	}
}
