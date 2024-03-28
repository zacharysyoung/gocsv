package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

var quoteflag = flag.Bool("quote", false, "print errors with quoted rows instead of pretty-printed")

func TestCmds(t *testing.T) {
	const suffix = "_cmd.txt"
	files, err := filepath.Glob("testdata/*" + suffix)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		cmdname := strings.TrimSuffix(filepath.Base(file), suffix)
		if cmdname != "view" {
			continue
		}
		cmd, ok := Commands[cmdname]
		if !ok {
			t.Fatalf("could get not Command %s", cmdname)
		}

		t.Run(cmdname, func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}

			// A cmdname-archive contains a least one input-file followed by pairs
			// of files for each test case.
			// A test-case pair is a cmd-flags file (named for the test case) and
			// a want-file.
			// Subsequent test cases use the previous input until another input-file
			// is found.

			var (
				buf []byte // cache input for multiple test cases
				i   = 0
			)
			for i < len(a.Files) {
				if a.Files[i].Name == "input" {
					buf = a.Files[i].Data
					i++
				}

				testname := a.Files[i].Name
				cmdflags := a.Files[i].Data
				i++
				wantb := a.Files[i].Data
				i++
				t.Run(testname, func(t *testing.T) {
					want := preprocess(wantb)
					args := toArgs(cmdflags)

					r := bytes.NewReader(buf)
					w := &bytes.Buffer{}
					cmd.Run(r, w, args...)
					got := w.String()
					if got != want {
						if *quoteflag {
							t.Errorf("\ngot:\n%q\nwant:\n%q", got, want)
						} else {
							t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
						}
					}
				})
			}
		})
	}
}

func preprocess(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "$\n", "\n") // remove terminal-marking $
	return s
}

func toArgs(b []byte) []string {
	s := string(b)
	s = strings.TrimSpace(s)
	ss := strings.Split(s, " ")
	return ss
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
