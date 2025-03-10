package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"
)

type testOutputCsv struct {
	rows [][]string
}

func (toc *testOutputCsv) Write(row []string) error {
	newRow := make([]string, len(row))
	copy(newRow, row)
	toc.rows = append(toc.rows, newRow)
	return nil
}

func assertRowsEqual(expectedRows, actualRows [][]string) error {
	if len(expectedRows) != len(actualRows) {
		return fmt.Errorf("expected %d rows but got %d", len(expectedRows), len(actualRows))
	}
	for i, row := range expectedRows {
		if len(row) != len(actualRows[i]) {
			return fmt.Errorf("expected row %d to have %d entries but got %d", i, len(row), len(actualRows[i]))
		}
		for j, cell := range row {
			if cell != actualRows[i][j] {
				return fmt.Errorf("expected %s in cell (%d, %d) but got %s", cell, i, j, actualRows[i][j])
			}
		}
	}
	return nil
}

func TestPrePostProcess(t *testing.T) {
	pretty := []byte(`
A,    B,    C,  D
Foo,  x123, i,  1
food, y4,   ii, 2
	`)

	normal := []byte(`
A,B,C,D
Foo,x123,i,1
food,y4,ii,2
	`)

	t.Run("preprocess", func(t *testing.T) {
		if got := preprocess(pretty); trimb(got) != trimb(normal) {
			t.Errorf("\n got:\n%s\nwant:\n%s", trimb(got), trimb(normal))
		}
	})

	t.Run("postprocess", func(t *testing.T) {
		if got := postprocess(normal); trimb(got) != trimb(pretty) {
			t.Errorf("\n got:\n%s\nwant:\n%s", trimb(got), trimb(pretty))
		}
	})
}

// preprocess reads a simple (no quoting), pretty CSV and
// returns a normalized CSV.
//
// csvIn:
//
//	A,    B,   C,    D
//	Foo,  x123,i,    1
//	food, y4,  ii,   2
//
// csvOut:
//
//	A,B,C,D
//	Foo,x123,i,1
//	food,y4,ii,2
func preprocess(csvIn []byte) (csvOut []byte) {
	csvIn = bytes.TrimSpace(csvIn)
	reader := csv.NewReader(bytes.NewReader(csvIn))
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err = w.WriteAll(records); err != nil {
		panic(err)
	}
	w.Flush()
	if err = w.Error(); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// postprocess reads a normalized CSV and returns a pretty CSV
// with padding prefixed to each following column (also strips
// quoting).
//
// csvIn:
//
//	A,B,C,D
//	Foo,x123,i,1
//	food,y4,ii,2
//
// csvOut:
//
//	A,    B,    C,  D
//	Foo,  x123, i,  1
//	food, y4,   ii, 2
func postprocess(csvIn []byte) (csvOut []byte) {
	csvIn = bytes.TrimSpace(csvIn)
	reader := csv.NewReader(bytes.NewReader(csvIn))

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	maxlens := make([]int, len(records[0]))
	for _, record := range records {
		for i, field := range record {
			maxlens[i] = _max(len(field), maxlens[i])
		}
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	// prefix the padding for A to B, with an extra space, etc...
	for _, record := range records {
		for i := len(record) - 1; i >= 1; i-- {
			maxn, n := maxlens[i-1], len(record[i-1])
			record[i] = strings.Repeat(" ", maxn-n+1) + record[i]
		}
		if err = w.Write(record); err != nil {
			panic(err)
		}
	}
	w.Flush()
	if err = w.Error(); err != nil {
		panic(err)
	}

	// writer enforces quotes for leading spaces; don't need
	// them for viewing simple test CSVs in error printouts
	csvOut = bytes.ReplaceAll(buf.Bytes(), []byte{'"'}, []byte{})

	return csvOut
}

// trimb converts b to a string and trims space from either end.
func trimb(b []byte) string {
	return strings.TrimSpace(string(b))
}

// normalize is shorthand for trimb(postprocess(preprocess(csvIn))).
func normalize(csvIn []byte) (csvOut string) {
	return trimb(postprocess(preprocess(csvIn)))
}

// TODO: upgrade to go 1.21 and replace _max w/built-in max.
func _max(a, b int) int {
	if b > a {
		return b
	}
	return a
}
