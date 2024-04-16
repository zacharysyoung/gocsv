package subcmd

import (
	"reflect"
	"testing"
)

func TestExclude(t *testing.T) {
	testCases := []struct {
		cols, excludes, want []int
	}{
		{
			cols:     []int{1, 2, 3, 4},
			excludes: []int{1},
			want:     []int{2, 3, 4},
		},
		{
			cols:     []int{1, 2, 3, 4},
			excludes: []int{2, 3},
			want:     []int{1, 4},
		},
		{
			cols:     []int{1, 2, 3, 4},
			excludes: []int{3, 1, 2}, // excludes can be in any order
			want:     []int{4},
		},
		{
			cols:     []int{1, 2, 3, 4},
			excludes: []int{}, // shouldn't happen in practice, but works in principle
			want:     []int{1, 2, 3, 4},
		},
	}
	for _, tc := range testCases {
		if got := exclude(tc.cols, tc.excludes); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("excludeCols(%v, %v) = %v; want %v", tc.cols, tc.excludes, got, tc.want)
		}
	}
}
