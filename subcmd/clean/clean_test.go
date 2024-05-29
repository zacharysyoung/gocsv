package clean

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func fromJSON(data []byte) (subcmd.Streamer, error) {
	clean := &Clean{}
	err := json.Unmarshal(data, clean)
	return clean, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/clean.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
