package cmd

import (
	"reflect"
	"testing"
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
		want []inferredType
	}{
		{
			name: "single uniform number",
			rows: [][]string{{"1"}, {"1.0"}, {"-0"}},
			want: []inferredType{numberType},
		},
		{
			name: "single uniform bool",
			rows: [][]string{{"true"}, {"false"}, {"f"}},
			want: []inferredType{boolType},
		},
		{
			name: "single uniform string",
			rows: [][]string{{"a"}, {"b"}, {"🤓"}},
			want: []inferredType{stringType},
		},
		{
			name: "single mixed string",
			rows: [][]string{{"1"}, {"a"}, {"1"}},
			want: []inferredType{stringType},
		},
		{
			name: "multi mixed string",
			rows: [][]string{
				{"1", "true"},
				{"a", "false"},
				{"1", "0"},
			},
			want: []inferredType{
				stringType, stringType,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferCols(tc.rows); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("inferCols for\n%s\n got %v\nwant %v", rows(tc.rows), got, tc.want)
			}
		})
	}
}
