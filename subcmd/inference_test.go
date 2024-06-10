package subcmd

import (
	"os"
	"reflect"
	"testing"
	"time"
)

var jan1 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func TestInfer(t *testing.T) {
	testCases := []struct {
		s string
		t InferredType
		v any
	}{
		{"-1.0", Number, -1.0},
		{"-1", Number, -1.0},
		{"0", Number, 0.0},
		{"0.0", Number, 0.0},
		{"1", Number, 1.0},
		{"1.0", Number, 1.0},

		{"True", Bool, true},
		{"true", Bool, true},
		{"TrUe", Bool, true},
		{"T", Bool, true},
		{"false", Bool, false},
		{"False", Bool, false},
		{"falSE", Bool, false},
		{"F", Bool, false},

		{"2000-1-1", Time, jan1},
		{"2000-01-01", Time, jan1},
		{"1/1/2000", Time, jan1},
		{"01/01/2000", Time, jan1},
	}
	for _, tc := range testCases {
		v, tt := Infer(tc.s)
		if tt != tc.t || v != tc.v {
			t.Errorf("Infer(%s) = %v, %s; want %v, %s", tc.s, v, tt, tc.v, tc.t)
		}
	}
}

func TestCustomLayout(t *testing.T) {
	var (
		newLayout = "Jan. 2, 2006"
		sTime     = "Jan. 1, 2000"
	)

	v, it := Infer(sTime)
	if it != String || v != sTime {
		t.Errorf("before adding %s=%q, Infer(%q) = %q, %s; want %q, %s", CSV_LAYOUTS, newLayout, sTime, v, it, sTime, String)
	}

	os.Setenv(CSV_LAYOUTS, newLayout)
	loadNewLayouts()

	v, it = Infer(sTime)
	if it != Time || v != jan1 {
		t.Errorf("after adding %s=%q, Infer(%q) = %q, %s; want %q, %s", CSV_LAYOUTS, newLayout, sTime, v, it, jan1, Time)
	}
}

func TestNoEmptyLayout(t *testing.T) {
	for _, x := range layouts {
		if x == "" {
			t.Errorf("found empty layout string in %q", layouts)
		}
	}
}

func TestInferCols(t *testing.T) {
	type cols []int
	type types []InferredType
	testCases := []struct {
		name string
		rows [][]string
		cols cols
		want types
	}{
		{
			name: "single uniform number",
			rows: [][]string{
				{"1"},
				{"1.0"},
				{"-0"},
			},
			cols: cols{1},
			want: types{Number},
		},
		{
			name: "single uniform bool",
			rows: [][]string{
				{"true"},
				{"false"},
				{"f"},
			},
			cols: cols{1},
			want: types{Bool},
		},
		{
			name: "single uniform string",
			rows: [][]string{
				{"a"},
				{"b"},
				{"🤓"},
			},
			cols: cols{1},
			want: types{String},
		},
		{
			name: "single mixed string",
			rows: [][]string{
				{"1"},
				{"a"},
				{"1"},
			},
			cols: cols{1},
			want: types{String},
		},
		{
			name: "multi mixed string",
			rows: [][]string{
				{"1", "true"},
				{"a", "false"},
				{"1", "0"},
			},
			cols: cols{1, 2},
			want: types{
				String, String,
			},
		},
		{
			name: "specific columns",
			rows: [][]string{
				{"1", "true", "1/10/2000"},
				{"2", "false", "1/11/2000"},
				{"3", "true", "1/12/2000"},
			},
			cols: cols{1, 3},
			want: types{
				Number, Time,
			},
		},
		{
			name: "datetime mixed layouts",
			rows: [][]string{
				{"1/10/2000"},
				{"2000-1-1"},
				{"2000-01-01"},
				{"01/12/2000"},
			},
			cols: cols{1},
			want: types{
				Time,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := InferCols(tc.rows, tc.cols); !reflect.DeepEqual(types(got), tc.want) {
				t.Errorf("\ninferCols(..., %v) for\n%s\ngot:  %v\nwant: %v", tc.cols, tc.rows, got, tc.want)
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
		if got := CompareBools(tc.a, tc.b); got != tc.want {
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
		it   InferredType
		a, b any
		want int
	}{
		{Bool, true, false, -1},
		{Bool, true, true, 0},
		{Bool, false, false, 0},
		{Bool, false, true, 1},

		{Number, 1.0, 2.0, -1},
		{Number, 2.0, 2.0, 0},
		{Number, 2.0, 1.0, 1},

		{Time, jan1, jan2, -1},
		{Time, jan2, jan2, 0},
		{Time, jan2, jan1, 1},

		{String, "a", "b", -1},
		{String, "b", "b", 0},
		{String, "b", "a", 1},
	} {
		if got := Compare1(tc.a, tc.b, tc.it); got != tc.want {
			switch tc.it {
			case Number:
				t.Errorf("compare(%g, %g) got %d; want %d", tc.a, tc.b, got, tc.want)
			case Bool:
				t.Errorf("compare(%t, %t) got %d; want %d", tc.a, tc.b, got, tc.want)
			case String:
				t.Errorf("compare(%q, %q) got %d; want %d", tc.a, tc.b, got, tc.want)
			case Time:
				t.Errorf("compare(%s, %s) got %d; want %d", tfmt(tc.a), tfmt(tc.b), got, tc.want)
			}
		}
	}
}
