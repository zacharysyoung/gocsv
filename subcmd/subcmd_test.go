package subcmd

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRowsStringer(t *testing.T) {
	Rows := Rows{
		{"Col1", "Col2"},
		{"foo", "12345"},
		{"barbaz", "2.0"},
	}
	want := `
[   Col1,  Col2
     foo, 12345
  barbaz,   2.0 ]`
	want = strings.TrimPrefix(want, "\n")

	if got := fmt.Sprint(Rows); got != want {
		t.Errorf("\ngot\n%q\nwant\n%q", got, want)
	}
}

func TestBase0Cols(t *testing.T) {
	for _, tc := range []struct {
		in, want []int
	}{
		{[]int{1, 2}, []int{0, 1}},
		{[]int{1, 1}, []int{0, 0}},
	} {
		if got := Base0Cols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("rebase0(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestFinalizeCols(t *testing.T) {
	// reference header w/4 columns
	var header = []string{"A", "B", "C", "D"}

	type Groups []ColGroup
	testCases := []struct {
		groups Groups
		want   []int
		err    error
	}{
		// Single columns
		{Groups{{1}, {2}}, []int{1, 2}, nil},
		{Groups{{2}, {1}, {4}}, []int{2, 1, 4}, nil},
		// Basic open ranges
		{Groups{{-1, -1}}, []int{1, 2, 3, 4}, nil},
		{Groups{{1, -1}}, []int{1, 2, 3, 4}, nil},
		{Groups{{-1, 4}}, []int{1, 2, 3, 4}, nil},
		{Groups{{2, -1}}, []int{2, 3, 4}, nil},
		{Groups{{-1, 3}}, []int{1, 2, 3}, nil},
		// Two open ranges that overlap
		{Groups{{-1, 3}, {2, -1}}, []int{1, 2, 3, 2, 3, 4}, nil},
		// No groups means all columns
		{nil, []int{1, 2, 3, 4}, nil},
	}

	for _, tc := range testCases {
		got, err := FinalizeCols(tc.groups, header)
		if err != nil {
			t.Fatalf("expandCols(%v, ...): %v", tc.groups, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("expandCols(%v, ...) = %v; want %v", tc.groups, got, tc.want)
		}
	}
}
