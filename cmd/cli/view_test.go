package main

import (
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
)

func TestPad(t *testing.T) {
	testCases := []struct {
		x, suf string
		it     subcmd.InferredType
		n      int
		want   string
	}{
		{"foo", "", subcmd.StringType, 5, "foo  "},
		{"foo", "", subcmd.NumberType, 5, "  foo"},
		{"foo", "", subcmd.BoolType, 5, "  foo"},
		{"foo", "", subcmd.TimeType, 5, "  foo"},

		{"foo", ",", subcmd.StringType, 5, "foo,  "},
		{"foo", ",", subcmd.NumberType, 5, "  foo,"},
	}

	for _, tc := range testCases {
		if got := pad(tc.x, tc.suf, tc.n, tc.it); got != tc.want {
			t.Errorf("pad(%q, %q, %s, %d) = %q; want %q", tc.x, tc.suf, tc.it, tc.n, got, tc.want)
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
	testCases := []struct {
		in, want [][]string
		newlines newlines
	}{
		{
			in: [][]string{
				{"1", "2\na"}},
			want: [][]string{
				{"1", "2"},
				{"", "a"}},
			newlines: newlines{1: nil},
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
			newlines: newlines{1: nil, 3: nil},
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
			newlines: newlines{1: nil, 3: nil, 4: nil},
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
			newlines: newlines{1: nil, 3: nil, 4: nil},
		},
	}
	for _, tc := range testCases {
		got, newlines := splitLinebreaks(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitLinebreaks\ngot\n%+q\nwant\n%+q", got, tc.want)
		}
		if !reflect.DeepEqual(newlines, tc.newlines) {
			t.Errorf("splitLinebreak(%q)\ngot  %v\nwant %v", tc.in, newlines, tc.newlines)
		}
	}
}
