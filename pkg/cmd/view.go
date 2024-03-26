package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type View struct{}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: view [-md] [-w] [file]")
	flag.PrintDefaults()
	os.Exit(2)
}

func (View) Run(r io.Reader, w io.Writer, args ...string) error {
	var (
		cmd    = flag.NewFlagSet("view", flag.ExitOnError)
		mdflag = cmd.Bool("md", false, "print as (extended) Markdown table")
		wflag  = cmd.Int("w", 0, "cap the width of printed cells; minimum of 3")
	)

	flag.Usage = usage
	cmd.Parse(args)

	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	widths := getColWidths(recs)
	types := inferCols(recs[1:])

	if *wflag > 3 {
		cap := false
		for i := range widths {
			if widths[i] > *wflag {
				widths[i] = *wflag
				cap = true
			}

		}
		if cap {
			capColWidths(recs, *wflag)
		}
	}

	if *mdflag {
		printMarkdown(w, recs, widths, types)
	} else {
		printSimple(w, recs, widths, types)
	}
	return nil
}

func printSimple(w io.Writer, recs [][]string, widths []int, types []inferredType) {
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

func printMarkdown(w io.Writer, recs [][]string, widths []int, types []inferredType) {
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

func getColWidths(recs [][]string) []int {
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

func capColWidths(recs [][]string, maxw int) {
	for i := range recs {
		for j := range recs[i] {
			r := []rune(recs[i][j])
			if n := len(r); n > maxw {
				recs[i][j] = fmt.Sprintf("%s...", (string(r[:maxw-3])))
			}
		}
	}
}

// pad pads x with n-number spaces; left-justify strings,
// right-justify all other inferredTypes.
func pad(x string, t inferredType, n int) string {
	if t == stringType {
		n *= -1
	}
	return fmt.Sprintf("%*s", n, x)
}
