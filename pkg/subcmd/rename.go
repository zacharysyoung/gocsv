package subcmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

// Select reads the input CSV record-by-record and writes only specific
// fields of each record to the output CSV.
type Rename struct {
	// ColGroups are the ColGroups that represent the columns to be
	// renamed.
	ColGroups []ColGroup

	// Names is list of replacement names that matches the order and
	// count of the final indexes in ColGroups.
	Names []string

	// Regexp is a regexp to match certain column names in ColGroups.
	Regexp string
	// Repl is the replacement string for names matched by Regexp.
	Repl string
}

func NewRename(groups []ColGroup, names []string, regexp, repl string) *Rename {
	return &Rename{ColGroups: groups, Names: names, Regexp: regexp, Repl: repl}
}

func (sc *Rename) fromJSON(p []byte) error {
	*sc = Rename{}
	return json.Unmarshal(p, sc)
}

func (sc *Rename) CheckConfig() error {
	return nil
}

func (sc *Rename) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)
	fmt.Printf("%#v\n", sc)
	var err error

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return errNoData
		}
		return err
	}

	cols, err := FinalizeCols(sc.ColGroups, header)
	if err != nil {
		return err
	}

	switch {
	case len(sc.Names) > 0:
		if len(cols) != len(sc.Names) {
			return fmt.Errorf("number of cols %d not equal to number of names %d", len(cols), len(sc.Names))
		}
		header = rename(header, cols, sc.Names)
	case sc.Regexp != "":
		header, err = replace(header, cols, sc.Regexp, sc.Repl)
		if err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("need non-empty Names: %v or Regexp: %q", sc.Names, sc.Regexp))
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

func rename(header []string, cols []int, names []string) []string {
	return header
}

func replace(header []string, cols []int, sre, repl string) ([]string, error) {
	re, err := regexp.Compile(sre)
	if err != nil {
		return nil, err
	}
	for _, x := range cols {
		x-- // 0-based
		header[x] = re.ReplaceAllString(header[x], repl)
	}
	return header, nil
}
