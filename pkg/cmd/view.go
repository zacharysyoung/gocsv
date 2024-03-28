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

		maxwflag, maxhflag int = -1, -1

		err error
	)
	cmd.Func("maxw", "cap the width of printed cells; minimum of 3", func(s string) error {
		maxwflag, err = strconv.Atoi(s)
		if err != nil {
			return err
		}
		if maxwflag < 3 {
			return errors.New("must be minimum of 3")
		}
		return nil
	})
	cmd.Func("maxh", "cap the height of printed multiline cells; can only be used with -box", func(s string) error {
		if !*boxflag {
			return errors.New("can only be used with -box")
		}
		maxhflag, err = strconv.Atoi(s)
		return err
	})

	flag.Usage = usage
	cmd.Parse(args)

	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	if !*boxflag {
		maxhflag = 1
	}

	types := inferCols(recs[1:], nil)

	for i := range recs {
		for j := range recs[i] {
			recs[i][j] = truncate(recs[i][j], maxwflag, maxhflag)
		}
	}

	widths := getColWidths(recs)

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
	const term = "\n"

	sep, comma := "", ","
	for i, x := range recs[0] {
		if i == len(recs[0])-1 {
			comma = ""
		}
		fmt.Fprintf(w, "%s%s", sep, pad(x, comma, stringType, widths[i]))
		sep = " "
	}
	fmt.Fprint(w, term)

	for i := 1; i < len(recs); i++ {
		sep, comma = "", ","
		for j, x := range recs[i] {
			if j == len(recs[i])-1 {
				comma = ""
			}
			fmt.Fprintf(w, "%s%s", sep, pad(x, comma, types[j], widths[j]))
			sep = " "
		}
		fmt.Fprint(w, term)
	}
}

func printMarkdown(w io.Writer, recs [][]string, widths []int, types []inferredType) {
	const term = "|\n"

	for i, x := range recs[0] {
		fmt.Fprintf(w, "| %s ", pad(x, "", stringType, widths[i]))
	}
	fmt.Fprint(w, term)

	var x string
	for i, t := range types {
		n := widths[i]
		switch t {
		case stringType:
			x = strings.Repeat("-", n)
		default:
			x = strings.Repeat("-", n-1) + ":"
		}
		fmt.Fprintf(w, "| %s ", x)
	}
	fmt.Fprint(w, term)

	for _, rec := range recs[1:] {
		for i := range rec {
			fmt.Fprintf(w, "| %s ", pad(rec[i], "", types[i], widths[i]))
		}
		fmt.Fprint(w, term)
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
