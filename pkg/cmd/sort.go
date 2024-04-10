package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"slices"
)

type Sort struct {
	Cols     []int // 1-based indices of columns to use as compare keys
	Reversed bool
	Stably   bool
}

func NewSort(cols []int, reversed, stably bool) *Sort {
	return &Sort{Cols: cols, Reversed: reversed, Stably: stably}
}

func (sc *Sort) fromJSON(p []byte) error {
	*sc = Sort{}
	return json.Unmarshal(p, sc)
}

func (sc *Sort) CheckConfig() error {
	return nil
}

func (sc *Sort) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	recs, err := rr.ReadAll()
	if err != nil {
		return err
	}

	cols := sc.Cols
	if cols == nil {
		cols = Base1Cols(recs[0])
	}

	order := 1
	if sc.Reversed {
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
	types := InferCols(recs, cols)
	cols = Base0Cols(cols)
	slices.SortFunc(recs, func(a, b []string) int {
		for i, ix := range cols {
			if x := compare2(a[ix], b[ix], types[i]); x != 0 {
				return x * order
			}
		}
		return 0
	})
}
