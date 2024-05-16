package rename

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/zacharysyoung/gocsv/subcmd"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Rename changes column names in the header.
type Rename struct {
	// Ranges are the Ranges of columns to be renamed.
	Ranges subcmd.Ranges

	// Names is list of replacement names that matches the order and
	// count of the finalized indexes in Ranges.
	Names []string

	// Regexp is a regexp to match certain column names in ColGroups.
	Regexp string
	// Repl is the replacement string for names matched by Regexp.
	Repl string

	// MakeKeys makes names good to use as keys in an html/template.
	MakeKeys bool
}

func NewRename(ranges []subcmd.Range, names []string, regexp, repl string, keys bool) *Rename {
	return &Rename{Ranges: ranges, Names: names, Regexp: regexp, Repl: repl, MakeKeys: keys}
}

func (xx *Rename) Run(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	ww := csv.NewWriter(w)

	var (
		err error

		header, row []string
		cols        []int
	)

	if header, err = rr.Read(); err != nil {
		if err == io.EOF {
			return subcmd.ErrNoHeader
		}
		return err
	}

	switch len(xx.Ranges) {
	case 0:
		cols = subcmd.Base1Cols(header)
	default:
		if cols, err = xx.Ranges.Finalize(header); err != nil {
			return err
		}
	}

	names, regexp, keys := len(xx.Names) > 0, xx.Regexp != "", xx.MakeKeys
	switch {
	// case names && regexp:
	// 	return fmt.Errorf("got both Names: %v and Regexp: %q; cannot use both", xx.Names, xx.Regexp)
	// case !names && !regexp:
	// 	return fmt.Errorf("got neither Names nor Regexp; must use one")

	case names:
		if header, err = rename(header, cols, xx.Names); err != nil {
			return err
		}
	case regexp:
		if header, err = replace(header, cols, xx.Regexp, xx.Repl); err != nil {
			return err
		}
	case keys:
		header = makeKeys(header, cols)
	}

	ww.Write(header)
	for {
		if row, err = rr.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ww.Write(row)
	}
	ww.Flush()

	return ww.Error()
}

var errWrongCounts = errors.New("cols and names must be the same length")

// rename changes values in header with pairs of 1-based cols and
// names.
func rename(header []string, cols []int, names []string) ([]string, error) {
	if len(names) != len(cols) {
		return nil, fmt.Errorf("%w: %d names != %d cols", errWrongCounts, len(names), len(cols))
	}
	cols = subcmd.Base0Cols(cols)
	for i, x := range cols {
		header[x] = names[i]
	}
	return header, nil
}

// replace changes values in header that match sre by doing a
// regexpReplaceAllString with repl.
func replace(header []string, cols []int, sre, repl string) (hdr []string, err error) {
	var re *regexp.Regexp
	if re, err = regexp.Compile(sre); err != nil {
		return
	}

	hdr = make([]string, len(header))
	copy(hdr, header)
	for _, x := range subcmd.Base0Cols(cols) {
		hdr[x] = re.ReplaceAllString(header[x], repl)
	}
	return
}

// makeKeys sanitizes the names in header for cols so that each
// name can be used as a valid field in an html/template.
func makeKeys(header []string, cols []int) (newHeader []string) {
	newHeader = make([]string, 0, len(cols))
	for _, x := range header {
		newHeader = append(newHeader, sanitizeIdentifier(x))
	}
	return newHeader
}

// Starting with the language-spec definition 'FieldDecl' under,
// https://go.dev/ref/spec#Struct_types, and follow through:
//
//	FieldDecl -> IdentifierList -> identifier -> { letter | unicode_digit } -> unicode_letter
const (
	// https://go.dev/ref/spec#unicode_letter
	unicode_letter = `\p{L}`

	// https://go.dev/ref/spec#unicode_digit
	unicode_digit = `\p{Nd}`

	// https://go.dev/ref/spec#letter, concatenated here because
	// they will end up in character class brackets
	letter = unicode_letter + `_`

	// Inverse of https://go.dev/ref/spec#identifier
	notIdentifier = `[^` + letter + unicode_digit + `]`

	leadingNumber = `^` + unicode_digit

	spaces = `\s+`
)

var (
	reNotIdentifier = regexp.MustCompile(notIdentifier)
	reLeadingNumber = regexp.MustCompile(leadingNumber)
	reSpaces        = regexp.MustCompile(spaces)

	toTitle = cases.Title(language.English, cases.NoLower)
)

func sanitizeIdentifier(s string) string {
	s = reSpaces.ReplaceAllString(s, "_")
	s = reNotIdentifier.ReplaceAllString(s, "")
	s = toTitle.String(s)
	if reLeadingNumber.MatchString(s) {
		s = "_" + s
	}
	return s
}
