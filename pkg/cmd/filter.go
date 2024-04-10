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

type Operator string

const (
	Ne  Operator = "ne"
	Eq  Operator = "eq"
	Gt  Operator = "gt"
	Gte Operator = "gte"
	Lt  Operator = "lt"
	Lte Operator = "lte"

	Re Operator = "re"
)

type Filter struct {
	Col             int // 1-based index of column to compare
	Operator        Operator
	Value           string
	CaseInsensitive bool // applies to any string comparison
	Exclude         bool // only write non-matches
}

func NewFilter(col int, operator Operator, val string, caseInsensitive, exclude bool) *Filter {
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
	if sc.Operator == Re {
		expr := sc.Value
		if sc.CaseInsensitive {
			expr = fmt.Sprintf("(?i)%s", expr)
		}
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
	case Re:
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

func match(s string, op Operator, val any, it InferredType) bool {
	switch it {
	case StringType:
		switch op {
		case Eq:
			return s == val.(string)
		case Ne:
			return s != val.(string)
		case Lt:
			return s < val.(string)
		case Lte:
			return s <= val.(string)
		case Gt:
			return s > val.(string)
		case Gte:
			return s >= val.(string)
		}
	case NumberType:
		x, _ := toNumber(s)
		switch op {
		case Eq:
			return x == val.(float64)
		case Ne:
			return x != val.(float64)
		case Lt:
			return x < val.(float64)
		case Lte:
			return x <= val.(float64)
		case Gt:
			return x > val.(float64)
		case Gte:
			return x >= val.(float64)
		}
	case TimeType:
		a, _ := toTime(s)
		b := val.(time.Time)
		switch op {
		case Eq:
			return a.Equal(b)
		case Ne:
			return !a.Equal(b)
		case Lt:
			return a.Before(b)
		case Lte:
			return a.Before(b) || a.Equal(b)
		case Gt:
			return a.After(b)
		case Gte:
			return a.After(b) || a.Equal(b)
		}
	case BoolType:
		x, _ := toBool(s)
		switch op {
		case Eq:
			return x == val.(bool)
		case Ne:
			return x != val.(bool)
		default:
			panic(fmt.Errorf("%s not allowed for boolean filter", op))
		}
	}
	return false
}
