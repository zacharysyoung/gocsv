package subcmd

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

var quoteflag = flag.Bool("quote", false, "print errors with quoted rows instead of pretty-printed")

type testSubCommander interface {
	SubCommander

	fromJSON([]byte) error // "zero out" the subcommand then configure it from JSON
}

var subcommands = map[string]testSubCommander{
	"clean":  &Clean{},
	"filter": &Filter{},
	"select": &Select{},
	"sort":   &Sort{},
}

func TestCmds(t *testing.T) {
	const suffix = ".txt"
	files, err := filepath.Glob("testdata/*" + suffix)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		fname := filepath.Base(file)
		scName := strings.TrimSuffix(fname, suffix)
		if _, ok := subcommands[scName]; !ok {
			t.Fatalf("found test file %s, but no subcommand %s", fname, scName)
		}

		t.Run(scName, func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}

			// A cmdname-archive contains a least one input-file followed by pairs
			// of files for each test case.
			// A test-case pair is a JSON file (named for the test case) and
			// a want-file.
			// Subsequent test cases use the previous input until another input-file
			// is found.

			var (
				cache []byte // cache input for multiple test cases
				i     = 0
			)

			for i < len(a.Files) {
				if a.Files[i].Name == "in" {
					cache = []byte(preprocess(a.Files[i].Data))
					i++
				}

				testname := a.Files[i].Name
				data := a.Files[i].Data
				i++
				wantb := a.Files[i].Data
				wantname := a.Files[i].Name
				i++
				t.Run(testname, func(t *testing.T) {
					want := preprocess(wantb)
					sc := subcommands[scName]

					if len(data) > 0 {
						if err := sc.fromJSON(data); err != nil {
							t.Fatal(err)
						}
					}
					if err := sc.CheckConfig(); err != nil {
						t.Fatal(err)
					}

					r := bytes.NewReader(cache)
					buf1, buf2 := &bytes.Buffer{}, &bytes.Buffer{}

					switch scName {
					case "clean":
						io.Copy(buf1, r)
					default:
						normalizeCSV(r, buf1)
					}

					err := sc.Run(buf1, buf2)

					switch {
					case err != nil:
						switch wantname {
						case "err":
							got := err.Error()
							want := strings.TrimSpace(want)
							if got != want {
								t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
							}
						default:
							t.Fatal(err)
						}
					default:
						switch scName {
						case "clean":
							io.Copy(buf1, buf2)
						default:
							viewCSV(buf2, buf1)
						}
						got := buf1.String()
						if got != want {
							if *quoteflag {
								t.Errorf("\ngot:\n%q\nwant:\n%q", got, want)
							} else {
								t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
							}
						}
					}
				})
			}
		})
	}
}

func TestRowsStringer(t *testing.T) {
	rows := rows{
		{"Col1", "Col2"},
		{"foo", "12345"},
		{"barbaz", "2.0"},
	}
	want := `
[   Col1,  Col2
     foo, 12345
  barbaz,   2.0 ]`
	want = strings.TrimPrefix(want, "\n")

	if got := fmt.Sprint(rows); got != want {
		t.Errorf("\ngot\n%q\nwant\n%q", got, want)
	}
}

func TestBase0Cols(t *testing.T) {
	for _, tc := range []struct {
		in, want []int
	}{
		{[]int{1, 2}, []int{0, 1}},
		{[]int{1, 1}, []int{0, 0}},
	} {
		if got := Base0Cols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("rebase0(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

// preprocess preprocesses a txtar file.
func preprocess(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "$\n", "\n") // remove terminal-marking $
	return s
}

// normalizeCSV trims leading spaces from visually aligned/padded
// CSV in txtar files.
func normalizeCSV(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	rr.TrimLeadingSpace = true

	recs, err := rr.ReadAll()
	if err != nil {
		return err
	}

	ww := csv.NewWriter(w)
	if err := ww.WriteAll(recs); err != nil {
		return err
	}
	ww.Flush()
	return ww.Error()
}

// viewCSV aligns/pads CSV.
func viewCSV(r io.Reader, w io.Writer) error {
	rr := csv.NewReader(r)
	recs, err := rr.ReadAll()
	if err != nil {
		return err
	}

	cols := make([]int, len(recs[0]))
	for i := range recs[0] {
		cols[i] = i + 1
	}

	types := make([]InferredType, 0)
	switch len(recs) {
	default:
		types = InferCols(recs[1:], cols)
	case 1:
		for range recs[0] {
			types = append(types, StringType)
		}
	}

	widths := getColWidths(recs)

	pad := func(x, suf string, n int, it InferredType) string {
		if suf != "" {
			n += len([]rune(suf))
		}
		if it == StringType {
			n *= -1
		}
		return fmt.Sprintf("%*s", n, x+suf)
	}

	const term = "\n"

	sep, comma := "", ","
	for i, x := range recs[0] {
		if i == len(recs[0])-1 {
			comma = ""
		}
		fmt.Fprintf(w, "%s%s", sep, pad(x, comma, widths[i], StringType))
		sep = " "
	}
	fmt.Fprint(w, term)

	for i := 1; i < len(recs); i++ {
		sep, comma = "", ","
		for j, x := range recs[i] {
			if j == len(recs[i])-1 {
				comma = ""
			}
			fmt.Fprintf(w, "%s%s", sep, pad(x, comma, widths[j], types[j]))
			sep = " "
		}
		fmt.Fprint(w, term)
	}

	return nil
}
