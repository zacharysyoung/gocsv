package filter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	subcmd "github.com/zacharysyoung/gocsv/subcmd"
)

// simpleMatchTest provides a simpler test of match without
// lower-case or match negation args.
type simpleMatchTest struct {
	s   string
	op  Operator
	val any

	want bool
}

func TestMatch(t *testing.T) {
	// date := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

	t.Run("String", func(t *testing.T) {
		testSimple(t, subcmd.String, []simpleMatchTest{
			{"1", Eq, "1", true},
			{"1", Ne, "2", true},
			{"2", Lt, "20", true},
			{"3", Gt, "20", true},
			{"4", Gte, "3", true},
			{"3", Lte, "4", true},

			{"a", Eq, "B", false},
			{"a", Ne, "a", false},
			{"a", Lt, "A", false},
			{"A", Gt, "a", false},
			{"a", Lte, "A", false},
			{"A", Gte, "a", false},
		})
	})

	t.Run("Number", func(t *testing.T) {
		testSimple(t, subcmd.Number, []simpleMatchTest{
			{"1", Eq, 1.0, true},
			{"1", Ne, 2.0, true},
			{"2", Lt, 3.0, true},
			{"3", Gt, 2.0, true},
			{"2", Lte, 3.0, true},
			{"3", Lte, 3.0, true},
			{"3", Gte, 3.0, true},
			{"4", Gte, 3.0, true},

			{"1.0", Eq, float64(1), true},

			{"1", Ne, 1.0, false},
			{"1", Eq, 2.0, false},
			{"2", Gt, 3.0, false},
			{"3", Lt, 2.0, false},
			{"2", Gte, 3.0, false},
			{"4", Lte, 3.0, false},
		})
	})

	t.Run("Time", func(t *testing.T) {
		var (
			jan1 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			jan2 = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
			jan3 = time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)
		)
		testSimple(t, subcmd.Time, []simpleMatchTest{
			{"2000-01-01", Eq, jan1, true},
			{"2000-01-01", Ne, jan2, true},
			{"2000-01-02", Lt, jan3, true},
			{"2000-01-03", Gt, jan2, true},
			{"2000-01-02", Lte, jan3, true},
			{"2000-01-03", Lte, jan3, true},
			{"2000-01-03", Gte, jan3, true},
			{"2000-01-04", Gte, jan3, true},

			{"2000-01-01", Ne, jan1, false},
			{"2000-01-01", Eq, jan2, false},
			{"2000-01-02", Gt, jan3, false},
			{"2000-01-03", Lt, jan2, false},
			{"2000-01-02", Gte, jan3, false},
			{"2000-01-04", Lte, jan3, false},
		})
	})

	t.Run("Bool", func(t *testing.T) {
		testSimple(t, subcmd.Bool, []simpleMatchTest{
			{"true", Eq, true, true},
			{"true", Ne, false, true},

			{"true", Eq, false, false},
			{"true", Ne, true, false},
		})
	})

	// match should panic for nonsense bool operators
	t.Run("BoolPanic", func(t *testing.T) {
		for _, tc := range []simpleMatchTest{
			{"false", Lt, false, false},
			{"false", Gt, false, false},
			{"false", Lte, false, false},
			{"false", Gte, false, false},
		} {
			name := fmt.Sprintf("%s_%s_%v", tc.s, tc.op, tc.val)
			t.Run(name, func(t *testing.T) {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("bool match `%q %s %v` did not panic; it should have", tc.s, tc.op, tc.val)
					}
				}()
				match(tc.s, tc.op, tc.val, subcmd.Bool, false, false)
			})
		}
	})
}

func testSimple(t *testing.T, it subcmd.InferredType, testCases []simpleMatchTest) {
	for _, tc := range testCases {
		sVal := fmt.Sprintf("%v", tc.val)
		if it == subcmd.Time {
			sVal = sVal[:10]
		}

		name := fmt.Sprintf("%s_%s_%s_is_True", tc.s, tc.op, sVal)
		if !tc.want {
			name = fmt.Sprintf("%s_%s_%s_is_False", tc.s, tc.op, sVal)
		}

		t.Run(name, func(t *testing.T) {
			got, _ := match(tc.s, tc.op, tc.val, it, false, false)
			if got != tc.want {
				t.Errorf("match(%s, %s, %v, %s, false, false) = %t; want %t",
					tc.s, tc.op, tc.val, it, got, tc.want)
			}
		})
	}
}

func fromJSON(data []byte) (subcmd.Streamer, error) {
	filter := &Filter{}
	err := json.Unmarshal(data, filter)
	return filter, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/filter.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
