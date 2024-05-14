package sort

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"slices"

	"github.com/zacharysyoung/gocsv/subcmd"
)

type Sort struct {
	ColGroups []subcmd.ColGroup // 1-based indices of columns to use as compare keys
	Reversed  bool
	Stably    bool
}

func NewSort(colGroups []subcmd.ColGroup, reversed, stably bool) *Sort {
	return &Sort{ColGroups: colGroups, Reversed: reversed, Stably: stably}
}

func (xx *Sort) fromJSON(p []byte) error {
	*xx = Sort{}
	return json.Unmarshal(p, xx)
}

func (xx *Sort) CheckConfig() error {
	return nil
}

func (xx *Sort) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	recs, err := rr.ReadAll()
	if err != nil {
		return err
	}

	groups := xx.ColGroups
	cols, err := subcmd.FinalizeCols(groups, recs[0])
	if err != nil {
		return err
	}

	order := 1
	if xx.Reversed {
		order = -1
	}

	sort(recs[1:], cols, order)

	if err := ww.WriteAll(recs); err != nil {
		return err
	}
	ww.Flush()
	return ww.Error()
}

// sort sorts recs according to 1-based cols; 1 for ascending order,
// -1 for descending, panics if order is neither.
func sort(recs [][]string, cols []int, order int) {
	if order != -1 && order != 1 {
		panic(errors.New("order must be -1 or 1"))
	}
	types := subcmd.InferCols(recs, cols)
	cols = subcmd.Base0Cols(cols)
	slices.SortFunc(recs, func(a, b []string) int {
		for i, ix := range cols {
			if x := subcmd.Compare2(a[ix], b[ix], types[i]); x != 0 {
				return x * order
			}
		}
		return 0
	})
}
