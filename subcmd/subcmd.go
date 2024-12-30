// Package subcmd defines an interface and helper types and
// methods for a subcommand ("subcmd") intended to read and/or
// write some form of tabular data, and usually CSV.
package subcmd

import (
	"errors"
	"fmt"
	"io"
)

// Streamer is an interface for manipulating a stream of data
// between a singular input and a singular output.
type Streamer interface {
	Run(io.Reader, io.Writer) error
}

// FilesReader is an interface for handling multiple input
// streams.
type FilesReader interface {
	Run([]io.Reader, io.Writer) error
}

// ErrNoHeader is the error returned by a Runner if its CSV reader
// doesn't find at least a header.
var ErrNoHeader = errors.New("no data")

// Base0Cols turns the friendly 1-based indexes in cols to
// 0-based.
func Base0Cols(cols []int) (newCols []int) {
	for _, x := range cols {
		newCols = append(newCols, x-1)
	}
	return
}

// Base1Cols returns friendly 1-based indexes for all columns
// in header.
func Base1Cols(header []string) (newCols []int) {
	for i := range header {
		newCols = append(newCols, i+1)
	}
	return
}

// A Range holds 1-based column indexes in one of two forms:
//   - a single column index, e.g., [2] for the second column
//   - an open range with a start index and -1, e.g., [3 -1] for
//     all columns from the third column on
type Range []int

// Range2 holds 1-based column indexes that can specificy
// a single index {a:>0,b=0}, or a range of indexes.
//
// A range can be ascending (a>b) or descending (b>a).
// A single value of -1 can be specified for a or b (not both)
// to mean the last column in the header:
// {a:1, b:-1} means all columns in the header;
// {a:-1, b:1} means all columns, reversed.
type Range2 struct{ a, b int }

func SingleIndex(a int) (Range2, error) {
	r := Range2{}
	if a < 1 {
		return r, ErrBadRange
	}
	r.a = a
	return r, nil
}

func RangeOfIndexes(a, b int) (Range2, error) {
	r := Range2{}
	if a == 0 || b == 0 || (a == -1 && b == -1) {
		return r, ErrBadRange
	}
	r.a = a
	r.b = b
	return r, nil
}

func goodSingle(a, b, n int) bool {
	return (a > 0 && b == 0) && a <= n
}

func goodRange(a, b, n int) bool {
	switch {
	case a == b:
		return false
	case a == 0 || b == 0:
		return false
	case a > n || b > n:
		return false
	case a == -1 && b == -1:
		return false
	default:
		return true
	}
}

func Finalize2(ranges []Range2, header []string) (finalCols []int, err error) {
	n := len(header)
	for _, r := range ranges {
		a, b := r.a, r.b
		switch {
		case goodSingle(a, b, n):
			fmt.Printf("single a=%d b=%d\n", a, b)
			finalCols = append(finalCols, a)
		case goodRange(a, b, n):
			fmt.Printf("range a=%d b=%d\n", a, b)
			if a == -1 {
				a = n
			}
			if b == -1 {
				b = n
			}

			if a < b {
				fmt.Println("  ascending")
				for i := a; i <= b; i++ {
					finalCols = append(finalCols, i)
				}
			} else {
				fmt.Println("  descending")
				for i := a; i >= b; i-- {
					finalCols = append(finalCols, i)
				}
			}
		default:
			return nil, fmt.Errorf("bad range %v", r)
		}
	}
	return finalCols, nil
}

// Ranges hold any number of [Range] things and provides the
// convenience method [Ranges.Finalize].  Empty or nil ranges
// have no meaning.
type Ranges []Range

// ErrBadRange represents some kind of problem when a bad [Ranges]
// object called [Ranges.Finalize].
var ErrBadRange = errors.New("bad range")

// Finalize expands ranges into a final, flat list of indexes,
// checking header to make sure all specified indexes are in
// bounds. It returns [ErrBadRange] with some specifics for any
// kind of problem.
//
// A subcmd that uses Ranges might benefit by calling this method.
func (ranges Ranges) Finalize(header []string) (newCols []int, err error) {
	if len(ranges) == 0 {
		err = fmt.Errorf("%w: empty or nil ranges", ErrBadRange)
		return
	}

	n := len(header) // 1-based
Error:
	for _, x := range ranges {
		err = fmt.Errorf("%w: %v must be either single-index [i] or range [i -1], , with i>=1 and i<=len(header)=%d", ErrBadRange, x, n)

		a, b := 0, 0
		switch len(x) {
		default:
			break Error
		case 1:
			a = x[0]
			if a < 1 || a > n {
				break Error
			}
			b = x[0]
		case 2:
			a, b = x[0], x[1]
			if a < 1 || a > n || b != -1 {
				break Error
			}
			b = n
		}
		err = nil

		for i := a; i <= b; i++ {
			newCols = append(newCols, i)
		}
	}
	return
}
