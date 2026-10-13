package cmd

import (
	"cmp"
	"math"
	"slices"
	"testing"
	"time"
)

type intTestCases []struct {
	values []int64
	want   int64
}

func TestStatsMin(t *testing.T) {
	testCases := intTestCases{
		{[]int64{0, 1}, 0},
		{[]int64{2, 1}, 1},
		{[]int64{2, 3, 4}, 2},
	}
	for _, tc := range testCases {
		got := calculateMin(tc.values)
		if got != tc.want {
			t.Errorf("for %v, got %v; want %v", tc.values, got, tc.want)
		}
	}
}

func TestStatsMax(t *testing.T) {
	testCases := intTestCases{
		{[]int64{0, 1}, 1},
		{[]int64{2, 1}, 2},
		{[]int64{2, 3, 4}, 4},
	}
	for _, tc := range testCases {
		got := calculateMax(tc.values)
		if got != tc.want {
			t.Errorf("for %v, got %v; want %v", tc.values, got, tc.want)
		}
	}
}

func TestStatsSum(t *testing.T) {
	testCases := intTestCases{
		{[]int64{0}, 0},
		{[]int64{1}, 1},
		{[]int64{1, 2}, 3},
		{[]int64{1, 2, 3}, 6},

		{[]int64{-1}, -1},
		{[]int64{-1, 1}, 0},
		{[]int64{-1, -2}, -3},
	}

	for _, tc := range testCases {
		got := calculateSum(tc.values)
		if got != tc.want {
			t.Errorf("for %v, got %v; want %v", tc.values, got, tc.want)
		}
	}
}

type floatTestCases []struct {
	values []int64
	want   float64
}

func TestStatsMean(t *testing.T) {
	testCases := floatTestCases{
		{[]int64{0}, 0},
		{[]int64{1}, 1},
		{[]int64{1, 2}, 1.5},
		{[]int64{1, 2, 3}, 2},

		{[]int64{-1}, -1},
		{[]int64{-1, 1}, 0},
		{[]int64{-1, -2}, -1.5},
	}
	for _, tc := range testCases {
		got := calculateMean(tc.values)
		if got != tc.want {
			t.Errorf("for %v, got %v; want %v", tc.values, got, tc.want)
		}
	}
}

func TestStatsMedian(t *testing.T) {
	testCases := floatTestCases{
		{[]int64{0}, 0},
		{[]int64{0, 1}, 0.5},
		{[]int64{0, 1, 2}, 1},
		{[]int64{0, 1, 2, 3}, 1.5},

		{[]int64{-1, 0, 1}, 0}, // negative

		{[]int64{2, 1, 0}, 1}, // unsorted
	}
	for _, tc := range testCases {
		got := calculateMedian(tc.values)
		if got != tc.want {
			t.Errorf("for %v, got %v; want %v", tc.values, got, tc.want)
		}
	}

	t.Run("no_side_effect", func(t *testing.T) {
		var (
			values = []int64{2, 1, 0}
			want   = []int64{2, 1, 0}
		)
		_ = calculateMedian(values)
		if !slices.Equal(values, want) {
			t.Errorf("values were mutated, got %v; want %v", values, want)
		}
	})
}

func TestStatsStdDev(t *testing.T) {
	values := []int64{2, 3, 4, 5}
	want := 1.291
	got := calculateStdDev(values)
	if !mostlyEqual(got, want) {
		t.Errorf("for values %v, got %.3f; want %.3f", values, got, want)
	}
}

var (
	jan1 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	jan3 = time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)
	jan4 = time.Date(2000, 1, 4, 0, 0, 0, 0, time.UTC)
	jan5 = time.Date(2000, 1, 5, 0, 0, 0, 0, time.UTC)
)

