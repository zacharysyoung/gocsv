package subcmd

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestBase0Cols(t *testing.T) {
	for _, tc := range []struct {
		in, want []int
	}{
		{[]int{1}, []int{0}},
		{[]int{1, 1}, []int{0, 0}},
		{[]int{1, 2}, []int{0, 1}},
		{[]int{1, 2, 3}, []int{0, 1, 2}},

		// meaningless, probably bad for caller, but no error
		{[]int{0}, []int{-1}},
	} {
		if got := Base0Cols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Base0Cols(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestBase1Cols(t *testing.T) {
	for _, tc := range []struct {
		in   []string
		want []int
	}{
		{[]string{"A"}, []int{1}},
		{[]string{"A", "B", "C", "D"}, []int{1, 2, 3, 4}},

		// meaningless, probably bad for caller, but no error
		{[]string{}, nil},
	} {
		if got := Base1Cols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Base1Cols(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestFinalize(t *testing.T) {
	// reference header w/4 columns
	var header = []string{"A", "B", "C", "D"}

	testCases := []struct {
		ranges Ranges
		want   []int
		err    error
	}{
		// single indexes
		{Ranges{{1}, {2}}, []int{1, 2}, nil},
		{Ranges{{2}, {1}}, []int{2, 1}, nil},
		// open ranges
		{Ranges{{1, -1}}, []int{1, 2, 3, 4}, nil},
		{Ranges{{2, -1}}, []int{2, 3, 4}, nil},
		{Ranges{{4, -1}}, []int{4}, nil},
		// mixed
		{Ranges{{1}, {2}, {3, -1}}, []int{1, 2, 3, 4}, nil},

		{nil, nil, ErrBadRange},             // meaningless nil ranges
		{Ranges{{0}}, nil, ErrBadRange},     // single index < 1
		{Ranges{{5}}, nil, ErrBadRange},     // single index > len(header)
		{Ranges{{0, -1}}, nil, ErrBadRange}, // open range, index < 1
		{Ranges{{5, -1}}, nil, ErrBadRange}, // open range, index > len(header)
		{Ranges{{1, 1}}, nil, ErrBadRange},  // open range, not ending with -1
	}

	for _, tc := range testCases {
		got, err := tc.ranges.Finalize(header)
		if !errors.Is(err, tc.err) || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%v.Finalize(%v) = %v, %v: want %v, %v", tc.ranges, header, got, err, tc.want, tc.err)
		}
	}
}

func TestGood(t *testing.T) {
	const headerLen = 6

	t.Run("single", func(t *testing.T) {
		for _, tc := range []struct {
			a, b, n int
			want    bool
		}{
			{1, 0, headerLen, true},
			{6, 0, headerLen, true},

			{7, 0, headerLen, false},
			{0, 0, headerLen, false},
			{0, -1, headerLen, false},
			{-1, 0, headerLen, false},
			{1, -1, headerLen, false},
			{-1, 1, headerLen, false},
		} {
			a, b, n, want := tc.a, tc.b, tc.n, tc.want

			if got := goodSingle(a, b, n); got != want {
				t.Errorf("goodSingle(%d %d %d)=%t; want %t", a, b, n, got, want)
			}
		}
	})
	t.Run("range", func(t *testing.T) {
		for _, tc := range []struct {
			a, b, n int
			want    bool
		}{
			{1, 6, headerLen, true},
			{6, 1, headerLen, true},
			{1, -1, headerLen, true},
			{-1, 1, headerLen, true},

			{1, 7, headerLen, false},
			{7, 1, headerLen, false},
			{1, 1, headerLen, false},
			{0, -1, headerLen, false},
			{-1, 0, headerLen, false},
			{-1, -1, headerLen, false},
		} {
			a, b, n, want := tc.a, tc.b, tc.n, tc.want

			if got := goodRange(a, b, n); got != want {
				t.Errorf("goodRange(%d %d %d)=%t; want %t", a, b, n, got, want)
			}
		}
	})
}

func TestFinalize2(t *testing.T) {
	// header w/6 columns
	header := []string{"A", "B", "C", "D", "E", "F"}

	testCases := []struct {
		ranges []Range2
		want   []int
		err    error
	}{
		{
			[]Range2{{1, 0}, {3, 0}, {5, 0}},
			[]int{1, 3, 5},
			nil,
		},
		{
			[]Range2{{5, 0}, {3, 0}, {1, 0}},
			[]int{5, 3, 1},
			nil,
		},
		{
			[]Range2{{1, 3}, {3, 1}},
			[]int{1, 2, 3, 3, 2, 1},
			nil,
		},
		{
			[]Range2{{1, -1}},
			[]int{1, 2, 3, 4, 5, 6},
			nil,
		},
		{
			[]Range2{{-1, 1}},
			[]int{6, 5, 4, 3, 2, 1},
			nil,
		},
		{
			[]Range2{{1, -1}, {-1, 1}},
			[]int{1, 2, 3, 4, 5, 6, 6, 5, 4, 3, 2, 1},
			nil,
		},
		{
			[]Range2{{0, 6}},
			nil,
			errors.New("some error"),
		},
		{
			[]Range2{{-1, 0}},
			nil,
			errors.New("some error"),
		},
		{
			[]Range2{{7, 0}},
			nil,
			errors.New("some error"),
		},
		{
			[]Range2{{1, 7}},
			nil,
			errors.New("some error"),
		},
		{
			[]Range2{{7, -1}},
			nil,
			errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		got, err := Finalize2(tc.ranges, header)
		if (tc.err == nil && err != nil) || (tc.err != nil && err == nil) {
			t.Fatalf("Finalize2(%v, A...E) = _,%v; want _,%v", tc.ranges, err, tc.err)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("Finalize2(%v, A...E) = %v,%v; want %v,%v", tc.ranges, got, err, tc.want, tc.err)
		}
	}
}
