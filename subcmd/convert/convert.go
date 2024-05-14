package subcmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"strings"

	md "github.com/zacharysyoung/rsc_markdown"
)

// Convert converts some non-CSV formats to CSV.
type Convert struct {
	Fields, Markdown bool
}

func NewConvert(fields, markdown bool) *Convert {
	return &Convert{
		Fields:   fields,
		Markdown: markdown,
	}
}

func (xx *Convert) fromJSON(p []byte) error {
	*xx = Convert{}
	return json.Unmarshal(p, xx)
}

func (xx *Convert) CheckConfig() error {
	return nil
}

func (xx *Convert) Run(r io.Reader, w io.Writer) error {
	ww := csv.NewWriter(w)

	var (
		Rows [][]string
		err  error
	)
	switch {
	case xx.Fields:
		Rows, err = convertFields(r)
	case xx.Markdown:
		Rows, err = convertMarkdown(r)
	default:
		panic("no valid conversion type specified")
	}

	if err != nil {
		return err
	}
	for _, row := range Rows {
		ww.Write(row)
	}

	ww.Flush()
	return ww.Error()
}

func convertFields(r io.Reader) ([][]string, error) {
	scanner := bufio.NewScanner(r)
	Rows := make([][]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		Rows = append(Rows, strings.Fields(line))
	}
	if err := scanner.Err(); err != nil {
		return Rows, err
	}
	return Rows, nil
}

var errNoMarkdownTable = errors.New("could not find Markdown table")

func convertMarkdown(r io.Reader) ([][]string, error) {
	var Rows [][]string = nil // couldn't get empty slices to compare equably in test, so nil

	b, err := io.ReadAll(r)
	if err != nil {
		return Rows, err
	}

	p := md.Parser{Table: true}
	doc := p.Parse(string(b))

	var (
		tbl *md.Table
		ok  bool
	)
	for _, x := range doc.Blocks {
		if tbl, ok = x.(*md.Table); ok {
			break
		}
	}
	if !ok {
		return Rows, errNoMarkdownTable
	}

	var (
		buf = &bytes.Buffer{}
		row []string
	)

	for _, x := range tbl.Header {
		buf.Reset()
		x.Inline[0].PrintText(buf)
		row = append(row, buf.String())
	}
	Rows = append(Rows, row)

	for _, x := range tbl.Rows {
		row = []string{}
		for _, y := range x {
			s := ""
			if len(y.Inline) > 0 {
				buf.Reset()
				y.Inline[0].PrintText(buf)
				s = buf.String()
			}
			row = append(row, s)
		}
		Rows = append(Rows, row)
	}

	return Rows, nil
}
