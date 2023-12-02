package cmd

import (
	"testing"
)

func TestRunCap(t *testing.T) {
	// Mimic my ideal scenario: call cap on a headerless CSV w/**no** flags;
	defaultName := "DefCol"

	// Starting with ../test-files/a-b.csv, which looks like:
	// ```
	// a,b
	// ```
	// expect something like:
	// ```
	// DefCol 1,DecCol 2
	// a,b
	// ```
	firstRow := []string{"a", "b"}

	for _, tc := range []struct {
		tname      string
		colNames   string
		truncate   bool
		wantHeader []string
	}{
		{
			tname:      "No names, all default", // my ideal scenario
			colNames:   "",
			truncate:   false,
			wantHeader: []string{"DefCol 1", "DefCol 2"},
		},
		{
			tname:      "One name, one default",
			colNames:   "Foo",
			truncate:   false,
			wantHeader: []string{"Foo", "DefCol 1"},
		},
		{
			tname:      "Two names",
			colNames:   "Foo,Bar",
			truncate:   false,
			wantHeader: []string{"Foo", "Bar"},
		},
		{
			tname:      "Too many names, must truncate",
			colNames:   "Baz,Bar,Foo",
			truncate:   true,
			wantHeader: []string{"Baz", "Bar"},
		},
	} {
		t.Run(tc.tname, func(t *testing.T) {
			ic, err := NewInputCsv("../test-files/a-b.csv")
			if err != nil {
				t.Error("Unexpected error", err)
			}
			toc := new(testOutputCsv)
			sub := new(CapSubcommand)
			sub.namesString = tc.colNames
			sub.truncateNames = tc.truncate
			sub.defaultName = defaultName
			sub.RunCap(ic, toc)

			if len(toc.rows) != 2 {
				t.Errorf("after cap, got %d rows; want 2", len(toc.rows))
			}

			if err := assertRowEqual(toc.rows[1], firstRow); err != nil {
				t.Errorf("first row wrong: %v", err)
			}

			if err := assertRowEqual(toc.rows[0], tc.wantHeader); err != nil {
				t.Errorf("cap(colNames:%q, truncate:%v) = %v; want %v", tc.colNames, tc.truncate, toc.rows[0], tc.wantHeader)
			}
		})
	}
}
