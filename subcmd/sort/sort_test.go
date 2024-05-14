package sort

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func TestSort(t *testing.T) {
	type cols []int

	for _, tc := range []struct {
		rows  subcmd.Rows
		cols  cols
		order int
		want  subcmd.Rows
	}{
		{
			rows: subcmd.Rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1},
			order: 1,
			want: subcmd.Rows{
				{"1", "b"},
				{"2", "c"},
				{"2", "a"},
			},
		},
		{
			rows: subcmd.Rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1},
			order: -1,
			want: subcmd.Rows{
				{"2", "c"},
				{"2", "a"},
				{"1", "b"},
			},
		},
		{
			rows: subcmd.Rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1, 2},
			order: 1,
			want: subcmd.Rows{
				{"1", "b"},
				{"2", "a"},
				{"2", "c"},
			},
		},
		{
			rows: subcmd.Rows{
				{"2", "b"},
				{"2", "c"},
				{"1", "b"},
			},
			cols:  cols{2, 1},
			order: 1,
			want: subcmd.Rows{
				{"1", "b"},
				{"2", "b"},
				{"2", "c"},
			},
		},
		{
			rows: subcmd.Rows{
				{"1", "b"},
				{"2", "b"},
				{"2", "c"},
			},
			cols:  cols{2, 1},
			order: -1,
			want: subcmd.Rows{
				{"2", "c"},
				{"2", "b"},
				{"1", "b"},
			},
		},
	} {
		// copy input
		in := append(subcmd.Rows{}, tc.rows...)
		sort(in, tc.cols, tc.order)
		if !reflect.DeepEqual(in, tc.want) {
			t.Errorf("sort(\n%s, %v,%d)\ngot\n%s\nwant\n%s", tc.rows, tc.cols, tc.order, in, tc.want)
		}
	}
}

func fromJSON(data []byte) (subcmd.Runner, error) {
	sort := &Sort{}
	err := json.Unmarshal(data, sort)
	return sort, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/sort.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
