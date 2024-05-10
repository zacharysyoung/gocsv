package subcmd

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	type cols []int

	for _, tc := range []struct {
		rows  rows
		cols  cols
		order int
		want  rows
	}{
		{
			rows: rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1},
			order: 1,
			want: rows{
				{"1", "b"},
				{"2", "c"},
				{"2", "a"},
			},
		},
		{
			rows: rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1},
			order: -1,
			want: rows{
				{"2", "c"},
				{"2", "a"},
				{"1", "b"},
			},
		},
		{
			rows: rows{
				{"2", "c"},
				{"1", "b"},
				{"2", "a"},
			},
			cols:  cols{1, 2},
			order: 1,
			want: rows{
				{"1", "b"},
				{"2", "a"},
				{"2", "c"},
			},
		},
		{
			rows: rows{
				{"2", "b"},
				{"2", "c"},
				{"1", "b"},
			},
			cols:  cols{2, 1},
			order: 1,
			want: rows{
				{"1", "b"},
				{"2", "b"},
				{"2", "c"},
			},
		},
		{
			rows: rows{
				{"1", "b"},
				{"2", "b"},
				{"2", "c"},
			},
			cols:  cols{2, 1},
			order: -1,
			want: rows{
				{"2", "c"},
				{"2", "b"},
				{"1", "b"},
			},
		},
	} {
		// copy input
		in := append(rows{}, tc.rows...)
		sort(in, tc.cols, tc.order)
		if !reflect.DeepEqual(in, tc.want) {
			t.Errorf("sort(\n%s, %v,%d)\ngot\n%s\nwant\n%s", tc.rows, tc.cols, tc.order, in, tc.want)
		}
	}
}
