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
			sc: &Filter{Col: 1, Operator: Eq, Value: "a", Exclude: false},
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
			sc: &Filter{Col: 1, Operator: Eq, Value: "a", Exclude: true},
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
			sc: &Filter{Col: 1, Operator: Gt, Value: "a", Exclude: false},
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
			sc: &Filter{Col: 1, Operator: Gt, Value: "a", Exclude: true},
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
		op   Operator
		val  any
		it   InferredType
		want bool
	}

	testCases := []testCase{
		{"1", Eq, "1", StringType, true},
		{"1", Ne, "1", StringType, false},

		{"a", Eq, "a", StringType, true},
		{"a", Eq, "A", StringType, false},

		{"1", Lt, "A", StringType, true},
		{"A", Lt, "a", StringType, true},
		{"1", Lt, "a", StringType, true},

		{"1.0", Eq, 1.0, NumberType, true},
		{"1", Eq, 1.0, NumberType, true},

		{"2000-01-02", Eq, date(2000, 1, 2), TimeType, true},
		{"2000-01-01", Ne, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", Lte, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", Gte, date(2000, 1, 2), TimeType, true},
		{"2000-01-01", Lt, date(2000, 1, 2), TimeType, true},
		{"2000-01-03", Gt, date(2000, 1, 2), TimeType, true},
		{"2000-01-02", Gt, date(2000, 1, 2), TimeType, false},

		{"true", Eq, true, BoolType, true},
		{"true", Ne, true, BoolType, false},
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
		{"false", Lt, false, BoolType, false},
		{"false", Gt, false, BoolType, false},
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
