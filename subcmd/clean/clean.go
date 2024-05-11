package subcmd

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

// Clean cleans a CSV.
type Clean struct {
	// Trim leading spaces in a field; placed their by some pretty-
	// printers, like View.
	Trim bool
}

func NewClean(trim bool) *Clean {
	return &Clean{
		Trim: trim,
	}
}

func (sc *Clean) fromJSON(p []byte) error {
	*sc = Clean{}
	return json.Unmarshal(p, sc)
}

func (sc *Clean) CheckConfig() error {
	return nil
}

func (sc *Clean) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	rr.TrimLeadingSpace = sc.Trim

	for {
		record, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ww.Write(record)
	}

	ww.Flush()
	return ww.Error()
}
