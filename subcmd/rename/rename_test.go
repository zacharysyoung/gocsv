package rename

import (
	"errors"
	"reflect"
	"testing"
)

func TestRename(t *testing.T) {
	header := []string{"A", "B", "C", "D"}
	testCases := []struct {
		cols        []int
		names, want []string
		err         error
	}{
		{[]int{1}, []string{"a"}, []string{"a", "B", "C", "D"}, nil},
		{[]int{1, 4}, []string{"a", "d"}, []string{"a", "B", "C", "d"}, nil},

		{[]int{1, 2}, []string{"a"}, nil, errWrongCounts},
	}

	for _, tc := range testCases {
		got, err := rename(header, tc.cols, tc.names)
		if !errors.Is(err, tc.err) {
			t.Fatalf("unexpected error %v; want %v", err, tc.err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("rename(..., %v, %v) = %v; want %v", tc.cols, tc.names, got, tc.want)
		}
	}
}

func TestReplace(t *testing.T) {
	type (
		hdr  []string
		cols []int
	)
	testCases := []struct {
		header    hdr
		cols      cols
		sre, repl string
		want      hdr
		err       error
	}{
		{hdr{"foo", "foo"}, cols{1}, "foo", "", hdr{"", "foo"}, nil},
		{hdr{"foo", "foo"}, cols{1, 2}, "foo", "", hdr{"", ""}, nil},
		{hdr{"foobar", "fizzbaz"}, cols{1, 2}, "(.+)ba.", "$1", hdr{"foo", "fizz"}, nil},
		{hdr{"foobar", "fizzbaz"}, cols{1, 2}, "ba.", "", hdr{"foo", "fizz"}, nil},

		{hdr{"foobar", "fizzbaz"}, cols{1, 2}, "(.+ba.", "$1", nil, errors.New("some error")},
	}

	for _, tc := range testCases {
		got, err := replace(tc.header, tc.cols, tc.sre, tc.repl)
		if err == nil && tc.err != nil {
			t.Fatalf("got nil error; want %v", tc.err)
		}
		if !reflect.DeepEqual(hdr(got), tc.want) {
			t.Errorf("replace(%v, %v, %q, %q) = %v; want %v", tc.header, tc.cols, tc.sre, tc.repl, got, tc.want)
		}
	}
}
