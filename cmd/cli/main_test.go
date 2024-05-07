package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
)

func dumpSlice[T any](s []T) string {
	return fmt.Sprintf("%T len:%d cap: %d %v", s, len(s), cap(s), s)
}

func TestParseCols(t *testing.T) {
	type Groups []subcmd.ColGroup
	testCases := []struct {
		s    string
		want Groups
		err  error
	}{
		{``, nil, nil},

		{`1,2`, Groups{{1}, {2}}, nil},
		{`2,1`, Groups{{2}, {1}}, nil},
		{`1-3`, Groups{{1, 3}}, nil},

		{`1,3-5,2`, Groups{{1}, {3, 5}, {2}}, nil},

		{`-3`, Groups{{-1, 3}}, nil},
		{`3-`, Groups{{3, -1}}, nil},
	}
	for _, tc := range testCases {
		got, err := parseCols(tc.s)
		if !errors.Is(err, tc.err) {
			t.Fatalf("parseCols(%q) got error %v; want %v", tc.s, err, tc.err)
		}
		if !reflect.DeepEqual(Groups(got), tc.want) {
			t.Errorf("parseCols(%q) got %v; want %v", tc.s, got, tc.want)
		}
	}
}

func TestSplitRange(t *testing.T) {
	someErr := errors.New("some parsing error")
	testCases := []struct {
		s    string
		want subcmd.ColGroup
		err  error
	}{
		// closed ranges
		{"1-4", subcmd.ColGroup{1, 4}, nil},
		{"4-1", subcmd.ColGroup{4, 1}, nil},
		{"1-1", subcmd.ColGroup{1, 1}, nil},
		// open ranges
		{"-", subcmd.ColGroup{-1, -1}, nil},
		{"-4", subcmd.ColGroup{-1, 4}, nil},
		{"4-", subcmd.ColGroup{4, -1}, nil},
		// bad ranges
		{"--", nil, someErr},
		{"1-a", nil, someErr},
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
