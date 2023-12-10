package cmd

import (
	"errors"
	"io"
	"testing"
)

func TestInputRead(t *testing.T) {
	testRows := [][]string{
		{"Col1", "Col2"},
		{"r1c1", "r1c2"},
		{"r2c1", "r2c2"},
		{"r3c1", "r3c2"},
	}

	var (
		gotRow []string
		err    error
	)

	// iterative reads
	ic := newTestInputCsv(testRows)

	for i := 0; i < len(testRows); i++ {
		gotRow, err = ic.Read()
		if err != nil {
			t.Errorf("for ic.Read() iteration %d got err %v; want nil", i+1, err)
		}
		if err = assertRowEqual(gotRow, testRows[i]); err != nil {
			t.Errorf("for ic.Read() iteration %d: %v", i+1, err)
		}
	}

	gotRow, err = ic.Read()
	if gotRow != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got record %v and err %v; want nil and io.EOF", gotRow, err)
	}
}
func TestInputReadAll(t *testing.T) {
	var (
		testRows = [][]string{
			{"Col1", "Col2"},
			{"r1c1", "r1c2"},
			{"r2c1", "r2c2"},
			{"r3c1", "r3c2"},
		}

		gotRow  []string
		gotRows [][]string
		err     error
	)

	ic := newTestInputCsv(testRows)

	gotRows, err = ic.ReadAll()
	if err != nil {
		t.Errorf("for ic.ReadAll() got err %v; want nil", err)
	}
	if err = assertRowsEqual(testRows, gotRows); err != nil {
		t.Errorf("for ic.ReadAll(): %v", err)
	}

	gotRow, err = ic.Read()
	if gotRow != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got record %v and err %v; want nil and io.EOF", gotRow, err)
	}

	gotRows, err = ic.ReadAll()
	if gotRow != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got records %v and err %v; want nil and io.EOF", gotRows, err)
	}

	// single read, then read all
	ic = newTestInputCsv(testRows)

	gotRow, err = ic.Read()
	if err != nil {
		t.Errorf("for first ic.Read() got err %v; want nil", err)
	}
	if err = assertRowEqual(gotRow, testRows[0]); err != nil {
		t.Errorf("for first ic.Read(): %v", err)
	}

	gotRows, err = ic.ReadAll()
	if err != nil {
		t.Errorf("for next ic.ReadAll() got err %v; want nil", err)
	}
	if err = assertRowsEqual(testRows[1:], gotRows); err != nil {
		t.Errorf("for next ic.ReadAll(): %v", err)
	}
}

func TestInputImmutable(t *testing.T) {
	const col1 = "Col1"
	testRows := [][]string{
		{col1, "Col2"},
	}

	ic := newTestInputCsv(testRows)

	testRows[0][0] = "Foo"

	gotRow, _ := ic.Read()
	if gotRow[0] != col1 {
		t.Errorf("after changing source row, ic.Read()=%v; want %v", gotRow, col1)
	}

}
func TestOutput(t *testing.T) {
	testRows := [][]string{
		{"Col1", "Col2"},
		{"r1c1", "r1c2"},
	}

	oc := new(testOutputCsv)

	for _, row := range testRows {
		if err := oc.Write(row); err != nil {
			t.Errorf("oc.Write(%v)=%v; want nil", row, err)
		}
	}

	gotRows := oc.getRows()
	if err := assertRowsEqual(testRows, gotRows); err != nil {
		t.Errorf("oc.getRows()=%v; want %v", gotRows, testRows)
	}
}

func TestOutputImmutable(t *testing.T) {
	const col1 = "Col1"
	testRow := []string{"Col1", "Col2"}

	oc := new(testOutputCsv)
	oc.Write(testRow)

	got := oc.getRows()[0][0]
	if got != col1 {
		t.Errorf("before modifying testRow, oc's first field %v!=%v", got, col1)
	}

	testRow[0] = "Foo"

	got = oc.getRows()[0][0]
	if got != col1 {
		t.Errorf("after modifying testRow, oc's first field %v!=%v", got, col1)
	}
}
