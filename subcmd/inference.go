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
	Number InferredType = iota
	Bool
	Time // date and datetime
	String
)

func (it InferredType) String() string {
	switch it {
	default:
		panic(fmt.Errorf("bad inferredType value: %#v", it))
	case Number:
		return "Number"
	case Bool:
		return "Bool"
	case Time:
		return "Datetime"
	case String:
		return "String"
	}
}

func Infer(s string) (val any, it InferredType) {
	if val, err := ToNumber(s); err == nil {
		return val, Number
	}
	if val, err := ToBool(s); err == nil {
		return val, Bool
	}
	if val, err := ToTime(s); err == nil {
		return val, Time
	}
	return s, String
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

	var (
		it    InferredType
		types = make([]InferredType, len(cols))
	)

	for i, col := range cols {
		_, it = Infer(recs[0][col])
		types[i] = it
	}

	if len(recs) == 1 {
		return types
	}

	for i := 1; i < len(recs); i++ {
		for j, col := range cols {
			if types[j] != String {
				_, it = Infer(recs[i][col])
				if it != types[j] {
					types[j] = String
				}
			}
		}
	}

	return types
}

// func isNumber(x string) bool {
// 	_, err := toNumber(x)
// 	return err == nil
// }

func ToNumber(x string) (float64, error) {
	return strconv.ParseFloat(x, 64)
}

// func isBool(x string) bool {
// 	_, err := toBool(x)
// 	return err == nil
// }

func ToBool(x string) (bool, error) {
	x = strings.ToLower(x)
	if x == "true" || x == "false" || x == "t" || x == "f" {
		return x == "true" || x == "t", nil
	}
	return false, errors.New("not a bool")
}

func CompareBools(a, b bool) int {
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

// func isTime(x string) bool {
// 	_, err := toTime(x)
// 	return err == nil
// }

func ToTime(x string) (time.Time, error) {
	for _, layout := range layouts {
		if t, err := time.Parse(layout, x); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("not a time")
}

func Compare1(a, b any, it InferredType) int {
	switch it {
	case Bool:
		return CompareBools(a.(bool), b.(bool))
	case Number:
		return cmp.Compare(a.(float64), b.(float64))
	case Time:
		return a.(time.Time).Compare(b.(time.Time))
	default:
		return cmp.Compare(a.(string), b.(string))
	}
}

func Compare2(a, b string, it InferredType) int {
	switch it {
	case Bool:
		x, _ := ToBool(a)
		y, _ := ToBool(b)
		return CompareBools(x, y)
	case Number:
		x, _ := ToNumber(a)
		y, _ := ToNumber(b)
		return cmp.Compare(x, y)
	case Time:
		x, _ := ToTime(a)
		y, _ := ToTime(b)
		return x.Compare(y)
	default:
		return cmp.Compare(a, b)
	}

}
