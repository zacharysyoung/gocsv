package cut

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func TestExclude(t *testing.T) {
	header := []string{"a", "b", "c", "d"}
	testCases := []struct {
		cols []int
		want []int
	}{
		{
			cols: []int{1},
			want: []int{2, 3, 4},
		},
		{
			cols: []int{2, 3},
			want: []int{1, 4},
		},
		{
			cols: []int{3, 1, 2}, // excludes can be in any order
			want: []int{4},
		},
		{
			cols: []int{}, // shouldn't happen in practice, but works in principle
			want: []int{1, 2, 3, 4},
		},
	}
	for _, tc := range testCases {
		if got := exclude(tc.cols, header); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("excludeCols(%v, ...) = %v; want %v", tc.cols, got, tc.want)
		}
	}
}

func fromJSON(data []byte) (subcmd.Runner, error) {
	cut := &Cut{}
	err := json.Unmarshal(data, cut)
	return cut, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/cut.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
