package cmd

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"

	"github.com/zacharysyoung/gocsv/pkg/cmd/view"
)

func TestCmds(t *testing.T) {
	a, err := txtar.ParseFile("testdata/view_cmd.txt")
	if err != nil {
		t.Fatal(err)
	}

	vc := view.View{}

	in := bytes.NewReader(a.Files[0].Data)
	flags := strings.Split(string(a.Files[1].Data), " ")
	want := string(a.Files[2].Data)

	buf := &bytes.Buffer{}
	vc.Run(in, buf, flags...)
	got := buf.String()

	if got != want {
		t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
	}

}
