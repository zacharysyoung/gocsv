package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"html"
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

/*
@ggrothendieck, my question was getting at this behavior: if we set up the replacer with this regex and template:
- `\d+`
- `{{ printf "%05d" (atoi $0) }}`

then we get, `a98 x123` → `a00098 x00123`.

Set it up like:
- `\d+(\d)`
- `{{ printf "%05d" (atoi $1) }}`

we get, `a98 x123` → `a90008 x12003`.

I've played around with something similar on [regexr.com/8d0nd](https://regexr.com/8d0nd)

Being able to
*/

func Test_templateReplacerFunc(t *testing.T) {
	testCases := []struct {
		re    string
		templ string
		field string

		want string
	}{
		// single match, no submatch
		{`\d+`, `{{ printf "%04s" $0 }}`, "ab1", "ab0001"},

		// single match, with submatches (equivalent to above)
		{`([a-z]+)(\d+)`, `{{ printf "%s%04s" $1 $2 }}`, "ab1", "ab0001"},

		// multiple matches, with submatches
		{`([a-z]+)(\d+)`, `{{ printf "%s%04s" $1 $2 }}`, "ab1 yz23", "ab0001 yz0023"},

		// multiple matches, with submatches
		{`\d+(\d)`, `{{ printf "%04s" $1 }}`, "ab1 yz23", "ab1 yz0003"},

		// no matches
		{`\d+`, `{{ printf "%02s" $0 }}`, "ab yz", "ab yz"},

		// specified non-existent submatch in template
		// native template functions fail because submatchData was never initialized
		{`\d+`, `{{ printf "%02s"  $1 }}`, "ab1 yz23", "ab%!s(<nil>) yz%!s(<nil>)"},

		// sprig's atoi treats the same problem differently and errors-out
		// {`\d+`, `{{ printf "%02d"  (atoi $1) }}`, "ab1 yz23", "ab yz"},

		//
		//
		// user-submitted, specific, but already covered by previous cases
		//{`(\d+)(\D+)(\d)(.*)`, `{{ printf "%05s%si%04s%s" $1 $2 $3 $4 }}`, "abc12def13xyz14", "abc00012defi00013xyz14"},
	}

	for _, tc := range testCases {
		re := regexp.MustCompile(tc.re)
		f := templateReplacerFunc(re, tc.templ)
		got := f(tc.field)

		// don't want to specify html-escaped values in want
		got = html.UnescapeString(got)

		if got != tc.want {
			t.Errorf("\n   re %s\ntempl %s\nfield %s\n----- ------\n  got %s\n want %s", tc.re, tc.templ, tc.field, got, tc.want)
		}
	}
}

// test parity between -repl and -templ when field values
// don't support the submatches in replacement/template
