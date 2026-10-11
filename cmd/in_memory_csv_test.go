package cmd

import (
	"math"
	"slices"
	"testing"
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
	want := 1.29
	got := calculateStdDev(values)
	if math.Abs(got-want) > 0.001 {
		t.Errorf("for values %v, got %.2f; want %.2f", values, got, want)
	}
}

func TestStatsCounts(t *testing.T) {
	testCases := []struct {
		values []string
		wantK  []string
		wantV  []int
	}{
		{
			[]string{"a", "a", "b"},

			[]string{"a", "b"},
			[]int{2, 1},
		},
		{
			[]string{"a", "b", "b"},

			[]string{"b", "a"},
			[]int{2, 1},
		},
		{
			[]string{"a", "a", "b", "c", "c", "c"},

			[]string{"c", "a", "b"},
			[]int{3, 2, 1},
		},
	}

	for _, tc := range testCases {
		i := 0
		for k, v := range calculateCounts(tc.values) {
			if tc.wantK[i] != k || tc.wantV[i] != v {
				t.Fatalf("for i=%d; got %v:%d; want %v:%d", i, k, v, tc.wantK[i], tc.wantV[i])
			}
			i++
		}
	}
}
