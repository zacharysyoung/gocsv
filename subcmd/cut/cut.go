// Package cut implements a Runner that selects or omits specified
// columns.
package cut

import (
	"encoding/csv"
	"io"

	"github.com/zacharysyoung/gocsv/subcmd"
)

// Cut reads the input CSV record-by-record and writes only specific
// fields of each record to the output CSV.
type Cut struct {
	ColGroups []subcmd.ColGroup // 1-based indices of the columns to include, or exclude

	Exclude bool
}

func NewCut(groups []subcmd.ColGroup, exclude bool) *Cut {
	return &Cut{ColGroups: groups, Exclude: exclude}
}

func (xx *Cut) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	cols, err := subcmd.FinalizeCols(xx.ColGroups, header)
	if err != nil {
		return err
	}

	if len(cols) > 0 && xx.Exclude {
		cols = exclude(cols, header)
	}
	cols = subcmd.Base0Cols(cols)

	row := make([]string, len(cols))
	write := func(rec []string) {
		row = row[:]
		for i, x := range cols {
			row[i] = rec[x]
		}
		ww.Write(row)
	}
	write(header)

	for {
		rec, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		write(rec)
	}
	ww.Flush()
	return ww.Error()
}

// exclude returns header's 1-based indexes minus cols.
func exclude(cols []int, header []string) []int {
	excludes := make(map[int]bool)
	for _, x := range cols {
		excludes[x] = true
	}
	final := make([]int, 0)
	for i := range header {
		i++ // 1-based
		if !excludes[i] {
			final = append(final, i)
		}
	}
	return final
}
