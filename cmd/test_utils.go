package cmd

import "fmt"

type testOutputCsv struct {
	rows [][]string
}

func (toc *testOutputCsv) Write(row []string) error {
	newRow := make([]string, len(row))
	copy(newRow, row)
	toc.rows = append(toc.rows, newRow)
	return nil
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
