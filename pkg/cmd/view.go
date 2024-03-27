package cmd

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
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
		cmd     = flag.NewFlagSet("view", flag.ExitOnError)
		mdflag  = cmd.Bool("md", false, "print as (extended) Markdown table")
		boxflag = cmd.Bool("box", false, "print complete cells in simple ascii boxes")
		wflag   int
	)
	cmd.Func("maxw", "cap the width of printed cells; minimum of 3", func(s string) error {
		var err error
		wflag, err = strconv.Atoi(s)
		if err != nil {
			return err
		}
		if wflag < 3 {
			return errors.New("must be minimum of 3")
		}
		return nil
	})

	flag.Usage = usage
	cmd.Parse(args)

	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	widths := getColWidths(recs)
	types := inferCols(recs[1:], nil)

	if wflag >= 3 {
		cap := false
		for i := range widths {
			if widths[i] > wflag {
				widths[i] = wflag
				cap = true
			}
		}
		if cap {
			truncateCells(recs, wflag)
		}
	}

	if *mdflag {
		printMarkdown(w, recs, widths, types)
	} else if *boxflag {
		fmt.Fprintln(w, "+------+\n| boxy |\n+------+")
	} else {
		printSimple(w, recs, widths, types)
	}

	return nil
}

func printSimple(w io.Writer, recs [][]string, widths []int, types []inferredType) {
	sep, comma := "", ","
	for i, x := range recs[0] {
		if i == len(recs[0])-1 {
			comma = ""
		}
		fmt.Fprintf(w, "%s%s", sep, pad(x, comma, stringType, widths[i]))
		sep = " "
	}
	fmt.Fprint(w, "\n")

	for i := 1; i < len(recs); i++ {
		sep, comma = "", ","
		for j, x := range recs[i] {
			if j == len(recs[i])-1 {
				comma = ""
			}
			fmt.Fprintf(w, "%s%s", sep, pad(x, comma, types[j], widths[j]))
			sep = " "
		}
		fmt.Fprint(w, "\n")
	}
}

func printMarkdown(w io.Writer, recs [][]string, widths []int, types []inferredType) {
	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", pad(x, "", stringType, widths[i]))
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
			fmt.Fprintf(w, "| %s ", pad(rec[i], "", types[i], widths[i]))
		}
		fmt.Fprint(w, "|\n")
	}
}

// getColWidths returns a slice of the widths of the widest
// cell in each column of recs.
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

// truncateCells truncates cells wider than maxw.  The final
// width of a truncated cell accounts for "..." being appended.
func truncateCells(recs [][]string, maxw int) {
	for i := range recs {
		for j := range recs[i] {
			r := []rune(recs[i][j])
			if n := len(r); n > maxw {
				recs[i][j] = fmt.Sprintf("%s...", (string(r[:maxw-3])))
			}
		}
	}
}

// pad pads x with n-number spaces; left-justify if it==stringType,
// right-justify otherwise.
func pad(x, suf string, it inferredType, n int) string {
	if suf != "" {
		n++
	}
	if it == stringType {
		n *= -1
	}
	return fmt.Sprintf("%*s", n, x+suf)
}

// splitLinebreaks splits field-data with line breaks
// by inserting new records into recs and copying the
// excess multi-line values down.
// E.g., [[a\nb c] [d e]] to [[a c][b  ][d e]]
func splitLinebreaks(recs [][]string) [][]string {
	for i := 0; i < len(recs); {
		grow := 0
		for j := range recs[i] {
			xs := strings.Split(recs[i][j], "\n")
			for k := range len(xs) {
				if k > grow {
					x := make([]string, len(recs[i]))
					recs = slices.Insert(recs, i+k, x)
					grow++
				}
				recs[i+k][j] = xs[k]
			}
		}
		i += 1 + grow
	}
	return recs
}
