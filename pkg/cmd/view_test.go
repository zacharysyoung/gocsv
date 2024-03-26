package cmd

import (
	"reflect"
	"testing"
)

func TestPad(t *testing.T) {
	testCases := []struct {
		x    string
		it   inferredType
		n    int
		want string
	}{
		{"foo", stringType, 5, "foo  "},
		{"foo", numberType, 5, "  foo"},
		{"foo", boolType, 5, "  foo"},
		{"foo", timeType, 5, "  foo"},
	}

	for _, tc := range testCases {
		if got := pad(tc.x, tc.it, tc.n); got != tc.want {
			t.Errorf("pad(%q, %s, %d) = %q != %q", tc.x, tc.it, tc.n, got, tc.want)
		}
	}
}

func TestCapWidth(t *testing.T) {
	testCases := []struct {
		in   rows
		maxw int
		want rows
	}{
		{
			in: rows{
				{"123456"},
				{"1234567"},
				{"12345678"},
				{"123456789"},
				{"1234567890"},
			},
			maxw: 8,
			want: rows{
				{"123456"},
				{"1234567"},
				{"12345678"},
				{"12345..."},
				{"12345..."},
			},
		},
	}

	for _, tc := range testCases {
		truncateCells(tc.in, tc.maxw)
		if !reflect.DeepEqual(tc.in, tc.want) {
			t.Errorf("capColWidths(..., %d)\ngot\n%s\nwant\n%s", tc.maxw, tc.in, tc.want)
		}
	}
}
