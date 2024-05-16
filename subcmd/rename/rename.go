package rename

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/zacharysyoung/gocsv/subcmd"
)

// Rename changes column names in the header.
type Rename struct {
	// Ranges are the Ranges of columns to be renamed.
	Ranges subcmd.Ranges

	// Names is list of replacement names that matches the order and
	// count of the finalized indexes in Ranges.
	Names []string

	// Regexp is a regexp to match certain column names in ColGroups.
	Regexp string
	// Repl is the replacement string for names matched by Regexp.
	Repl string
}

func NewRename(ranges []subcmd.Range, names []string, regexp, repl string) *Rename {
	return &Rename{Ranges: ranges, Names: names, Regexp: regexp, Repl: repl}
}

func (xx *Rename) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var (
		err error

		header, row []string
		cols        []int
	)

	if header, err = rr.Read(); err != nil {
		if err == io.EOF {
			return subcmd.ErrNoHeader
		}
		return err
	}

	switch len(xx.Ranges) {
	case 0:
		cols = subcmd.Base1Cols(header)
	default:
		if cols, err = xx.Ranges.Finalize(header); err != nil {
			return err
		}
	}

	names, regexp := len(xx.Names) > 0, xx.Regexp != ""
	fmt.Println(xx.Names, names, xx.Regexp, regexp)
	switch {
	case names && regexp:
		return fmt.Errorf("got both Names: %v and Regexp: %q; cannot use both", xx.Names, xx.Regexp)
	case !names && !regexp:
		return fmt.Errorf("got neither Names nor Regexp; must use one")

	case names:
		if header, err = rename(header, cols, xx.Names); err != nil {
			return err
		}
	case regexp:
		if header, err = replace(header, cols, xx.Regexp, xx.Repl); err != nil {
			return err
		}
	}

	ww.Write(header)
	for {
		if row, err = rr.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ww.Write(row)
	}
	ww.Flush()

	return ww.Error()
}

var errWrongCounts = errors.New("cols and names must be the same length")

// rename changes values in header with pairs of 1-based cols and
// names.
func rename(header []string, cols []int, names []string) ([]string, error) {
	if len(names) != len(cols) {
		return nil, fmt.Errorf("%w: %d names != %d cols", errWrongCounts, len(names), len(cols))
	}
	cols = subcmd.Base0Cols(cols)
	for i, x := range cols {
		header[x] = names[i]
	}
	return header, nil
}

// replace changes values in header that match sre by doing a
// regexpReplaceAllString with repl.
func replace(header []string, cols []int, sre, repl string) (hdr []string, err error) {
	var re *regexp.Regexp
	if re, err = regexp.Compile(sre); err != nil {
		return
	}

	hdr = make([]string, len(header))
	copy(hdr, header)
	for _, x := range subcmd.Base0Cols(cols) {
		hdr[x] = re.ReplaceAllString(header[x], repl)
	}
	return
}
