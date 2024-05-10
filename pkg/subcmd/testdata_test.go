package subcmd_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
	"github.com/zacharysyoung/gocsv/pkg/subcmd/cut"
	"golang.org/x/tools/txtar"
)

var quoteflag = flag.Bool("quote", false, "print errors with quoted rows instead of pretty-printed")

type testSubCommander interface {
	subcmd.SubCommander
}

var subcommands = map[string]testSubCommander{
	"convert": &subcmd.Convert{},
	"clean":   &subcmd.Clean{},
	"filter":  &subcmd.Filter{},
	"head":    &subcmd.Head{},
	"cut":     &cut.Cut{},
	"sort":    &subcmd.Sort{},
	"tail":    &subcmd.Tail{},
}

func fromJSON(name string, data []byte) (subcmd.SubCommander, error) {
	if len(data) == 0 {
		data = []byte("{}")
	}
	var (
		sc  subcmd.SubCommander
		err error
	)
	switch name {
	case "convert":
		sc = &subcmd.Convert{}
		err = json.Unmarshal(data, sc)
	case "clean":
		sc = &subcmd.Clean{}
		err = json.Unmarshal(data, sc)
	case "cut":
		sc = &cut.Cut{}
		err = json.Unmarshal(data, sc)
	case "filter":
		sc = &subcmd.Filter{}
		err = json.Unmarshal(data, sc)
	case "head":
		sc = &subcmd.Head{}
		err = json.Unmarshal(data, sc)
	case "sort":
		sc = &subcmd.Sort{}
		err = json.Unmarshal(data, sc)
	case "tail":
		sc = &subcmd.Tail{}
		err = json.Unmarshal(data, sc)
	default:
		panic(fmt.Errorf("name %q not registered with json.Unmarshaler", name))
	}
	return sc, err
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
					sc, err := fromJSON(scName, data)
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

					err = sc.Run(r, buf)

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

// preprocess preprocesses a txtar file.
func preprocess(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "$\n", "\n") // remove terminal-marking $
	return s
}
