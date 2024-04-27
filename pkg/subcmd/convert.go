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

func (sc *Convert) fromJSON(p []byte) error {
	*sc = Convert{}
	return json.Unmarshal(p, sc)
}

func (sc *Convert) CheckConfig() error {
	return nil
}

func (sc *Convert) Run(r io.Reader, w io.Writer) error {
	ww := csv.NewWriter(w)

	var (
		rows [][]string
		err  error
	)
	switch {
	case sc.Fields:
		rows, err = convertFields(r)
	case sc.Markdown:
		rows, err = convertMarkdown(r)
	default:
		panic("no valid conversion type specified")
	}

	if err != nil {
		return err
	}
	for _, row := range rows {
		ww.Write(row)
	}

	ww.Flush()
	return ww.Error()
}

func convertFields(r io.Reader) ([][]string, error) {
	scanner := bufio.NewScanner(r)
	rows := make([][]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		rows = append(rows, strings.Fields(line))
	}
	if err := scanner.Err(); err != nil {
		return rows, err
	}
	return rows, nil
}

var errNoMarkdownTable = errors.New("could not find Markdown table")

func convertMarkdown(r io.Reader) ([][]string, error) {
	var rows [][]string = nil // couldn't get empty slices to compare equably in test, so nil

	b, err := io.ReadAll(r)
	if err != nil {
		return rows, err
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
		return rows, errNoMarkdownTable
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
	rows = append(rows, row)

	for _, x := range tbl.Rows {
		row = []string{}
		for _, y := range x {
			buf.Reset()
			y.Inline[0].PrintText(buf)
			row = append(row, buf.String())
		}
		rows = append(rows, row)
	}

	return rows, nil
}
