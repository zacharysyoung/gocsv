package cmd

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	type rows [][]string
	type testCase struct {
		sc       *Filter
		in, want rows
	}

	for _, tc := range []testCase{
		{
			sc: &Filter{Col: 1, Operator: eq, Value: "a", Exclude: false},
			in: rows{
				{"a"},
				{"b"},
				{"c"},
			},
			want: rows{
				{"a"},
			},
		},
		{
			sc: &Filter{Col: 1, Operator: eq, Value: "a", Exclude: true},
			in: rows{
				{"a"},
				{"b"},
				{"c"},
			},
			want: rows{
				{"b"},
				{"c"},
			},
		},
		{
			sc: &Filter{Col: 1, Operator: gt, Value: "a", Exclude: false},
			in: rows{
				{"a"},
				{"b"},
				{"c"},
			},
			want: rows{
				{"b"},
				{"c"},
			},
		},
		{
			sc: &Filter{Col: 1, Operator: gt, Value: "a", Exclude: true},
			in: rows{
				{"a"},
				{"b"},
				{"c"},
			},
			want: rows{
				{"a"},
			},
		},
	} {
		in, sc := tc.in, tc.sc
		name := fmt.Sprintf("filter %v %d%s%s", in, sc.Col, sc.Operator, sc.Value)
		if sc.Exclude {
			name += " exclude"
		}
		t.Run(name, func(t *testing.T) {
			got := append(rows{}, sc.filter(in, nil)...) // filter returns sliced version of in, so capacity may not match want
			if !reflect.DeepEqual(got[:], tc.want[:len(tc.want)]) {
				t.Errorf("\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	date := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

	type testCase struct {
		s    string
		op   operator
		val  any
		it   InferredType
		want bool
	}

	testCases := []testCase{
		{"1", eq, "1", StringType, true},
		{"1", ne, "1", StringType, false},

		{"a", eq, "a", StringType, true},
		{"a", eq, "A", StringType, false},

		{"1", lt, "A", StringType, true},
		{"A", lt, "a", StringType, true},
		{"1", lt, "a", StringType, true},

		{"1.0", eq, 1.0, NumberType, true},
		{"1", eq, 1.0, NumberType, true},

		{"2000-01-02", eq, date(2000, 1, 2), TimeType, true},
		{"2000-01-01", ne, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", lte, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", gte, date(2000, 1, 2), TimeType, true},
		{"2000-01-01", lt, date(2000, 1, 2), TimeType, true},
		{"2000-01-03", gt, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", gt, date(2000, 1, 2), TimeType, false},

		{"true", eq, true, BoolType, true},
		{"true", ne, true, BoolType, false},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_%s_%v", tc.s, tc.op, tc.val)
		t.Run(name, func(t *testing.T) {
			if got := match(tc.s, tc.op, tc.val, tc.it); got != tc.want {
				t.Errorf("match(%s, %s, %v, %s) = %t; want %t",
					tc.s, tc.op, tc.val, tc.it, got, tc.want)
			}
		})
	}

	for _, tc := range []testCase{
		{"false", lt, false, BoolType, false},
		{"false", gt, false, BoolType, false},
	} {
		name := fmt.Sprintf("%s_%s_%v", tc.s, tc.op, tc.val)
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("match(%s, %s, %v, %s) should have paniced, but didn't", tc.s, tc.op, tc.val, tc.it)
				}
			}()
			if got := match(tc.s, tc.op, tc.val, tc.it); got != tc.want {
				t.Errorf("match(%s, %s, %v, %s) = %t; want %t",
					tc.s, tc.op, tc.val, tc.it, got, tc.want)
			}
		})
	}
}

// func TestMatchRE(t *testing.T) {
// 	testCases := []struct {
// 		s, val string
// 		want   bool
// 	}{
// 		{"foobar", "foo", true},
// 		{"foobar", "f.*", true},
// 	}
// }
