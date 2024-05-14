package head

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Head prints top rows of the input CSV.
type Head struct {
	N          int
	FromBottom bool
}

func NewHead(n int, fromBottom bool) *Head {
	return &Head{
		N:          n,
		FromBottom: fromBottom,
	}
}

func (xx *Head) fromJSON(p []byte) error {
	*xx = Head{}
	return json.Unmarshal(p, xx)
}

func (xx *Head) CheckConfig() error {
	return nil
}

func (xx *Head) Run(r io.Reader, w io.Writer) error {
	if xx.N < 1 {
		panic(fmt.Errorf("N = %d; N must be greater than 0", xx.N))
	}

	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var (
		header []string
		err    error
	)
	header, err = rr.Read()
	if err != nil {
		if err == io.EOF {
			return errors.New("no data")
		}
	}
	ww.Write(header)

	switch xx.FromBottom {
	case true:
		err = headFromBottom(rr, ww, xx.N)
	case false:
		err = headFromTop(rr, ww, xx.N)
	}
	if err != nil {
		return err
	}

	ww.Flush()
	return ww.Error()
}

func headFromTop(r *csv.Reader, w *csv.Writer, n int) error {
	for i := 0; i < n; i++ {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		w.Write(record)
	}
	return nil
}

func headFromBottom(r *csv.Reader, w *csv.Writer, n int) error {
	buf := make([][]string, 0, n)
	for i := 0; ; i++ {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch len(buf) {
		case n:
			w.Write(buf[0])
			buf = append(buf[1:], record)
		default:
			buf = append(buf, record)
		}
	}
	if len(buf) == n {
		w.Write(buf[0])
	}
	return nil
}
