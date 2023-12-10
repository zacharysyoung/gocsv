package cmd

import (
	"fmt"
	"io"
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

type testInputCSV struct {
	rows    [][]string // all rows
	i       int        // index of next row in rows to read/start from
	numRows int        // record len of rows during init
}

// newTestInputCSV prepares a testInputCSV, copying all of inputRows.
func newTestInputCSV(inputRows [][]string) *testInputCSV {
	_len := len(inputRows)

	_rows := make([][]string, _len)
	for i, x := range inputRows {
		_row := make([]string, len(x))
		copy(_row, x)
		_rows[i] = _row
	}

	return &testInputCSV{
		rows:    _rows,
		i:       0,
		numRows: _len,
	}
}

func (ic *testInputCSV) Read() ([]string, error) {
	if ic.i == ic.numRows {
		return nil, io.EOF
	}
	row := ic.rows[ic.i]
	ic.i += 1
	return row, nil
}
func (ic *testInputCSV) ReadAll() ([][]string, error) {
	if ic.i == ic.numRows {
		return nil, io.EOF
	}
	records := ic.rows[ic.i:]
	ic.i = ic.numRows
	return records, nil
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
