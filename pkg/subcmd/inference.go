package subcmd

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type InferredType int

const (
	NumberType InferredType = iota
	BoolType
	TimeType // date and datetime
	StringType
)

func (it InferredType) String() string {
	switch it {
	default:
		panic(fmt.Errorf("bad inferredType value: %#v", it))
	case NumberType:
		return "Number"
	case BoolType:
		return "Bool"
	case TimeType:
		return "Datetime"
	case StringType:
		return "String"
	}
}

func inferType(x string) InferredType {
	switch {
	default:
		return StringType
	case isNumber(x):
		return NumberType
	case isBool(x):
		return BoolType
	case isTime(x):
		return TimeType
	}
}

// InferCols infers the types of 1-based cols of recs; panics
// if recs or cols is empty.
func InferCols(recs [][]string, cols []int) []InferredType {
	switch {
	case len(recs) == 0:
		panic(errors.New("empty recs"))
	case len(cols) == 0:
		panic(errors.New("empty cols"))
	}

	cols = Base0Cols(cols)

	types := make([]InferredType, len(cols))
	for i, xi := range cols {
		types[i] = inferType(recs[0][xi])
	}

	if len(recs) == 1 {
		return types
	}

	var t InferredType
	for i := 1; i < len(recs); i++ {
		for j, jx := range cols {
			if types[j] != StringType {
				t = inferType(recs[i][jx])
				if t != types[j] {
					types[j] = StringType
				}
			}
		}
	}

	return types
}

func isNumber(x string) bool {
	_, err := toNumber(x)
	return err == nil
}

func toNumber(x string) (float64, error) {
	return strconv.ParseFloat(x, 64)
}

func isBool(x string) bool {
	_, err := toBool(x)
	return err == nil
}

func toBool(x string) (bool, error) {
	x = strings.ToLower(x)
	if x == "true" || x == "false" || x == "t" || x == "f" {
		return x == "true" || x == "t", nil
	}
	return false, errors.New("not a bool")
}

func compareBools(a, b bool) int {
	if a && !b {
		return -1
	}
	if !a && b {
		return 1
	}
	return 0
}

var layouts = []string{
	"2006-1-2",
	"2006-01-02",
	"1/2/2006",
	"01/02/2006",
}

func isTime(x string) bool {
	_, err := toTime(x)
	return err == nil
}

func toTime(x string) (time.Time, error) {
	for _, layout := range layouts {
		if t, err := time.Parse(layout, x); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("not a time")
}

func compare1(a, b any, it InferredType) int {
	switch it {
	case BoolType:
		return compareBools(a.(bool), b.(bool))
	case NumberType:
		return cmp.Compare(a.(float64), b.(float64))
	case TimeType:
		return a.(time.Time).Compare(b.(time.Time))
	default:
		return cmp.Compare(a.(string), b.(string))
	}
}

func compare2(a, b string, it InferredType) int {
	switch it {
	case BoolType:
		x, _ := toBool(a)
		y, _ := toBool(b)
		return compareBools(x, y)
	case NumberType:
		x, _ := toNumber(a)
		y, _ := toNumber(b)
		return cmp.Compare(x, y)
	case TimeType:
		x, _ := toTime(a)
		y, _ := toTime(b)
		return x.Compare(y)
	default:
		return cmp.Compare(a, b)
	}

}
