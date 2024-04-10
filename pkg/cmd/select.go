package cmd

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

type Select struct {
	Cols    []int // 1-based indices of the columns to include, or exclude
	Exclude bool
}

func NewSelect(cols []int, exclude bool) *Select {
	return &Select{Cols: cols, Exclude: exclude}
}

func (sc *Select) fromJSON(p []byte) error {
	*sc = Select{}
	return json.Unmarshal(p, sc)
}

func (sc *Select) CheckConfig() error {
	return nil
}

func (sc *Select) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	cols := Base1Cols(header)
	if len(sc.Cols) > 0 {
		if sc.Exclude {
			cols = exclude(cols, sc.Cols)
		} else {
			cols = sc.Cols
		}
	}

	row := make([]string, len(cols))
	write := func(rec []string) {
		row = row[:]
		for i, ii := range cols {
			row[i] = rec[ii-1]
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

// exclude returns cols minus excludes.
func exclude(cols []int, excludes []int) []int {
	em := make(map[int]bool)
	for _, x := range excludes {
		em[x] = true
	}

	final := make([]int, len(cols)-len(excludes))
	for i, j := 0, 0; i < len(cols); i++ {
		x := cols[i]
		if !em[x] {
			final[j] = x
			j++
		}
	}

	return final
}
