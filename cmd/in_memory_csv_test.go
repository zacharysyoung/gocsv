package cmd

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
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

type floatCount = struct {
	value float64
	count int
}

type floatStats struct {
	min         float64
	median      float64
	max         float64
	sum         float64
	mean        float64
	stdDev      float64
	uniqueCount int
	valueCounts []floatCount
}

type intCount = struct {
	value int
	count int
}

type intStats struct {
	min         int
	median      float64
	max         int
	sum         int
	mean        float64
	stdDev      float64
	uniqueCount int
	valueCounts []intCount
}

type numeric interface{ float64 | int }

func getNumericStats[Num numeric](nums []Num) (
	minN Num,
	medianN float64,
	maxN Num,
	sum Num,
	mean float64,
	stdDev float64,
	uniqueCount int,
	valueCounts []struct {
		value Num
		count int
	}) {

	n := len(nums)

	// min, median, max
	slices.Sort(nums)
	minN, maxN = nums[0], nums[n-1]
	switch {
	case n%2 == 0:
		i, j := n/2-1, n/2
		medianN = float64(nums[i]+nums[j]) / 2
	default:
		i := (n - 1) / 2
		medianN = float64(nums[i])
	}

	// sum and raw value counts
	m := make(map[Num]int)
	for _, x := range nums {
		sum += x
		m[x]++
	}

	// mean
	mean = float64(sum) / float64(n)

	// stdDev
	acc := 0.0
	for _, x := range nums {
		y := float64(x) - mean
		acc += (y * y)
	}
	stdDev = math.Sqrt(acc / float64(n-1))

	// uniqueCount
	uniqueCount = len(m)

	// convert raw map to properly sorted slice of value counts
	valueCounts = make([]struct {
		value Num
		count int
	}, 0, len(m))
	for val, ct := range m {
		valueCounts = append(valueCounts, struct {
			value Num
			count int
		}{val, ct})
	}
	slices.SortFunc(valueCounts, func(a, b struct {
		value Num
		count int
	}) int {
		// count (descending); break tie on value (ascending)
		if a.count == b.count {
			return int(a.value - b.value)
		}
		return b.count - a.count
	})

	return minN,
		medianN,
		maxN,
		sum,
		mean,
		stdDev,
		uniqueCount,
		valueCounts
}

var (
	posInf = math.Inf(0)

	// equal-ish for floats
	equal = func(a, b float64) bool {
		if a == posInf || b == posInf {
			if a == posInf && b == posInf {
				return true
			}
			return false
		}
		return math.Abs(a-b) <= 1e-14
	}
)

