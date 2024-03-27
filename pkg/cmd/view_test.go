package cmd

import (
	"fmt"
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
		if got := pad(tc.x, "", tc.it, tc.n); got != tc.want {
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

func TestLinebreaks(t *testing.T) {
	testCases := []struct{ in, want [][]string }{
		{
			in: [][]string{
				{"1", "2\na"}},
			want: [][]string{
				{"1", "2"},
				{"", "a"}},
		},
		{
			in: [][]string{
				{"1", "2\na"},
				{"3\nb", "4"}},
			want: [][]string{
				{"1", "2"},
				{"", "a"},
				{"3", "4"},
				{"b", ""}},
		},
		{
			in: [][]string{
				{"1", "2\na"},
				{"3\nb", "4\nd\ne"}},
			want: [][]string{
				{"1", "2"},
				{"", "a"},
				{"3", "4"},
				{"b", "d"},
				{"", "e"}},
		},
		{
			in: [][]string{
				{"1", "2\na"},
				{"3", "4\nd\ne"}},
			want: [][]string{
				{"1", "2"},
				{"", "a"},
				{"3", "4"},
				{"", "d"},
				{"", "e"}},
		},
	}
	for _, tc := range testCases {
		got := splitLinebreaks(tc.in)
		fmt.Printf("%+q\n", got)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitLinebreaks\ngot\n%+q\nwant\n%+q", got, tc.want)
		}
	}
}
