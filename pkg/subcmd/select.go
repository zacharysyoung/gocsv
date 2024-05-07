package subcmd

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

// Select reads the input CSV record-by-record and writes only specific
// fields of each record to the output CSV.
type Select struct {
	ColGroups []ColGroup // 1-based indices of the columns to include, or exclude

	Exclude bool
}

func NewSelect(groups []ColGroup, exclude bool) *Select {
	return &Select{ColGroups: groups, Exclude: exclude}
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

	cols, err := FinalizeCols(sc.ColGroups, header)
	if err != nil {
		return err
	}

	if len(cols) > 0 && sc.Exclude {
		cols = exclude(cols, header)
	}
	cols = base0Cols(cols)

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
