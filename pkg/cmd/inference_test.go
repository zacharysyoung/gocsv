package cmd

import (
	"reflect"
	"testing"
	"time"
)

func TestInferType(t *testing.T) {
	testCases := []struct {
		s    string
		want inferredType
	}{
		{"-1.0", numberType},
		{"-1", numberType},
		{"0", numberType},
		{"0.0", numberType},
		{"1", numberType},
		{"1.0", numberType},
		{"True", boolType},
		{"true", boolType},
		{"TrUe", boolType},
		{"T", boolType},
		{"false", boolType},
		{"False", boolType},
		{"falSE", boolType},
		{"F", boolType},
		{"2000-1-1", timeType},
		{"2000-01-01", timeType},
		{"1/1/2000", timeType},
		{"01/01/2000", timeType},
	}

	for _, tc := range testCases {
		if got := inferType(tc.s); got != tc.want {
			t.Errorf("inferType(%q) = %v != %v", tc.s, got, tc.want)
		}
	}

}

func TestInferCols(t *testing.T) {
	testCases := []struct {
		name string
		rows [][]string
		cols []int
		want []inferredType
	}{
		{
			name: "single uniform number",
			rows: [][]string{{"1"}, {"1.0"}, {"-0"}},
			cols: nil,
			want: []inferredType{numberType},
		},
		{
			name: "single uniform bool",
			rows: [][]string{{"true"}, {"false"}, {"f"}},
			cols: nil,
			want: []inferredType{boolType},
		},
		{
			name: "single uniform string",
			rows: [][]string{{"a"}, {"b"}, {"🤓"}},
			cols: nil,
			want: []inferredType{stringType},
		},
		{
			name: "single mixed string",
			rows: [][]string{{"1"}, {"a"}, {"1"}},
			cols: nil,
			want: []inferredType{stringType},
		},
		{
			name: "multi mixed string",
			rows: [][]string{
				{"1", "true"},
				{"a", "false"},
				{"1", "0"},
			},
			cols: nil,
			want: []inferredType{
				stringType, stringType,
			},
		},
		{
			name: "specific columns",
			rows: [][]string{
				{"1", "true", "1/10/2000"},
				{"2", "false", "1/11/2000"},
				{"3", "true", "1/12/2000"},
			},
			cols: []int{0, 2},
			want: []inferredType{
				numberType, timeType,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferCols(tc.rows, tc.cols); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\ninferCols(..., %v) for\n%s\ngot:  %v\nwant: %v", tc.cols, rows(tc.rows), got, tc.want)
			}
		})
	}
}

func TestCompareBools(t *testing.T) {
	for _, tc := range []struct {
		a, b bool
		want int
	}{
		{true, false, -1},
		{true, true, 0},
		{false, true, 1},
		{false, false, 0},
	} {
		if got := compareBools(tc.a, tc.b); got != tc.want {
			t.Errorf("compareBools(%t, %t) got %v; want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompare(t *testing.T) {
	var (
		jan1, _ = time.Parse("1/2/2006", "1/1/2000")
		jan2, _ = time.Parse("1/2/2006", "1/2/2000")
	)
	tfmt := func(x any) string {
		return x.(time.Time).Format("1/2/2006")
	}

	for _, tc := range []struct {
		it   inferredType
		a, b any
		want int
	}{
		{boolType, true, false, -1},
		{boolType, true, true, 0},
		{boolType, false, false, 0},
		{boolType, false, true, 1},

		{numberType, 1.0, 2.0, -1},
		{numberType, 2.0, 2.0, 0},
		{numberType, 2.0, 1.0, 1},

		{timeType, jan1, jan2, -1},
		{timeType, jan2, jan2, 0},
		{timeType, jan2, jan1, 1},

		{stringType, "a", "b", -1},
		{stringType, "b", "b", 0},
		{stringType, "b", "a", 1},
	} {
		if got := compare1(tc.a, tc.b, tc.it); got != tc.want {
			switch tc.it {
			case numberType:
				t.Errorf("compare(%g, %g) got %d; want %d", tc.a, tc.b, got, tc.want)
			case boolType:
				t.Errorf("compare(%t, %t) got %d; want %d", tc.a, tc.b, got, tc.want)
			case stringType:
				t.Errorf("compare(%q, %q) got %d; want %d", tc.a, tc.b, got, tc.want)
			case timeType:
				t.Errorf("compare(%s, %s) got %d; want %d", tfmt(tc.a), tfmt(tc.b), got, tc.want)
			}
		}
	}
}
