package subcmd

import (
	"errors"
	"reflect"
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
