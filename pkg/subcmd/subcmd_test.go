package subcmd

import (
	"bytes"
	"flag"
	"fmt"
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
	"convert": &Convert{},
	"clean":   &Clean{},
	"filter":  &Filter{},
	"head":    &Head{},
	"select":  &Select{},
	"sort":    &Sort{},
	"tail":    &Tail{},
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
					buf := &bytes.Buffer{}

					defer func() {
						if err := recover(); err != nil {
							switch wantname {
							case "panic":
								got := fmt.Sprint(err)
								want := strings.TrimSpace(want)
								if got != want {
									t.Errorf("\n got: %s\nwant: %s", got, want)
								}
							default:
								t.Fatal(err)
							}
						}
					}()

					err := sc.Run(r, buf)

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
						got := buf.String()
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
		if got := base0Cols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("rebase0(%v) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestFinalizeCols(t *testing.T) {
	// reference header w/4 columns
	var header = []string{"A", "B", "C", "D"}

	type Groups []ColGroup
	testCases := []struct {
		groups Groups
		want   []int
		err    error
	}{
		// Single columns
		{Groups{{1}, {2}}, []int{1, 2}, nil},
		{Groups{{2}, {1}, {4}}, []int{2, 1, 4}, nil},
		// Basic open ranges
		{Groups{{-1, -1}}, []int{1, 2, 3, 4}, nil},
		{Groups{{1, -1}}, []int{1, 2, 3, 4}, nil},
		{Groups{{-1, 4}}, []int{1, 2, 3, 4}, nil},
		{Groups{{2, -1}}, []int{2, 3, 4}, nil},
		{Groups{{-1, 3}}, []int{1, 2, 3}, nil},
		// Two open ranges that overlap
		{Groups{{-1, 3}, {2, -1}}, []int{1, 2, 3, 2, 3, 4}, nil},
		// No groups means all columns
		{nil, []int{1, 2, 3, 4}, nil},
	}

	for _, tc := range testCases {
		got, err := FinalizeCols(tc.groups, header)
		if err != nil {
			t.Fatalf("expandCols(%v, ...): %v", tc.groups, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("expandCols(%v, ...) = %v; want %v", tc.groups, got, tc.want)
		}
	}
}

// preprocess preprocesses a txtar file.
func preprocess(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "$\n", "\n") // remove terminal-marking $
	return s
}
