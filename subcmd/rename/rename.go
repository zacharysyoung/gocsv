package rename

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/zacharysyoung/gocsv/subcmd"
)

// Rename changes column names in the header.
type Rename struct {
	// ColGroups are the ColGroups that represent the columns to be
	// renamed.
	ColGroups []subcmd.ColGroup

	// Names is list of replacement names that matches the order and
	// count of the final indexes in ColGroups.
	Names []string

	// Regexp is a regexp to match certain column names in ColGroups.
	Regexp string
	// Repl is the replacement string for names matched by Regexp.
	Repl string
}

func NewRename(groups []subcmd.ColGroup, names []string, regexp, repl string) *Rename {
	return &Rename{ColGroups: groups, Names: names, Regexp: regexp, Repl: repl}
}

func (xx *Rename) fromJSON(p []byte) error {
	*xx = Rename{}
	return json.Unmarshal(p, xx)
}

func (xx *Rename) CheckConfig() error {
	return nil
}

func (xx *Rename) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var err error

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return subcmd.ErrNoData
		}
		return err
	}

	cols, err := subcmd.FinalizeCols(xx.ColGroups, header)
	if err != nil {
		return err
	}

	switch {
	case len(xx.Names) > 0:
		if header, err = rename(header, cols, xx.Names); err != nil {
			return err
		}
	case xx.Regexp != "":
		if header, err = replace(header, cols, xx.Regexp, xx.Repl); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("need non-empty Names: %v or Regexp: %q", xx.Names, xx.Regexp))
	}

	ww.Write(header)
	for {
		record, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ww.Write(record)
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
