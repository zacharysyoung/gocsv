package sort

import (
	"encoding/csv"
	"errors"
	"io"
	"slices"

	"github.com/zacharysyoung/gocsv/subcmd"
)

type Sort struct {
	Ranges   subcmd.Ranges
	Reversed bool
	Stably   bool
}

func NewSort(ranges []subcmd.Range, reversed, stably bool) *Sort {
	return &Sort{Ranges: ranges, Reversed: reversed, Stably: stably}
}

func (xx *Sort) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var (
		err error

		rows [][]string
		cols []int
	)

	if rows, err = rr.ReadAll(); err != nil {
		return err
	}

	switch len(xx.Ranges) {
	case 0:
		cols = subcmd.Base1Cols(rows[0])
	default:
		if cols, err = xx.Ranges.Finalize(rows[0]); err != nil {
			return err
		}
	}

	order := 1
	if xx.Reversed {
		order = -1
	}
	sort(rows[1:], cols, order)

	if err := ww.WriteAll(rows); err != nil {
		return err
	}
	ww.Flush()
	return ww.Error()
}

// sort sorts rows according to 1-based cols; 1 for ascending order,
// -1 for descending, panics if order is neither.
func sort(rows [][]string, cols []int, order int) {
	if order != -1 && order != 1 {
		panic(errors.New("order must be -1 or 1"))
	}
	types := subcmd.InferCols(rows, cols)
	cols = subcmd.Base0Cols(cols)
	slices.SortFunc(rows, func(a, b []string) int {
		for i, ix := range cols {
			if x := subcmd.Compare2(a[ix], b[ix], types[i]); x != 0 {
				return x * order
			}
		}
		return 0
	})
}
