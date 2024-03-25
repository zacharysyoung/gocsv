package main

import (
	"fmt"
	"os"

	"github.com/zacharysyoung/gocsv/pkg/cmds"
	"github.com/zacharysyoung/gocsv/pkg/cmds/view"
)

var subCommands = map[string]cmds.Command{
	"view": view.View{},
}

func main() {
	r := os.Stdin
	w := os.Stdout
	for name, sc := range subCommands {
		if os.Args[1] == name {
			args := []string{}
			if len(os.Args) > 2 {
				args = os.Args[2:]
			}
			if err := sc.Run(r, w, args...); err != nil {
				errorOut("", err)
			}
		}
	}

}

func errorOut(msg string, err error) {
	if msg != "" && err != nil {
		msg = fmt.Sprintf("error: %s: %v", msg, err)
	} else if msg != "" {
		msg = fmt.Sprintf("error: %s", msg)
	} else {
		msg = fmt.Sprintf("error: %v", err)
	}
	fmt.Fprintf(os.Stderr, msg)
	os.Exit(2)
}
