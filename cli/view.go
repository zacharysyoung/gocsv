package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
)

type View struct {
	box, md    bool
	maxh, maxw int
}

func (sc *View) CheckConfig() error {
	return nil
}

func (sc *View) Run(r io.Reader, w io.Writer) error {
	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	types := subcmd.InferCols(recs[1:], subcmd.Base1Cols(recs[0]))
	truncateCells(recs, sc.maxw, sc.maxh)
	widths := getColWidths(recs)

	switch {
	default:
		printSimple(w, recs, widths, types)
	case sc.md:
		printMarkdown(w, recs, widths, types)
	case sc.box:
		printBoxes(w, recs, types)
	}

	return nil
}

func printSimple(w io.Writer, recs [][]string, widths []int, types []subcmd.InferredType) {
	const term = "\n"

	sep, comma := "", ","
	for i, x := range recs[0] {
		if i == len(recs[0])-1 {
			comma = ""
		}
		fmt.Fprintf(w, "%s%s", sep, pad(x, comma, widths[i], subcmd.StringType))
		sep = " "
	}
	fmt.Fprint(w, term)

	for i := 1; i < len(recs); i++ {
		sep, comma = "", ","
		for j, x := range recs[i] {
			if j == len(recs[i])-1 {
				comma = ""
			}
			fmt.Fprintf(w, "%s%s", sep, pad(x, comma, widths[j], types[j]))
			sep = " "
		}
		fmt.Fprint(w, term)
	}
}

func printMarkdown(w io.Writer, recs [][]string, widths []int, types []subcmd.InferredType) {
	const term = "|\n"

	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", pad(x, "", widths[i], subcmd.StringType))
	}
	fmt.Fprint(w, term)

	var x string
	for i, t := range types {
		n := widths[i]
		switch t {
		case subcmd.StringType:
			x = strings.Repeat("-", n)
		default:
			x = strings.Repeat("-", n-1) + ":"
		}
		fmt.Fprintf(w, "| %s ", x)
	}
	fmt.Fprint(w, term)

	for _, rec := range recs[1:] {
		for i := range rec {
			fmt.Fprintf(w, "| %s ", pad(rec[i], "", widths[i], types[i]))
		}
		fmt.Fprint(w, term)
	}
}

func printBoxes(w io.Writer, recs [][]string, types []subcmd.InferredType) {
	var newlines newlines
	recs, newlines = splitLinebreaks(recs)
	widths := getColWidths(recs)

	const term = "|\n"
	printHR := func() {
		fmt.Fprint(w, "+")
		for i := range widths {
			fmt.Fprint(w, strings.Repeat("-", widths[i]+2))
			fmt.Fprint(w, "+")
		}
		fmt.Fprint(w, "\n")
	}

	printHR()
	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", pad(x, "", widths[i], subcmd.StringType))
	}
	fmt.Fprint(w, term)
	printHR()

	for i := 1; i < len(recs); i++ {
		for j := range recs[i] {
			fmt.Fprintf(w, "| %s ", pad(recs[i][j], "", widths[j], types[j]))
		}
		fmt.Fprint(w, term)
		if _, ok := newlines[i+1]; ok {
			continue
		}
		printHR()
	}
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

func truncateCells(recs [][]string, maxw, maxh int) {
	for i := range recs {
		for j := range recs[i] {
			recs[i][j] = truncate(recs[i][j], maxw, maxh)
		}
	}
}

// truncate truncates x if wider than maxw and cuts multiline
// values down to maxh lines.  The final width of lines accounts
// for "..." being appended if truncated.
func truncate(x string, maxw, maxh int) string {
	s := strings.Split(x, "\n")
	cuth := false
	if maxh > 0 && len(s) > maxh {
		s = s[:maxh]
		cuth = true
	}

	for i, x := range s {
		// abbreivate x if wider than maxw
		r := []rune(x)
		if maxw > 3 && len(r) > maxw {
			x = string(r[:maxw-3]) + "..."
		}
		// abbreviate x if last line after multiline cut, and not already abbreivated
		if i == len(s)-1 && cuth && x[len(x)-3:] != "..." {
			if maxw > 3 && len(r)+3 > maxw {
				x = string(r[:maxw-3]) + "..."
			} else {
				x += "..."
			}
		}
		s[i] = x
	}

	return strings.Join(s, "\n")
}

func padCells(recs [][]string, suf string, widths []int, types []subcmd.InferredType) {

}

// pad pads x and suf with n-number spaces.  Left-justify if
// it is stringType, right-justify otherwise.
func pad(x, suf string, n int, it subcmd.InferredType) string {
	if suf != "" {
		n += len([]rune(suf))
	}
	if it == subcmd.StringType {
		n *= -1
	}
	return fmt.Sprintf("%*s", n, x+suf)
}

// newlines keeps track of which lines were inserted
// because of split mulitlines.
type newlines map[int]interface{}

// splitLinebreaks splits field-data with line breaks
// by inserting new records into recs and copying the
// excess multi-line values down.
// E.g., [[a\nb c] [d e]] to [[a c][b  ][d e]]
func splitLinebreaks(recs [][]string) ([][]string, newlines) {
	newlines := make(newlines)
	for i := 0; i < len(recs); {
		grow := 0
		for j := range recs[i] {
			xs := strings.Split(recs[i][j], "\n")
			for k := range len(xs) {
				if k > grow {
					x := make([]string, len(recs[i]))
					recs = slices.Insert(recs, i+k, x)
					newlines[i+k] = nil
					grow++

				}
				recs[i+k][j] = xs[k]
			}
		}
		i += 1 + grow
	}
	return recs, newlines
}