func Test_getNumericStats(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		testCases := []struct {
			nums []int

			min         int
			median      float64
			max         int
			sum         int
			mean        float64
			stdDev      float64
			uniqueCount int
			valueCounts []intCount
		}{
			{
				nums:        []int{1, 2, 3},
				min:         1,
				median:      2,
				max:         3,
				sum:         6,
				mean:        2,
				stdDev:      1,
				uniqueCount: 3,
				valueCounts: []intCount{
					{1, 1},
					{2, 1},
					{3, 1},
				},
			},
			{
				nums:        []int{1, 2, 3, 4},
				min:         1,
				median:      2.5,
				max:         4,
				sum:         10,
				mean:        2.5,
				stdDev:      1.2909944487358055997817,
				uniqueCount: 4,
				valueCounts: []intCount{
					{1, 1},
					{2, 1},
					{3, 1},
					{4, 1},
				},
			},
			{
				nums:        []int{1, 2, 3, 3, 4, 4, 4},
				min:         1,
				median:      3,
				max:         4,
				sum:         21,
				mean:        3,
				stdDev:      1.15470053837925,
				uniqueCount: 4,
				valueCounts: []intCount{
					{4, 3},
					{3, 2},
					{1, 1},
					{2, 1},
				},
			},
			// e17, biggest exponent on my machine before overflow
			{
				nums:        []int{1e17, 2e17, 3e17, 3e17, 4e17, 4e17, 4e17},
				min:         1e17,
				median:      3e17,
				max:         4e17,
				sum:         21e17,
				mean:        3e17,
				stdDev:      115_470_053_837_925_136, // WolframAlpha computed 115_470_053_837_925_168
				uniqueCount: 4,
				valueCounts: []intCount{
					{4e17, 3},
					{3e17, 2},
					{1e17, 1},
					{2e17, 1},
				},
			},
			// e18, biggest exponent on my machine the compiler allows
			{
				nums:        []int{1e18, 2e18, 3e18, 3e18, 4e18, 4e18, 4e18},
				min:         1e18,
				median:      3e18,
				max:         4e18,
				sum:         2_553_255_926_290_448_384,
				mean:        364_750_846_612_921_216,
				stdDev:      3_071_692_440_739_881_984,
				uniqueCount: 4,
				valueCounts: []intCount{
					{4e18, 3},
					{3e18, 2},
					{1e18, 1},
					{2e18, 1},
				},
			},
		}

		for _, tc := range testCases {
			assertNumericStats(`
for %v:
  min:    %d; %d
  median: %.14f; %.14f
  max:    %d; %d
  sum:    %d; %d
  mean:   %.14f; %.14f
  stdDev: %.14f; %.14f
  unique: %d; %d
  counts: %v; %v
`,

				tc.nums,
				tc.min,
				tc.median,
				tc.max,
				tc.sum,
				tc.mean,
				tc.stdDev,
				tc.uniqueCount,
				tc.valueCounts,

				t)
		}
	})

	t.Run("floats", func(t *testing.T) {
		testCases := []struct {
			nums []float64

			min         float64
			median      float64
			max         float64
			sum         float64
			mean        float64
			stdDev      float64
			uniqueCount int
			valueCounts []floatCount
		}{
			{
				nums:        []float64{1, 2, 3},
				min:         1,
				median:      2.0,
				max:         3,
				sum:         6,
				mean:        2.0,
				stdDev:      1.0,
				uniqueCount: 3,
				valueCounts: []floatCount{
					{1, 1},
					{2, 1},
					{3, 1},
				},
			},
			{
				nums:        []float64{1, 2, 3, 4},
				min:         1,
				median:      2.5,
				max:         4,
				sum:         10,
				mean:        2.5,
				stdDev:      1.2909944487358055997817,
				uniqueCount: 4,
				valueCounts: []floatCount{
					{1, 1},
					{2, 1},
					{3, 1},
					{4, 1},
				},
			},
			// e22, biggest exponent on my machine before overflow
			{
				nums:        []float64{1e22, 2e22, 3e22, 3e22, 4e22, 4e22, 4e22},
				min:         10_000_000_000_000_000_000_000,
				median:      30_000_000_000_000_000_000_000,
				max:         40_000_000_000_000_000_000_000,
				sum:         210_000_000_000_000_012_582_912,
				mean:        30_000_000_000_000_000_000_000,
				stdDev:      11_547_005_383_792_515_874_816,
				uniqueCount: 4,
				valueCounts: []floatCount{
					{4e22, 3},
					{3e22, 2},
					{1e22, 1},
					{2e22, 1},
				},
			},
			// e307, biggest exponent on my machine the compiler allows
			{
				nums:        []float64{1e307, 2e307, 3e307, 3e307, 4e307, 4e307, 4e307},
				min:         9999999999999999860310597602564577717002641838126363875249660735883565852672743849064846414228960666786379280392654615393353172850252103336275952370615397010730691664689375178569039851073146339641623266071126720011020169553304018596457812688561947201171488461172921822139066929851282122002676667750021070848.0,
				median:      29999999999999998333531599348493850865774979866354987833591944435489734118926205023937106824584340884760409408280650665341030240930446782526054114365850153070209701066048487835703573958790891195463793895486513170874712543319959559957616903615141848547979922603491167024466633992206320339533970484285057007616.0,
				max:         39999999999999999441242390410258310868010567352505455500998642943534263410690975396259385656915842667145517121570618461573412691401008413345103809482461588042922766658757500714276159404292585358566493064284506880044080678213216074385831250754247788804685953844691687288556267719405128488010706671000084283392.0,
				sum:         posInf,
				mean:        posInf,
				stdDev:      posInf,
				uniqueCount: 4,
				valueCounts: []floatCount{
					{4e307, 3},
					{3e307, 2},
					{1e307, 1},
					{2e307, 1},
				},
			},
		}

		for _, tc := range testCases {
			assertNumericStats(`
for %v:
  min:    %.14f; %.14f
  median: %.14f; %.14f
  max:    %.14f; %.14f
  sum:    %.14f; %.14f
  mean:   %.14f; %.14f
  stdDev: %.14f; %.14f
  unique: %d; %d
  counts: %v; %v
`,

				tc.nums,
				tc.min,
				tc.median,
				tc.max,
				tc.sum,
				tc.mean,
				tc.stdDev,
				tc.uniqueCount,
				tc.valueCounts,

				t)
		}
	})
}

func assertNumericStats[Num numeric](
	format string,

	nums []Num,
	min Num,
	median float64,
	max Num,
	sum Num,
	mean float64,
	stdDev float64,
	uniqueCount int,
	valueCounts []struct {
		value Num
		count int
	},

	t *testing.T) {

	t.Helper()

	_min, _median, _max, _sum, _mean, _stdDev, _uniqueCount, _valueCounts := getNumericStats(nums)

	if _min != min ||
		!equal(_median, median) ||
		_max != max ||
		_sum != sum ||
		!equal(_mean, mean) ||
		!equal(_stdDev, stdDev) ||
		_uniqueCount != uniqueCount ||
		!reflect.DeepEqual(_valueCounts, valueCounts) {
		t.Errorf(
			format,

			nums,

			_min, min,
			_median, median,
			_max, max,
			_sum, sum,
			_mean, mean,
			_stdDev, stdDev,
			_uniqueCount, uniqueCount,
			_valueCounts, valueCounts,
		)
	}
}
