package cmd

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/aotimme/gocsv/csv"
)

var (
	hasSuffix = strings.HasSuffix
	join      = func(s []string) string { return strings.Join(s, "\n") }
	split     = func(s string) []string { return strings.Split(s, "\n") }
	trim      = strings.TrimSpace
)

func TestPrintStats(t *testing.T) {
	// find the "N most frequent values..." section of lines in
	// printedStats, sort just the individual value-count lines
	// with a simple string sort, return all the joined lines.
	//
	// TODO: remove once we've addressed the non-deterministic sorting of values with the same count
	sortMostFrequentVals := func(printedStats string) string {
		lines := split(trim(printedStats))
		n, line := 0, ""
		for n, line = range lines {
			if hasSuffix(line, "most frequent values:") {
				break
			}
		}
		slices.Sort(lines[n+1:])
		return join(lines)
	}

	const header = "Col_A"

	testCases := []struct {
		name string
		col  []string // single column of values
		want string   // put most-freq-vals in the desired order; the actual order will be ignored till we address the non-deterministic sorting of values with the same count
	}{
		{
			"null",
			[]string{"", "", ""},
			`
1. Col_A
  Type: null
  Number NULL: 3
`,
		},

		{
			"int",
			[]string{"1", "2", "3", "3", "4", "4", "4"},
			`
1. Col_A
  Type: int
  Number NULL: 0
  Min: 1
  Max: 4
  Sum: 21
  Mean: 3.000000
  Median: 3.000000
  Standard Deviation: 1.154701
  Unique values: 4
  4 most frequent values:
      4: 3
      3: 2
      1: 1
      2: 1
`,
		},

		{
			"float",
			[]string{"1.0", "2.0", "3.0", "3.0", "4.0", "4.0", "4.0"},
			`
			1. Col_A
  Type: float
  Number NULL: 0
  Min: 1.000000
  Max: 4.000000
  Sum: 21.000000
  Mean: 3.000000
  Median: 3.000000
  Standard Deviation: 1.154701
  Unique values: 4
  4 most frequent values:
      4.000000: 3
      3.000000: 2
      1.000000: 1
      2.000000: 1
`,
		},

		{
			"bool",
			[]string{"true", "true", "false", "false", "false"},
			`
1. Col_A
  Type: boolean
  Number NULL: 0
  Number TRUE: 2
  Number FALSE: 3
`,
		},

		{
			"date",
			[]string{"2000-01-01", "2000-01-02", "2000-01-03", "2000-01-03", "2000-01-04", "2000-01-04", "2000-01-04"},
			`
1. Col_A
  Type: date
  Number NULL: 0
  Min: 2000-01-01
  Max: 2000-01-04
  Unique values: 4
  4 most frequent values:
      2000-01-04: 3
      2000-01-03: 2
      2000-01-01: 1
      2000-01-02: 1
`,
		},

		{
			"datetime",
			[]string{"2000-01-01T00:00:00Z", "2000-01-02T00:00:00Z", "2000-01-03T00:00:00Z", "2000-01-03T00:00:00Z", "2000-01-04T00:00:00Z", "2000-01-04T00:00:00Z", "2000-01-04T00:00:00Z"},
			`
1. Col_A
  Type: datetime
  Number NULL: 0
  Min: 2000-01-01T00:00:00Z
  Max: 2000-01-04T00:00:00Z
  Unique values: 4
  4 most frequent values:
      2000-01-04T00:00:00Z: 3
      2000-01-03T00:00:00Z: 2
      2000-01-01T00:00:00Z: 1
      2000-01-02T00:00:00Z: 1
`,
		},

		{
			"string",
			[]string{"a", "bb", "ccc", "ccc", "dddd", "dddd", "dddd"},
			`
1. Col_A
  Type: string
  Number NULL: 0
  Unique values: 4
  Max length: 4
  4 most frequent values:
      dddd: 3
      ccc: 2
      a: 1
      bb: 1
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imc := NewInMemoryCsvFromInputCsv(
				&InputCsv{
					reader: csv.NewReader(strings.NewReader(header + "\n" + join(tc.col) + "\n")),
				},
			)
			buf_A := imc.GetPrintStatsForColumn(0)

			buf := &bytes.Buffer{} // substitute for stdout, which imc.PrintStats uses
			fmt.Fprintln(buf, buf_A.String())

			got := sortMostFrequentVals(buf.String())

			// TODO: remove sorting once we've addressed the non-deterministic sorting of values with the same count
			want := sortMostFrequentVals(tc.want)

			if got != want {
				t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
