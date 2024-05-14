package convert

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func TestConvertFields(t *testing.T) {
	testCases := []struct {
		in   []string
		want [][]string
	}{
		{
			in: []string{
				"c1  c2   c3",
				"a     1  i ",
				"b   2.0  ii",
			},
			want: [][]string{
				{"c1", "c2", "c3"},
				{"a", "1", "i"},
				{"b", "2.0", "ii"},
			}},
	}

	for _, tc := range testCases {
		r := strings.NewReader(strings.Join(tc.in, "\n"))
		got, err := convertFields(r)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("convertFields\n got: %v\nwant: %v", got, tc.want)
		}
	}
}

func TestConvertMarkdown(t *testing.T) {
	testCases := []struct {
		in   []string
		want [][]string
		err  error
	}{

		{
			in: []string{
				"| c1 | c2  | c3 |",
			},
			want: nil,
			err:  errNoMarkdownTable,
		},
		{
			in: []string{
				"| c1 | c2  | c3 |",
				"| -- | --: | -- |",
				"| a  |   1 | i  |",
			},
			want: [][]string{
				{"c1", "c2", "c3"},
				{"a", "1", "i"},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		r := strings.NewReader(strings.Join(tc.in, "\n"))
		got, err := convertMarkdown(r)
		if !errors.Is(err, tc.err) {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("convertMarkdown\n got: %q\nwant: %q", got, tc.want)
		}
	}
}

func fromJSON(data []byte) (subcmd.Runner, error) {
	convert := &Convert{}
	err := json.Unmarshal(data, convert)
	return convert, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/convert.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
