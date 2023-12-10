package cmd

import (
	"fmt"
	"io"
)

// testOutputCsv satisfies the OutputCsvWriter interface.
type testOutputCsv struct {
	rows [][]string // all written rows, for comparison later
}

func (toc *testOutputCsv) Write(row []string) error {
	rrow := make([]string, len(row))
	copy(rrow, row)
	toc.rows = append(toc.rows, rrow)
	return nil
}

func (toc *testOutputCsv) getRows() [][]string {
	return toc.rows
}

// testInputCsv satisfies the InputCsvReader interface.
type testInputCsv struct {
	rows  [][]string // all rows
	i     int        // index of next row in rows to read from
	nRows int        // len of input rows during init
}

// newTestInputCsv prepares a testInputCSV, deep-copying all of rows.
func newTestInputCsv(rows [][]string) *testInputCsv {
	llen := len(rows)

	rrows := make([][]string, llen)
	for i, row := range rows {
		rrow := make([]string, len(row))
		copy(rrow, row)
		rrows[i] = rrow
	}

	return &testInputCsv{
		rows:  rrows,
		i:     0,
		nRows: llen,
	}
}

func (ic *testInputCsv) Read() ([]string, error) {
	if ic.i == ic.nRows {
		return nil, io.EOF
	}
	row := ic.rows[ic.i]
	ic.i += 1
	return row, nil
}

func (ic *testInputCsv) ReadAll() ([][]string, error) {
	if ic.i == ic.nRows {
		return nil, io.EOF
	}
	rows := ic.rows[ic.i:]
	ic.i = ic.nRows
	return rows, nil
}

func assertRowEqual(got, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("got %d fields; want %d", len(got), len(want))
	}
	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			return fmt.Errorf("field %d, %s != %s", i+1, got[i], want[i])
		}
	}
	return nil
}

func assertRowsEqual(expectedRows, actualRows [][]string) error {
	if len(expectedRows) != len(actualRows) {
		return fmt.Errorf("expected %d rows but got %d", len(expectedRows), len(actualRows))
	}
	for i, row := range expectedRows {
		if err := assertRowEqual(actualRows[i], row); err != nil {
			return fmt.Errorf("row %d: %v", i+1, err)
		}
	}
	return nil
}
