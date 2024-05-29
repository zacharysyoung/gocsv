package tail

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/zacharysyoung/gocsv/subcmd"
)

func fromJSON(data []byte) (subcmd.Streamer, error) {
	tail := &Tail{}
	err := json.Unmarshal(data, tail)
	return tail, err
}

func TestTestdata(t *testing.T) {
	path, err := filepath.Abs("./testdata/tail.txt")
	if err != nil {
		t.Fatal(err)
	}
	tdr := subcmd.NewTestdataRunner(path, fromJSON, t)
	tdr.Run()
}
