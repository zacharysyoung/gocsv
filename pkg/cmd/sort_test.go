package cmd

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	type cols []int
	hdr := rows{{"C1", "C2", "C3"}}

	for _, tc := range []struct {
		rows rows
		cols cols
		want rows
	}{
		{
			rows: rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols: cols{0},
			want: rows{
				{"1", "b"},
				{"2", "c"},
				{"2", "a"},
			},
		},
		{
			rows: rows{
				{"3", "b"},
				{"1", "b"},
				{"2", "a"},
			},
			cols: cols{1},
			want: rows{
				{"2", "a"},
				{"3", "b"},
				{"1", "b"},
			},
		},
		{
			rows: rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols: cols{0, 1},
			want: rows{
				{"1", "b"},
				{"2", "a"},
				{"2", "c"},
			},
		},
	} {
		// prepend header for inferCols
		in := append(hdr, tc.rows...)
		sort2(in, tc.cols)
		// chop header
		in = in[1:]

		if !reflect.DeepEqual(in, tc.want) {
			t.Errorf("sort1(\n%s, %v)\ngot\n%s\nwant\n%s", tc.rows, tc.cols, in, tc.want)
		}
	}

}
