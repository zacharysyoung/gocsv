package stack

import (
	"bytes"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestStack(t *testing.T) {
	a, err := txtar.ParseFile("testdata/stack.txt")
	if err != nil {
		t.Fatal(err)
	}

	reIn := regexp.MustCompile(`in\d`)

	i := 0
	for i < len(a.Files) {
		readers := make([]io.Reader, 0)
		for {
			if reIn.MatchString(a.Files[i].Name) {
				readers = append(readers, bytes.NewReader(a.Files[i].Data))
				i++
			} else {
				break
			}
		}

		name := a.Files[i].Name
		want := a.Files[i].Data
		i++

		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			stack := NewStack()
			err = stack.Run(readers, buf)

			if err != nil {
				switch name {
				default:
					t.Fatal(err)
				case "err":
					gotErr := err.Error()
					wantErr := strings.TrimSpace(string(want))
					if gotErr != wantErr {
						t.Errorf("got error %q; want %q", gotErr, wantErr)
					}
					return
				}
			}

			if got := buf.Bytes(); !reflect.DeepEqual(got, want) {
				t.Errorf("\ngot:\n%s\nwant:\n%s", postprocess(got), postprocess(want))
			}
		})

	}
}

func postprocess(s []byte) []byte {
	out := make([]byte, 0, len(s))
	for _, x := range s {
		if x == '\n' {
			out = append(out, '$')
		}
		out = append(out, x)
	}
	return out
}
