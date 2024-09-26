package headers

import (
	"encoding/json"
	"path/filepath"
	"testing"

	xx "github.com/zacharysyoung/gocsv/subcmd"
)

func fromJSON(data []byte) (xx.Streamer, error) {
	header := &Headers{}
	err := json.Unmarshal(data, header)
	return header, err
}

func TestHeadTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/header.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := xx.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
