package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"time"
)

type operator string

const (
	ne  operator = "ne"
	eq  operator = "eq"
	gt  operator = "gt"
	gte operator = "gte"
	lt  operator = "lt"
	lte operator = "lte"

	re operator = "re"
)

type Filter struct {
	Col             int // 1-based index of column to compare
	Operator        operator
	Value           string
	CaseInsensitive bool // applies to any string comparison
	Exclude         bool // only write non-matches
}

func NewFilter(col int, operator operator, val string, caseInsensitive, exclude bool) *Filter {
	return &Filter{
		Col:             col,
		Operator:        operator,
		Value:           val,
		CaseInsensitive: caseInsensitive,
		Exclude:         exclude,
	}
}

func (sc *Filter) fromJSON(p []byte) error {
	*sc = Filter{}
	return json.Unmarshal(p, sc)
}

func (sc *Filter) CheckConfig() error {
	return nil
}

func (sc *Filter) Run(r io.Reader, w io.Writer) error {
	var (
		reMatcher *regexp.Regexp
		err       error
	)
	if sc.Operator == re {
		expr := sc.Value
		if sc.CaseInsensitive {
			expr = fmt.Sprintf("(?i)%s", expr)
		}
		fmt.Println(expr)
		if reMatcher, err = regexp.Compile(expr); err != nil {
			return err
		}
	}

	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	recs, err := rr.ReadAll()
	if err != nil {
		return err
	}

	ww.Write(recs[0])

	recs = sc.filter(recs[1:], reMatcher)

	ww.WriteAll(recs)
	ww.Flush()
	return ww.Error()
}

func (sc *Filter) filter(recs [][]string, reMatcher *regexp.Regexp) [][]string {
	// flip matched if Excluded
	var flip = func(matched bool) bool {
		if sc.Exclude {
			return !matched
		}
		return matched
	}

	col := Base0Cols([]int{sc.Col})[0]
	switch sc.Operator {
	case re:
		for i := len(recs) - 1; i >= 0; i-- {
			matched := reMatcher.MatchString(recs[i][col])
			matched = flip(matched)
			if !matched {
				recs = slices.Delete(recs, i, i+1)
			}
		}
	default:
		val := sc.Value
		if sc.CaseInsensitive {
			val = strings.ToLower(val)
		}
		for i := len(recs) - 1; i >= 0; i-- {
			s := recs[i][col]
			if sc.CaseInsensitive {
				s = strings.ToLower(s)
			}
			matched := match(s, sc.Operator, val, StringType)
			matched = flip(matched)
			if !matched {
				recs = slices.Delete(recs, i, i+1)
			}
		}
	}

	return recs
}

func match(s string, op operator, val any, it InferredType) bool {
	switch it {
	case StringType:
		switch op {
		case eq:
			return s == val.(string)
		case ne:
			return s != val.(string)
		case lt:
			return s < val.(string)
		case lte:
			return s <= val.(string)
		case gt:
			return s > val.(string)
		case gte:
			return s >= val.(string)
		}
	case NumberType:
		x, _ := toNumber(s)
		switch op {
		case eq:
			return x == val.(float64)
		case ne:
			return x != val.(float64)
		case lt:
			return x < val.(float64)
		case lte:
			return x <= val.(float64)
		case gt:
			return x > val.(float64)
		case gte:
			return x >= val.(float64)
		}
	case TimeType:
		a, _ := toTime(s)
		b := val.(time.Time)
		switch op {
		case eq:
			return a.Equal(b)
		case ne:
			return !a.Equal(b)
		case lt:
			return a.Before(b)
		case lte:
			return a.Before(b) || a.Equal(b)
		case gt:
			return a.After(b)
		case gte:
			return a.After(b) || a.Equal(b)
		}
	case BoolType:
		x, _ := toBool(s)
		switch op {
		case eq:
			return x == val.(bool)
		case ne:
			return x != val.(bool)
		default:
			panic(fmt.Errorf("%s not allowed for boolean filter", op))
		}
	}
	return false
}
