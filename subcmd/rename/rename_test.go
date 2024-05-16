package rename

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

type (
	hdr  []string
	cols []int
)

func TestRename(t *testing.T) {
	header := []string{"A", "B", "C", "D"}
	testCases := []struct {
		cols        cols
		names, want []string
		err         error
	}{
		{cols{1}, []string{"a"}, []string{"a", "B", "C", "D"}, nil},
		{cols{1, 4}, []string{"a", "d"}, []string{"a", "B", "C", "d"}, nil},

		{cols{1, 2}, []string{"a"}, nil, errWrongCounts},
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

func TestSanitizeIdentifier(t *testing.T) {
	testCases := []struct {
		name string
		want string
	}{
		{"foo", "Foo"},
		{"foo bar", "Foo_bar"},
		{"[foo]", "Foo"},
		{"FOo", "FOo"},
		{"123", "_123"},
	}
	for _, tc := range testCases {
		if got := sanitizeIdentifier(tc.name); got != tc.want {
			t.Errorf("sanitizeIdentifier(%q) = %v; want %v", tc.name, got, tc.want)
		}
	}
}

func fromJSON(data []byte) (subcmd.Runner, error) {
	rename := &Rename{}
	err := json.Unmarshal(data, rename)
	return rename, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/rename.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
