package subcmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

// UnmarshalFunc unmarshals jsonData into the target [Streamer]
// for a [TestdataRunner] to run.
type UnmarshalFunc func(jsonData []byte) (Streamer, error)

// TestdataRunner provides a standard means of testing a [Streamer].
// It reads the [golang.org/x/tools/txtar] file at TxtarPath and
// iterates the test files in the archive, calling the Runner-supplied
// FromJSON for each test case to create a [Streamer] and call its
// Run method with input and want files.
//
// TxtarPath refers to a txtar file archive that follows
// this format:
//   - An optional comment, according to the txtar spec.
//   - The first file is named "-- in --" and contains the input
//     text that Runner reads from its io.Reader.
//   - The second file must be named something (not the empty
//     string). If the file's text is not empty it will be
//     interpreted as JSON and used to deserialize the test Runner
//     with an [UnmarshalFunc].  The name of the file becomes the name
//     of the test case.
//   - The third file can be named either "-- want --" or "-- error --".
//     A want-file contains text that must match what Runner writes to its io.Writer.
//     An error-file contains text that must match the string of the expected error returned by the Run method.
//   - Any number of pairs of JSON-file and want/error-file can refer to a previous, single, in-file.
//
// For example, the following txtar file contains a comment, and four test cases:
//   - The lower and upper tests both refer to the first in-file, `Foo`.
//   - The reverse test refers to its own in-file.
//   - The int test expects an error.
//
// stringfuncs.txt:
//
//	-- in --
//	Foo
//	-- lower --
//	{"Func": "lower"}
//	-- want --
//	foo
//	-- upper --
//	{"Func": "upper"}
//	-- want --
//	FOO
//	-- in --
//	Foo🤓
//	-- reverse --
//	{"Func": "reverse"}
//	-- want --
//	🤓ooF
//	-- in --
//	1.0
//	-- int --
//	{"Func": "toInt"}
//	-- error --
//	parsing "1.0": invalid syntax
type TestdataRunner struct {
	TxtarPath string
	FromJSON  UnmarshalFunc

	t *testing.T
}

// NewTestdataRunner returns a new [TestdataRunner].
func NewTestdataRunner(path string, fromJSON UnmarshalFunc, t *testing.T) TestdataRunner {
	return TestdataRunner{path, fromJSON, t}
}

// Run runs the tests.
func (tdr TestdataRunner) Run() {
	fname := filepath.Base(tdr.TxtarPath)
	scName := strings.TrimSuffix(fname, ".txt")

	tdr.t.Run(scName, func(t *testing.T) {
		a, err := txtar.ParseFile(tdr.TxtarPath)
		if err != nil {
			t.Fatal(err)
		}

		var (
			inputCache []byte // cache input for multiple test cases
			i          = 0
		)

		for i < len(a.Files) {
			if a.Files[i].Name == "in" {
				inputCache = []byte(preprocess(a.Files[i].Data))
				i++
			}

			testname := a.Files[i].Name
			data := a.Files[i].Data
			i++
			want := a.Files[i].Data
			wantname := a.Files[i].Name
			i++
			t.Run(testname, func(t *testing.T) {
				want := preprocess(want)
				if len(data) == 0 {
					data = []byte("{}")
				}
				xx, err := tdr.FromJSON(data)
				if err != nil {
					t.Fatal(err)
				}

				r := bytes.NewReader(inputCache)
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

				err = xx.Run(r, buf)

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
						// if quote {
						// 	t.Errorf("\ngot:\n%q\nwant:\n%q", got, want)
						// } else {
						t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
						// }
					}
				}
			})
		}
	})
}

// preprocess preprocesses a txtar file.
func preprocess(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "$\n", "\n") // remove terminal-marking $
	return s
}
