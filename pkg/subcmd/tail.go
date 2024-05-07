package subcmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Tail reads the last n-number rows of the input CSV.
type Tail struct {
	N       int
	FromTop bool
}

func NewTail(n int, fromTop bool) *Tail {
	return &Tail{
		N:       n,
		FromTop: fromTop,
	}
}

func (sc *Tail) fromJSON(p []byte) error {
	*sc = Tail{}
	return json.Unmarshal(p, sc)
}

func (sc *Tail) CheckConfig() error {
	return nil
}

func (sc *Tail) Run(r io.Reader, w io.Writer) error {
	if sc.N < 1 {
		panic(fmt.Errorf("N = %d; N must be greater than 0", sc.N))
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

	switch sc.FromTop {
	case true:
		err = tailFromTop(rr, ww, sc.N)
	case false:
		err = tailFromBottom(rr, ww, sc.N)
	}
	if err != nil {
		return err
	}

	ww.Flush()
	return ww.Error()
}

// tailFromBottom reads from r, ignoring records up to the n-th
// record, then starts writing to w.
func tailFromBottom(r *csv.Reader, w *csv.Writer, n int) error {
	var (
		buf = make([][]string, n)
		i   = 0
	)
	for ; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf = append(buf[1:], record)
	}
	n = len(buf)
	if i < n {
		n = i
	}

	w.WriteAll(buf[len(buf)-n:])

	return nil
}

// tailFromTop reads from r, dumps records up to n, then
// writes records to w, starting at n.
func tailFromTop(r *csv.Reader, w *csv.Writer, n int) error {
	for i := 0; i < n-1; i++ {
		_, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	for {
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
