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
	i    int
	len  int
	rows [][]string
}

func newTestInputCSV(rows [][]string) *testInputCSV {
	return &testInputCSV{i: 0, len: len(rows), rows: rows}
}

func (ic *testInputCSV) Read() ([]string, error) {
	if ic.i == ic.len {
		return nil, io.EOF
	}
	row := ic.rows[ic.i]
	ic.i += 1
	return row, nil
}
func (ic *testInputCSV) ReadAll() ([][]string, error) {
	if ic.i == ic.len {
		return nil, io.EOF
	}
	ic.i = ic.len
	return ic.rows[ic.i:], nil
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
