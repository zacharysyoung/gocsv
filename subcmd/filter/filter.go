package filter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	subcmd "github.com/zacharysyoung/gocsv/subcmd"
)

type Operator string

const (
	Ne  Operator = "ne"  // not-equal
	Eq  Operator = "eq"  // equal
	Gt  Operator = "gt"  // greater-than
	Gte Operator = "gte" // greater-than-or-equal-to
	Lt  Operator = "lt"  // less-than
	Lte Operator = "lte" // less-than-or-equal-to

	Re Operator = "re" // regular expression
)

// Filter reads the input CSV record-by-record, compares a specified field
// in each record to a reference value, and writes matches to the output CSV.
//
// The comparison can be made with any of the mathematical equality
// and inequality operators for all of GoCSV's inferred types.  The
// special regular expression operator can match strings.
type Filter struct {
	// Col is the 1-based index of the column (field) in each
	// record to evaluate.
	Col int
	// Operator is one of the mathematical operators Eq, Ne, Gt,
	// Gte, Lt, Lte, or the regular-expression matcher Re.
	Operator Operator
	// Value is the reference value to compare each record's field
	// value against.
	Value string

	// CaseInsensitive makes any string comparison case-insensitive.
	CaseInsensitive bool
	// Exclude inverts the filter, writing non-matches.
	Exclude bool
	// NoInference forces the reference value and each record's field
	// value to be string.
	NoInference bool
}

func NewFilter(col int, operator Operator, val string) *Filter {
	return &Filter{
		Col:      col,
		Operator: operator,
		Value:    val,
	}
}

func (subcmd *Filter) fromJSON(p []byte) error {
	*subcmd = Filter{}
	return json.Unmarshal(p, subcmd)
}

func (xx *Filter) Run(r io.Reader, w io.Writer) error {
	var (
		reMatcher *regexp.Regexp
		err       error
	)
	if xx.Operator == Re {
		expr := xx.Value
		if xx.CaseInsensitive {
			expr = fmt.Sprintf("(?i)%s", expr)
		}
		if reMatcher, err = regexp.Compile(expr); err != nil {
			return err
		}
	}

	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return errors.New("no data")
		}
		return err
	}
	ww.Write(header)

	col := subcmd.Base0Cols([]int{xx.Col})[0]
	var record []string
	switch xx.Operator {
	case Re:
		for {
			if record, err = rr.Read(); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			matched := reMatcher.MatchString(record[col])
			if xx.Exclude {
				matched = !matched
			}
			if matched {
				if err = ww.Write(record); err != nil {
					return err
				}
			}
		}
	default:
		var (
			val any
			it  subcmd.InferredType
		)
		switch xx.NoInference {
		case true:
			it = subcmd.String
			val = xx.Value
		case false:
			val, it = subcmd.Infer(xx.Value)
		}

		if it == subcmd.String && xx.CaseInsensitive {
			val = strings.ToLower(val.(string))
		}

		for i := 1; ; i++ {
			if record, err = rr.Read(); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			m, err := match(record[col], xx.Operator, val, it, xx.CaseInsensitive, xx.Exclude)
			if err != nil {
				_, sit := subcmd.Infer(record[col])
				return fmt.Errorf("evaluating row %d cell %d, could not compare %s %q to %s %v", i, col+1, sit, record[col], it, val)
			}
			if m {
				if err = ww.Write(record); err != nil {
					return err
				}
			}
		}
	}

	ww.Flush()
	return ww.Error()
}

// match returns the result of the inequality expression of the inferred
// value of s compared with operator op to the reference value v,
// applying the supplementary conditions of a lower-case comparison,
// and/or then negating the match.
//
// If lower is specified, s will be lower-cased before comparing to
// an assumed already lower-cased v.
func match(s string, op Operator, v any, it subcmd.InferredType, lower, negate bool) (bool, error) {
	_match := func() (bool, error) {
		switch it {
		case subcmd.String:
			if lower {
				s = strings.ToLower(s)
			}
			switch op {
			case Eq:
				return s == v.(string), nil
			case Ne:
				return s != v.(string), nil
			case Lt:
				return s < v.(string), nil
			case Lte:
				return s <= v.(string), nil
			case Gt:
				return s > v.(string), nil
			case Gte:
				return s >= v.(string), nil
			}
		case subcmd.Number:
			x, err := subcmd.ToNumber(s)
			switch op {
			case Eq:
				return x == v.(float64), err
			case Ne:
				return x != v.(float64), err
			case Lt:
				return x < v.(float64), err
			case Lte:
				return x <= v.(float64), err
			case Gt:
				return x > v.(float64), err
			case Gte:
				return x >= v.(float64), err
			}
		case subcmd.Time:
			a, err := subcmd.ToTime(s)
			b := v.(time.Time)
			switch op {
			case Eq:
				return a.Equal(b), err
			case Ne:
				return !a.Equal(b), err
			case Lt:
				return a.Before(b), err
			case Lte:
				return a.Before(b) || a.Equal(b), err
			case Gt:
				return a.After(b), err
			case Gte:
				return a.After(b) || a.Equal(b), err
			}
		case subcmd.Bool:
			x, err := subcmd.ToBool(s)
			switch op {
			case Eq:
				return x == v.(bool), err
			case Ne:
				return x != v.(bool), err
			default:
				panic(fmt.Errorf("%s not allowed for boolean filter", op))
			}
		}
		panic(fmt.Errorf("did not evaluate %s %s %v.(%T) as %s", s, op, v, v, it))
	}

	m, err := _match()
	if err != nil {
		return false, err
	}

	if negate {
		m = !m
	}

	return m, nil
}
