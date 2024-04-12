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
		s             string
		op            Operator
		val           any
		it            InferredType
		lower, invert bool

		want bool
	}

	testCases := []testCase{
		{s: "1", op: Eq, val: "1", it: StringType, want: true},
		{s: "1", op: Ne, val: "1", it: StringType, want: false},

		{s: "a", op: Eq, val: "a", it: StringType, want: true},
		{s: "a", op: Eq, val: "A", it: StringType, want: false},

		{s: "1", op: Lt, val: "A", it: StringType, want: true},
		{s: "A", op: Lt, val: "a", it: StringType, want: true},
		{s: "1", op: Lt, val: "a", it: StringType, want: true},

		{s: "1.0", op: Eq, val: 1.0, it: NumberType, want: true},
		{s: "1", op: Eq, val: 1.0, it: NumberType, want: true},

		{s: "2000-01-02", op: Eq, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-01", op: Ne, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-02", op: Lte, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-02", op: Gte, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-01", op: Lt, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-03", op: Gt, val: date(2000, 1, 2), it: TimeType, want: true},
		{s: "2000-01-02", op: Gt, val: date(2000, 1, 2), it: TimeType, want: false},

		{s: "true", op: Eq, val: true, it: BoolType, want: true},
		{s: "true", op: Ne, val: true, it: BoolType, want: false},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_%s_%v", tc.s, tc.op, tc.val)
		t.Run(name, func(t *testing.T) {
			got := match(tc.s, tc.op, tc.val, tc.it, tc.lower, tc.invert)
			if got != tc.want {
				t.Errorf("match(%s, %s, %v, %s) = %t; want %t",
					tc.s, tc.op, tc.val, tc.it, got, tc.want)
			}
		})
	}

	for _, tc := range []testCase{
		{s: "false", op: Lt, val: false, it: BoolType, want: false},
		{s: "false", op: Gt, val: false, it: BoolType, want: false},
	} {
		name := fmt.Sprintf("%s_%s_%v", tc.s, tc.op, tc.val)
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("match(%s, %s, %v, %s) should have paniced, but didn't", tc.s, tc.op, tc.val, tc.it)
				}
			}()
			got := match(tc.s, tc.op, tc.val, tc.it, tc.lower, tc.invert)
			if got != tc.want {
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
