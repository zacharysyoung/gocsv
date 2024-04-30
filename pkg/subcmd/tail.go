package subcmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
)

// Tail reads the last n-number rows of the input CSV.
type Tail struct {
	N int
}

func NewTail(n int) *Tail {
	return &Tail{
		N: n,
	}
}

func (sc *Tail) fromJSON(p []byte) error {
	*sc = Tail{}
	return json.Unmarshal(p, sc)
}

func (sc *Tail) CheckConfig() error {
	return nil
}

// [[1],[2],[3],[4],[5],[6],[7]]
// n=3
// buf=[_,_,_,]
// step 0: buf=[_,_,1]
// step 1: buf=[_,1,2]
// step 2: buf=[1,2,3]
// step 3: buf=[2,3,4]
// step 4: buf=[3,4,5]
// step 5: buf=[4,5,6]
// step 6: buf=[5,6,7]

func (sc *Tail) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return errors.New("no data")
		}
	}
	ww.Write(header)

	var (
		i   = 0
		n   = sc.N
		buf = make([][]string, n)
	)
	for ; ; i++ {
		record, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		buf = append(buf[1:], record)
	}

	if i < n {
		n = i
	}

	ww.WriteAll(buf[len(buf)-n:])
	ww.Flush()
	return ww.Error()
}
