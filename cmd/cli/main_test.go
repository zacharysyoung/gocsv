package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseCols(t *testing.T) {
	testCases := []struct {
		s    string
		want []int
		err  error
	}{
		{``, nil, nil},

		{`1,2`, []int{1, 2}, nil},
		{`1-3`, []int{1, 2, 3}, nil},
		{`3-1`, []int{3, 2, 1}, nil},

		{`1,3-5,2`, []int{1, 3, 4, 5, 2}, nil},

		{`1-`, nil, errOpenendedRange},
	}
	for _, tc := range testCases {
		got, err := parseCols(tc.s)
		if !errors.Is(err, tc.err) {
			t.Fatalf("parseCols(%q) got error %v; want %v", tc.s, err, tc.err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseCols(%q) got %v; want %v", tc.s, got, tc.want)
		}
	}
}

func TestSplitRange(t *testing.T) {
	testCases := []struct {
		s    string
		want []int
		err  error
	}{
		{"1-4", []int{1, 2, 3, 4}, nil},
		{"9-7", []int{9, 8, 7}, nil},

		{"1-a", nil, errors.New("some parsing error")},
	}

	for _, tc := range testCases {
		got, err := splitRange(tc.s)
		if err != nil && tc.err == nil {
			t.Fatalf("splitRange(%q) got unexpected error %v", tc.s, err)
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitRange(%q) got %v; want %v", tc.s, got, tc.want)
		}
	}
}
