package cmd

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestInput(t *testing.T) {
	rows := [][]string{
		{"Col1", "Col2"},
		{"r1c1", "r1c2"},
		{"r2c1", "r2c2"},
		{"r3c1", "r3c2"},
	}

	var (
		record  []string
		records [][]string
		err     error
	)

	// iterative reads
	ic := newTestInputCSV(rows)

	for i := 0; i < len(rows); i++ {
		record, err = ic.Read()
		if err != nil {
			t.Errorf("for ic.Read() iteration %d got err %v; want nil", i+1, err)
		}
		if err = assertRowEqual(record, rows[i]); err != nil {
			t.Errorf("for ic.Read() iteration %d: %v", i+1, err)
		}
	}

	record, err = ic.Read()
	if record != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got record %v and err %v; want nil and io.EOF", record, err)
	}

	// read all
	ic = newTestInputCSV(rows)

	records, err = ic.ReadAll()
	if err != nil {
		t.Errorf("for ic.ReadAll() got err %v; want nil", err)
	}
	if err = assertRowsEqual(rows, records); err != nil {
		t.Errorf("for ic.ReadAll(): %v", err)
	}

	record, err = ic.Read()
	if record != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got record %v and err %v; want nil and io.EOF", record, err)
	}

	records, err = ic.ReadAll()
	if record != nil || !errors.Is(err, io.EOF) {
		t.Errorf("read past end of rows, got records %v and err %v; want nil and io.EOF", records, err)
	}

	// single read, then read all
	ic = newTestInputCSV(rows)

	record, err = ic.Read()
	if err != nil {
		t.Errorf("for first ic.Read() got err %v; want nil", err)
	}
	if err = assertRowEqual(record, rows[0]); err != nil {
		t.Errorf("for first ic.Read(): %v", err)
	}

	records, err = ic.ReadAll()
	if err != nil {
		t.Errorf("for next ic.ReadAll() got err %v; want nil", err)
	}
	if err = assertRowsEqual(rows[1:], records); err != nil {
		t.Errorf("for next ic.ReadAll(): %v", err)
	}
}

func TestInputImmutable(t *testing.T) {
	testRows := [][]string{{"Col1", "Col2"}}
	ic := newTestInputCSV(testRows)

	fmt.Println(ic.rows)
	testRows[0][0] = "Foo"
	fmt.Println(ic.rows)

	record, _ := ic.Read()
	if record[0] != "Col1" {
		t.Errorf("after changing source row, first field of ic.Read() is %q; want \"Col1\"", record[0])
	}

}
