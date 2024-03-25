package view

import (
	"encoding/csv"
	"fmt"
	"io"
)

type View struct{}

func (View) Run(rdr io.Reader, wrtr io.Writer, args ...string) error {
	r := csv.NewReader(rdr)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	w := wrtr

	widths := maxWidths(records)

	for _, x := range records {
		sep := ""
		for i, y := range x {
			if _, err := w.Write([]byte(fmt.Sprintf("%s%*s", sep, widths[i], y))); err != nil {
				return err
			}
			sep = ", "
		}
		w.Write([]byte{'\n'})
	}

	return nil
}

func maxWidths(recs [][]string) []int {
	widths := make([]int, len(recs[0]))
	for _, rec := range recs {
		for i, x := range rec {
			if n := len(x); n > widths[i] {
				widths[i] = n
			}
		}
	}
	return widths
}