func TestStatsCounts(t *testing.T) {
	// Made sure bools worked with generic implementation before
	// I realized that [FprintStatsForColumnAsBoolean] doesn't use
	// value-count semantics.  Including for completeness, for now.
	t.Run("bools", func(t *testing.T) {
		testCases := []struct {
			values []bool
			want   []valueCount[bool]
		}{
			{
				[]bool{true, true, false},
				[]valueCount[bool]{{true, 2}, {false, 1}},
			},
			{
				[]bool{true, false, false},
				[]valueCount[bool]{{false, 2}, {true, 1}},
			},
			{
				[]bool{false, true},
				[]valueCount[bool]{{true, 1}, {false, 1}},
			},
			{
				[]bool{true, false},
				[]valueCount[bool]{{true, 1}, {false, 1}},
			},
			{
				[]bool{true},
				[]valueCount[bool]{{true, 1}},
			},
			{
				[]bool{false},
				[]valueCount[bool]{{false, 1}},
			},
		}

		for _, tc := range testCases {
			got := getValueCounts(tc.values, boolCompare)
			if !slices.Equal(got, tc.want) {
				t.Errorf(" for %v\n got %v\nwant %v", tc.values, got, tc.want)
			}
		}
	})

	t.Run("floats", func(t *testing.T) {
		values := []float64{1, 2, 3, 3, 3, 4, 4}
		want := []valueCount[float64]{{3, 3}, {4, 2}, {1, 1}, {2, 1}}

		counts := getValueCounts(values, cmp.Compare)

		if !slices.Equal(counts, want) {
			t.Errorf(" for %v\n got %v\nwant %v", values, counts, want)
		}
	})

	t.Run("ints", func(t *testing.T) {
		values := []int64{1, 2, 3, 3, 3, 4, 4}
		want := []valueCount[int64]{{3, 3}, {4, 2}, {1, 1}, {2, 1}}

		counts := getValueCounts(values, cmp.Compare)

		if !slices.Equal(counts, want) {
			t.Errorf(" for %v\n got %v\nwant %v", values, counts, want)
		}
	})

	t.Run("strings", func(t *testing.T) {
		values := []string{"1", "2", "3", "3", "3", "4", "4"}
		want := []valueCount[string]{{"3", 3}, {"4", 2}, {"1", 1}, {"2", 1}}

		counts := getValueCounts(values, cmp.Compare)

		if !slices.Equal(counts, want) {
			t.Errorf(" for %v\n got %v\nwant %v", values, counts, want)
		}
	})

	t.Run("times", func(t *testing.T) {
		values := []time.Time{jan1, jan2, jan3, jan3, jan3, jan4, jan4}
		want := []valueCount[time.Time]{{jan3, 3}, {jan4, 2}, {jan1, 1}, {jan2, 1}}

		counts := getValueCounts(values, time.Time.Compare)

		if !slices.Equal(counts, want) {
			t.Errorf(" for %v\n got %v\nwant %v", values, counts, want)
		}
	})
}

// mostlyEqual asserts the equality of a and b within .001.
func mostlyEqual(a, b float64) bool { return math.Abs(a-b) < math.Pow10(-3) }

