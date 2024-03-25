package cmd

import (
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

func inferCols(rows [][]string) []inferredType {
	types := make([]inferredType, len(rows[0]))

	for i, x := range rows[0] {
		types[i] = inferType(x)
	}
	if len(rows) == 1 {
		return types
	}

	var t inferredType
	for _, row := range rows[1:] {
		for i, x := range row {
			if types[i] != stringType {
				t = inferType(x)
				if t != types[i] {
					types[i] = stringType
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
