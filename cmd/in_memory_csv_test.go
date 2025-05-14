package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/aotimme/gocsv/csv"
)

func TestStats(t *testing.T) {
	var (
		data = strings.TrimSpace(`
A,B
1,x
2,y
3,z
4,z
`)

		want = strings.TrimSpace(`
1. A
  Type: int
  Number NULL: 0
  Min: 1
  Max: 4
  Sum: 10
  Mean: 2.500000
  Median: 2.000000
  Standard Deviation: 1.290994
  Unique values: 4
  4 most frequent values:
      1: 1
      2: 1
      3: 1
      4: 1

2. B
  Type: string
  Number NULL: 0
  Unique values: 3
  Max length: 1
  3 most frequent values:
      z: 2
      x: 1
      y: 1
`)
	)

	imc := NewInMemoryCsvFromInputCsv(&InputCsv{
		reader: csv.NewReader(strings.NewReader(data)),
	})

	var (
		buf = &bytes.Buffer{} // substitute for stdout, which imc.PrintStats uses

		bufA = imc.GetPrintStatsForColumn(0)
		bufB = imc.GetPrintStatsForColumn(1)
	)

	fmt.Fprintln(buf, bufA.String())
	fmt.Fprintln(buf, bufB.String())

	got := strings.TrimSpace(buf.String())

	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// Fails intermittently because the order of the most frequent
// values is derived from the keys in a map, which have
// non-deterministic ordering.
func TestStats_MostFrequentValues(t *testing.T) {
	var (
		data = strings.TrimSpace(`
A
1
2
3
`)

		want = strings.TrimSpace(`
3 most frequent values:
      1: 1
      2: 1
      3: 1
`)
	)

	imc := NewInMemoryCsvFromInputCsv(&InputCsv{
		reader: csv.NewReader(strings.NewReader(data)),
	})

	var (
		buf = &bytes.Buffer{} // substitute for stdout, which imc.PrintStats uses

		bufA = imc.GetPrintStatsForColumn(0)
	)

	fmt.Fprintln(buf, bufA.String())

	// separate out just the "N most frequent values" part
	parts := strings.Split(buf.String(), "Unique values: 3")
	got := strings.TrimSpace(parts[1])

	if got != want {
		t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
	}
}
