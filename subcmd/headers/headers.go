package headers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
)

// Headers prints pairs of Idx/Name for each column in the
// header of the input CSV.
type Headers struct {
	// Print 0-based indexes instead of default 1-based.
	ZeroBased bool
}

func NewHeaders(zeroBased bool) *Headers {
	return &Headers{zeroBased}
}

func (xx *Headers) Run(r io.Reader, w io.Writer) error {
	idx := 1
	if xx.ZeroBased {
		idx = 0
	}

	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	header, err := rr.Read()
	if err != nil {
		if err == io.EOF {
			return errors.New("no data")
		}
		return err
	}

	ww.Write([]string{"Idx", "Name"})
	for _, x := range header {
		ww.Write([]string{fmt.Sprintf("%d", idx), x})
		idx++
	}

	ww.Flush()
	return ww.Error()
}
