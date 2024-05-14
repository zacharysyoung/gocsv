package head

import (
	"encoding/json"
	"path/filepath"
	"testing"

	xx "github.com/zacharysyoung/gocsv/subcmd"
)

func fromJSON(data []byte) (xx.Runner, error) {
	head := &Head{}
	err := json.Unmarshal(data, head)
	return head, err
}

func TestHeadTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/head.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := xx.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
