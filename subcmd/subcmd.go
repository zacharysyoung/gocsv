// Package subcmd defines an interface and helper types and
// methods for a subcommand ("subcmd") intended to read and/or
// write some form of tabular data, and usually CSV.
package subcmd

import (
	"errors"
	"fmt"
	"io"
)

// Runner is an interface for manipulating streams of data,
// primarily tabular, and primarily CSV at that.
type Runner interface {
	Run(io.Reader, io.Writer) error
}

// ErrNoHeader is the error returned by a Runner if its CSV reader
// doesn't find a header.
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

// Ranges hold any number of [Range] things and provides the
// convenience method [Ranges.Finalize].  Empty or nil ranges
// have no meaning.
type Ranges []Range

// ErrBadRange represents some kind of problem when a bad [Ranges]
// object called [Ranges.Finalize].
var ErrBadRange = errors.New("bad range")

// Finalize expands ranges into a final, flat list of indexes,
// checking header to make sure all specified indexes are in
// bounds. It returns [ErrBadRange] with some specifices for any
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
