package subcmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

// UnmarshalFunc unmarshals the byte slice as JSON into a
// *subcmd.Runner.
type UnmarshalFunc func(data []byte) (Runner, error)

// TestdataRunner provides a standard means of testing a [Runner].
type TestdataRunner struct {
	Path     string // path of a Txtar test file.
	FromJSON UnmarshalFunc
	T        *testing.T
}

// NewTestdataRunner returns a new [TestdataRunner].
func NewTestdataRunner(path string, f UnmarshalFunc, t *testing.T) TestdataRunner {
	return TestdataRunner{path, f, t}
}

// Run runs the tests.
func (tdr TestdataRunner) Run() {
	fname := filepath.Base(tdr.Path)
	scName := strings.TrimSuffix(fname, ".txt")

	tdr.T.Run(scName, func(t *testing.T) {
		a, err := txtar.ParseFile(tdr.Path)
		if err != nil {
			t.Fatal(err)
		}

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
				if len(data) == 0 {
					data = []byte("{}")
				}
				xx, err := tdr.FromJSON(data)
				if err != nil {
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
