// Package cut implements a [subcmd.Runner] that selects or omits specified
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
	Ranges  subcmd.Ranges
	Exclude bool
}

func NewCut(ranges []subcmd.Range, exclude bool) *Cut {
	return &Cut{Ranges: ranges, Exclude: exclude}
}

func (xx *Cut) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var (
		err error

		header, row []string
		cols        []int

		write func([]string) // handy, closes over ww and final cols
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
	if xx.Exclude {
		cols = exclude(cols, header)
	}
	row = make([]string, len(cols))

	cols = subcmd.Base0Cols(cols)
	write = func(rec []string) {
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
