package cmd

import (
	"reflect"
	"testing"
)

func TestCapWidth(t *testing.T) {
	testCases := []struct {
		in   rows
		maxw int
		want rows
	}{
		{
			in: rows{
				{"1234"},
				{"12345"},
			},
			maxw: 4,
			want: rows{
				{"1234"},
				{"1..."},
			},
		},
		{
			in: rows{
				{"12345"},
				{"123456"},
			},
			maxw: 5,
			want: rows{
				{"12345"},
				{"12..."},
			},
		},
		{
			in: rows{
				{"123456"},
				{"1234567"},
			},
			maxw: 6,
			want: rows{
				{"123456"},
				{"123..."},
			},
		},
	}

	for _, tc := range testCases {
		capColWidths(tc.in, tc.maxw)
		if !reflect.DeepEqual(tc.in, tc.want) {
			t.Errorf("capColWidths(..., %d)\ngot\n%s\nwant\n%s", tc.maxw, tc.in, tc.want)
		}
	}

}