func TestStatsAll(t *testing.T) {
	t.Run("floats", func(t *testing.T) {
		var (
			values = []float64{1, 2, 2, 4, 5, 5, 5, 3}

			min, max, sum float64 = 1, 5, 27

			mean, median, stdDev = 3.375, 3.5, 1.598

			counts = []valueCount[float64]{
				{5, 3}, {2, 2}, {1, 1}, {3, 1}, {4, 1},
			}
		)

		fcs := NewFloatColumnsStats(values)
		fcs.CalculateAllStats()

		if fcs.min != min {
			t.Errorf("min, got %f; want %f", fcs.min, min)
		}
		if fcs.max != max {
			t.Errorf("max, got %f; want %f", fcs.max, max)
		}
		if fcs.sum != sum {
			t.Errorf("sum, got %f; want %f", fcs.sum, sum)
		}
		if !mostlyEqual(fcs.mean, mean) {
			t.Errorf("mean, got %.3f; want %.3f", fcs.mean, mean)
		}
		if !mostlyEqual(fcs.median, median) {
			t.Errorf("median, got %.3f; want %.3f", fcs.median, median)
		}
		if !mostlyEqual(fcs.stdev, stdDev) {
			t.Errorf("stdDev, got %.3f; want %.3f", fcs.stdev, stdDev)
		}
		if !slices.Equal(fcs.valueCounts, counts) {
			t.Errorf("\n for %v\n got %v\nwant %v", values, fcs.valueCounts, counts)
		}
	})

	t.Run("ints", func(t *testing.T) {
		var (
			values = []int64{1, 2, 2, 4, 5, 5, 5, 3}

			min, max, sum int64 = 1, 5, 27

			mean, median, stdDev = 3.375, 3.5, 1.598

			counts = []valueCount[int64]{
				{5, 3}, {2, 2}, {1, 1}, {3, 1}, {4, 1},
			}
		)

		ics := NewIntColumnsStats(values)
		ics.CalculateAllStats()

		if ics.min != min {
			t.Errorf("min, got %d; want %d", ics.min, min)
		}
		if ics.max != max {
			t.Errorf("max, got %d; want %d", ics.max, max)
		}
		if ics.sum != sum {
			t.Errorf("sum, got %d; want %d", ics.sum, sum)
		}
		if !mostlyEqual(ics.mean, mean) {
			t.Errorf("mean, got %.3f; want %.3f", ics.mean, mean)
		}
		if !mostlyEqual(ics.median, median) {
			t.Errorf("median, got %.3f; want %.3f", ics.median, median)
		}
		if !mostlyEqual(ics.stdev, stdDev) {
			t.Errorf("stdDev, got %.3f; want %.3f", ics.stdev, stdDev)
		}
		if !slices.Equal(ics.valueCounts, counts) {
			t.Errorf("\n for %v\n got %v\nwant %v", values, ics.valueCounts, counts)
		}
	})

	t.Run("strings", func(t *testing.T) {
		var (
			values = []string{"a", "b", "b", "c", "dddd", "e", "e", "e"}

			maxLen int = 4

			counts = []valueCount[string]{
				{"e", 3}, {"b", 2}, {"a", 1}, {"c", 1}, {"dddd", 1},
			}
		)

		scs := NewStringColumnsStats(values)
		scs.CalculateAllStats()

		if scs.maxLength != maxLen {
			t.Errorf("maxLen, got %d; want %d", scs.maxLength, maxLen)
		}
		if !slices.Equal(scs.valueCounts, counts) {
			t.Errorf("\n for %v\n got %v\nwant %v", values, scs.valueCounts, counts)
		}
	})

	t.Run("times", func(t *testing.T) {
		var (
			values = []time.Time{
				jan1, jan2, jan2, jan4, jan5, jan5, jan5, jan3,
			}

			min, max = jan1, jan5

			counts = []valueCount[time.Time]{
				{jan5, 3}, {jan2, 2}, {jan1, 1}, {jan3, 1}, {jan4, 1},
			}
		)

		dcs := NewDateColumnsStats(values, "")
		dcs.CalculateAllStats()

		if !dcs.min.Equal(min) {
			t.Errorf("min, got %v; want %v", dcs.min, min)
		}
		if !dcs.max.Equal(max) {
			t.Errorf("max, got %v; want %v", dcs.max, max)
		}

		if !slices.Equal(dcs.valueCounts, counts) {
			strValues := debugTimes(values)
			strGot := debugTimeCounts(dcs.valueCounts)
			strWant := debugTimeCounts(counts)
			t.Errorf("\n for %v\n got %v\nwant %v", strValues, strGot, strWant)
		}
	})
}

const debugLayout = "2" // legible representation of "Jan 2"

func debugTimes(values []time.Time) []string {
	strValues := []string{}
	for _, x := range values {
		s := x.Format(debugLayout)
		strValues = append(strValues, s)
	}
	return strValues
}

func debugTimeCounts(counts []valueCount[time.Time]) []valueCount[string] {
	strCounts := []valueCount[string]{}
	for _, x := range counts {
		s := x.value.Format(debugLayout)
		strCounts = append(strCounts, valueCount[string]{s, x.count})
	}
	return strCounts
}
