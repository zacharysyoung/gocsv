package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	gocsv "github.com/aotimme/gocsv/csv"
	"golang.org/x/tools/txtar"
)

func TestRunReplace(t *testing.T) {
	testCases := []struct {
		columnsString   string
		regex           string
		repl            string
		caseInsensitive bool
		rows            [][]string
	}{
		{"String", "Two", "Dos", false, [][]string{
			{"Number", "String"},
			{"1", "One"},
			{"2", "Dos"},
			{"-1", "Minus One"},
			{"2", "Another Dos"},
		}},
		{"String", "^one", "UNO", true, [][]string{
			{"Number", "String"},
			{"1", "UNO"},
			{"2", "Two"},
			{"-1", "Minus One"},
			{"2", "Another Two"},
		}},
	}
	for i, tt := range testCases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			ic, err := NewInputCsv("../test-files/simple-sort.csv")
			if err != nil {
				t.Error("Unexpected error", err)
			}
			toc := new(testOutputCsv)
			sub := new(ReplaceSubcommand)
			sub.Col = tt.columnsString
			sub.Regexp = tt.regex
			sub.repl = tt.repl
			sub.caseInsensitive = tt.caseInsensitive
			sub.RunReplace(ic, toc)
			err = assertRowsEqual(tt.rows, toc.rows)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	path := filepath.Join("testdata", "replace.txt")
	archive, err := txtar.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	files := archive.Files
	if n := len(files) % 3; n != 0 {
		t.Fatalf("got %d files in %s; want multiple of 3", n, path)
	}

	for i := 0; i < len(files); i += 3 {
		for j, prefix := range []string{
			"test: ",
			"in:",
			"want:"} {
			name := files[i+j].Name
			if !strings.HasPrefix(name, prefix) {
				t.Fatalf("for (group %d, file %d) got %q; want %q...", (i/3)+1, j+1, name, prefix)
			}
			files[i+j].Name = strings.TrimPrefix(name, prefix)
		}
		var (
			testFile = files[i]
			inFile   = files[i+1]
			wantFile = files[i+2]
		)

		t.Run(testFile.Name, func(t *testing.T) {
			var sc ReplaceSubcommand

			err := json.Unmarshal(testFile.Data, &sc)
			if err != nil {
				// t.Error(err)
				sc = ReplaceSubcommand{}
				fs := flag.NewFlagSet(sc.Name(), flag.ExitOnError)
				sc.SetFlags(fs)

				args := strings.Fields(trimb(testFile.Data))
				err := fs.Parse(args)
				if err != nil {
					t.Fatalf("could not parse args %q: %v", args, err)
				}
			}

			in := &InputCsv{
				reader: gocsv.NewReader(bytes.NewReader(preprocess(inFile.Data))),
			}
			buf := &bytes.Buffer{}
			out := &OutputCsv{
				csvWriter: csv.NewWriter(buf),
			}

			sc.RunReplace(in, out)

			got := normalize(buf.Bytes())
			want := normalize(wantFile.Data)

			if got != want {
				t.Errorf("\n===  got:\n%s\n=== want:\n%s", got, want)
			}
		})
	}
}

func Test_templateReplacerFunc(t *testing.T) {
	testCases := []struct {
		re    string
		templ string
		field string

		want string
		err  error
	}{
		// single match, no submatch
		{`\d+`, `{{ printf "%02d" (atoi $0) }}`, "yz2", "yz02", nil},

		// single match, with submatches (equivalent to above)
		{`([a-z]+)(\d+)`, `{{ printf "%s%02d" $1 (atoi $2) }}`, "yz2", "yz02", nil},

		// multiple matches, with submatches
		{`\d+`, `{{ printf "%02d" (atoi $0) }}`, "ab1 yz2", "ab01 yz02", nil},

		// no matches
		{`\d+`, `{{ printf "%02d" (atoi $0) }}`, "ab yz", "ab yz", nil},

		// specified non-existent submatch in template
		{`\d+`, `{{ printf "%02d" (atoi $1) }}`, "ab1 yz2", "ab yz", errors.New("at <.Submatch_1>: invalid value")},
	}

	for _, tc := range testCases {
		re := regexp.MustCompile(tc.re)
		f := templateReplacerFunc(re, tc.templ)
		got, err := f(tc.field)

		if err != nil || tc.err != nil {
			switch {
			case err != nil && tc.err == nil:
				t.Errorf("got %v; want <nil>", err)
			case err == nil && tc.err != nil:
				t.Errorf("got <nil>; want ...%v...", tc.err)

			case !strings.Contains(err.Error(), tc.err.Error()):
				t.Errorf("got %v; want ...%v...", err, tc.err)
			}
			continue
		}

		if got != tc.want {
			t.Errorf("(re=%s templ=%s field=%s) got %s; want %s", tc.re, tc.templ, tc.field, got, tc.want)
		}
	}
}
