package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestCmds(t *testing.T) {
	const suffix = "_cmd.txt"
	files, err := filepath.Glob("testdata/*" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !strings.Contains(file, "view") {
			continue
		}

		cmdname := strings.TrimSuffix(filepath.Base(file), suffix)
		cmd, ok := Commands[cmdname]
		if !ok {
			t.Fatalf("could get not Command %s", cmdname)
		}
		t.Run(cmdname, func(t *testing.T) {
			a, err := txtar.ParseFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var (
				buf []byte        // cache input bytes for multiple subtests
				r   *bytes.Reader // per-subtest reader of buf
				i   = 0
			)
			for i < len(a.Files) {
				if a.Files[i].Name == "input" {
					buf = a.Files[i].Data
					i++
				}
				r = bytes.NewReader(buf)

				t.Run(a.Files[i].Name, func(t *testing.T) {
					flagLine := strings.TrimSpace(string(a.Files[i].Data))
					args := strings.Split(flagLine, " ")
					i++
					want := string(a.Files[i].Data)
					i++

					w := &bytes.Buffer{}
					cmd.Run(r, w, args...)
					got := w.String()

					if got != want {
						t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
					}
				})
			}
		})
	}
}
