package cmd

import "slices"

func sort2(recs [][]string, cols []int) {
	types := inferCols(recs[1:], cols)
	slices.SortFunc(recs, func(a, b []string) int {
		for i, ix := range cols {
			if x := compare2(a[ix], b[ix], types[i]); x != 0 {
				return x
			}
		}
		return 0
	})
}
