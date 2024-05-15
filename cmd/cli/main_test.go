package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func dumpSlice[T any](s []T) string {
	return fmt.Sprintf("%T len:%d cap: %d %v", s, len(s), cap(s), s)
}

func TestParseCols(t *testing.T) {

	testCases := []struct {
		s    string
		want subcmd.Ranges
		err  error
	}{
		{"", nil, nil},

		{"1,2", subcmd.Ranges{{1}, {2}}, nil},
		{"2,1", subcmd.Ranges{{2}, {1}}, nil},

		{"-3", subcmd.Ranges{{1}, {2}, {3}}, nil},
		{"1-3", subcmd.Ranges{{1}, {2}, {3}}, nil},
		{"3-1", subcmd.Ranges{{3}, {2}, {1}}, nil},

		{"3-", subcmd.Ranges{{3, -1}}, nil},

		{"1,3-5,2", subcmd.Ranges{{1}, {3}, {4}, {5}, {2}}, nil},
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
	someErr := errors.New("some parsing error")
	testCases := []struct {
		s    string
		want _range
		err  error
	}{
		// closed ranges
		{"1-3", _range{1, 3}, nil},
		{"3-1", _range{3, 1}, nil},
		{"1-1", _range{1, 1}, nil},
		// open ranges
		{"-", _range{-1, -1}, nil},
		{"-4", _range{-1, 4}, nil},
		{"4-", _range{4, -1}, nil},
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
