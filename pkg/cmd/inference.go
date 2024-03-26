package cmd

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type inferredType int

const (
	numberType inferredType = iota
	boolType
	timeType // date and datetime
	stringType
)

func (it inferredType) String() string {
	switch it {
	default:
		panic(fmt.Errorf("bad inferredType value: %#v", it))
	case numberType:
		return "Number"
	case boolType:
		return "Bool"
	case timeType:
		return "Datetime"
	case stringType:
		return "String"
	}
}

func inferType(x string) inferredType {
	switch {
	default:
		return stringType
	case isNumber(x):
		return numberType
	case isBool(x):
		return boolType
	case isTime(x):
		return timeType
	}
}

func inferCols(recs [][]string, cols []int) []inferredType {
	if cols == nil {
		cols = make([]int, len(recs[0]))
		for i := range recs[0] {
			cols[i] = i
		}
	}

	types := make([]inferredType, len(cols))
	for i, xi := range cols {
		types[i] = inferType(recs[0][xi])
	}
	if len(recs) == 1 {
		return types
	}

	var t inferredType
	for i := 1; i < len(recs); i++ {
		for j, jx := range cols {
			if types[j] != stringType {
				t = inferType(recs[i][jx])
				if t != types[j] {
					types[j] = stringType
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

func compare1(a, b any, it inferredType) int {
	switch it {
	case boolType:
		return compareBools(a.(bool), b.(bool))
	case numberType:
		return cmp.Compare(a.(float64), b.(float64))
	case timeType:
		return a.(time.Time).Compare(b.(time.Time))
	default:
		return cmp.Compare(a.(string), b.(string))
	}
}

func compare2(a, b string, it inferredType) int {
	switch it {
	case boolType:
		x, _ := toBool(a)
		y, _ := toBool(b)
		return compareBools(x, y)
	case numberType:
		x, _ := toNumber(a)
		y, _ := toNumber(b)
		return cmp.Compare(x, y)
	case timeType:
		x, _ := toTime(a)
		y, _ := toTime(b)
		return x.Compare(y)
	default:
		return cmp.Compare(a, b)
	}

}
