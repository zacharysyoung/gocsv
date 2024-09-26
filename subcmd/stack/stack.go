package stack

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/zacharysyoung/gocsv/subcmd"
)

// Stack reads the input CSVs and stacks one on top of the other,
// asserting that all subsequent headers match the initial header.
type Stack struct {
	Paths []string
}

func NewStack() *Stack {
	return new(Stack)
}

func (xx *Stack) fromJSON(p []byte) error {
	panic("not implemented")
}

func (xx *Stack) CheckConfig() error {
	panic("not implemented")
}

func (xx *Stack) Run(readers []io.Reader, w io.Writer) error {
	if len(readers) < 2 {
		panic("expect at least two readers")
	}

	var (
		csvw = csv.NewWriter(w)

		refHeader []string
		err       error
	)
	for i, r := range readers {
		if refHeader, err = stack(r, refHeader, csvw); err != nil {
			return fmt.Errorf("reader #%d: %w", i+1, err)
		}
	}
	csvw.Flush()
	return csvw.Error()
}

func stack(reader io.Reader, refHeader []string, csvw *csv.Writer) (header []string, err error) {
	var row []string

	csvr := csv.NewReader(reader)

	if header, err = csvr.Read(); err != nil {
		if err == io.EOF {
			return nil, subcmd.ErrNoHeader
		}
		return nil, err
	}
	switch {
	case refHeader == nil:
		csvw.Write(header)
	case !slices.Equal(header, refHeader):
		return nil, errors.New("header doesn't match first header")
	}

	for {
		if row, err = csvr.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		csvw.Write(row)
	}

	return header, nil
}
