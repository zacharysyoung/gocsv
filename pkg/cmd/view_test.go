package cmd

import (
	"reflect"
	"testing"
)

func TestPad(t *testing.T) {
	testCases := []struct {
		x, suf string
		it     inferredType
		n      int
		want   string
	}{
		{"foo", "", stringType, 5, "foo  "},
		{"foo", "", numberType, 5, "  foo"},
		{"foo", "", boolType, 5, "  foo"},
		{"foo", "", timeType, 5, "  foo"},

		{"foo", ",", stringType, 5, "foo,  "},
		{"foo", ",", numberType, 5, "  foo,"},
	}

	for _, tc := range testCases {
		if got := pad(tc.x, tc.suf, tc.it, tc.n); got != tc.want {
			t.Errorf("pad(%q, %q, %s, %d) = %q != %q", tc.x, tc.suf, tc.it, tc.n, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		for _, tc := range []struct {
			maxw     int
			in, want string
		}{
			{6, "123456", "123456"},
			{5, "123456", "12..."},
			{5, "12345", "12345"},
			{4, "12345", "1..."},
			{4, "1234", "1234"},
		} {
			if got := truncate(tc.in, tc.maxw, 1); got != tc.want {
				t.Errorf("truncate(%s, %d, 1)\ngot  %s\nwant %s", tc.in, tc.maxw, got, tc.want)
			}
		}
	})
	t.Run("multiple lines", func(t *testing.T) {
		for _, tc := range []struct {
			maxh     int
			in, want string
		}{
			{2, "abc\ndef", "abc\ndef"},
			{1, "abc\ndef", "abc..."},
			{3, "abc\ndef\nghi", "abc\ndef\nghi"},
			{2, "abc\ndef\nghi", "abc\ndef..."},
		} {
			if got := truncate(tc.in, -1, tc.maxh); got != tc.want {
				t.Errorf("truncate(%q, -1, %d)\ngot  %q\nwant %q", tc.in, tc.maxh, got, tc.want)
			}
		}
	})
	t.Run("maxw and maxh", func(t *testing.T) {
		for _, tc := range []struct {
			maxw, maxh int
			in, want   string
		}{
			{-1, -1, "abcdef\nghijkl", "abcdef\nghijkl"},
			{6, 1, "abcdef\nghijkl", "abc..."},
			{5, -1, "abcdef\nghijkl", "ab...\ngh..."},
			{5, 1, "abcdef\nghijkl", "ab..."},
			{5, 2, "abcdef\nghijkl", "ab...\ngh..."},
		} {
			if got := truncate(tc.in, tc.maxw, tc.maxh); got != tc.want {
				t.Errorf("truncate(%q, %d, %d)\ngot  %q\nwant %q", tc.in, tc.maxw, tc.maxh, got, tc.want)
			}
		}
	})
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
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitLinebreaks\ngot\n%+q\nwant\n%+q", got, tc.want)
		}
	}
}
