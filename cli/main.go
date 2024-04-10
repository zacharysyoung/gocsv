package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/zacharysyoung/gocsv/pkg/cmd"
)

type streamer func(...string) (cmd.SubCommander, []string)

var streamers = map[string]streamer{
	"select": newSelect,
	"sort":   newSort,
	"view":   newView,
}

func main() {
	name := os.Args[1]

	if newfunc, ok := streamers[name]; ok {
		sc, tailArgs := newfunc(os.Args[2:]...)

		var (
			r   io.Reader
			err error
		)
		switch len(tailArgs) {
		case 0:
			r = os.Stdin
		case 1:
			if r, err = os.Open(tailArgs[0]); err != nil {
				errorBadArgs(err)
			}
		default:
			errorBadArgs(fmt.Errorf("got %d extra args; %s allows for only one named file", len(tailArgs), name))
		}

		if err = sc.CheckConfig(); err != nil {
			errorBadArgs(err)
		}

		if err = sc.Run(r, os.Stdout); err != nil {
			errorOut("", err)
		}

		return
	}
}

func newSelect(args ...string) (cmd.SubCommander, []string) {
	var (
		fs       = flag.NewFlagSet("select", flag.ExitOnError)
		colsflag = fs.String("cols", "", "a range of columns to select, e.g., 1,3-5,2")
		exflag   = fs.Bool("exclude", false, "invert the cols selection to exclude those columns")
	)
	fs.Parse(args)
	cols, _ := parseCols(*colsflag)
	return cmd.NewSelect(cols, *exflag), fs.Args()
}

func newSort(args ...string) (cmd.SubCommander, []string) {
	var (
		fs       = flag.NewFlagSet("sort", flag.ExitOnError)
		colsflag = fs.String("cols", "", "a range of columns to use as the sort key, e.g., 1,3-5,2")
		revflag  = fs.Bool("reversed", false, "sort descending")
	)
	fs.Parse(args)
	cols, _ := parseCols(*colsflag)
	return cmd.NewSort(cols, *revflag, false), fs.Args()
}

func newView(args ...string) (cmd.SubCommander, []string) {
	var (
		fs      = flag.NewFlagSet("view", flag.ExitOnError)
		mdflag  = fs.Bool("md", false, "print as (extended) Markdown table")
		boxflag = fs.Bool("box", false, "print complete cells in simple ascii boxes")

		maxwflag, maxhflag int = -1, -1

		err error
	)
	fs.Func("maxw", "cap the width of printed cells; minimum of 3", func(s string) error {
		maxwflag, err = strconv.Atoi(s)
		if err != nil {
			return err
		}
		if maxwflag < 3 {
			return errors.New("minimum of 3")
		}
		return nil
	})
	fs.Func("maxh", "cap the height of printed multiline cells; must be preceded by -box; minimum of 1", func(s string) error {
		maxhflag, err = strconv.Atoi(s)
		if err != nil {
			return err
		}
		if !*boxflag {
			return errors.New("must be preceded by -box")
		}
		if maxhflag < 1 {
			return errors.New("minimum of 1")
		}
		return nil
	})
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: view [-box [-maxh] | -md] [-maxw]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)
	return &View{box: *boxflag, md: *mdflag, maxh: maxhflag, maxw: maxwflag}, fs.Args()
}

func parseCols(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}

	r := csv.NewReader(strings.NewReader(s))
	recs, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	cols := make([]int, 0)
	for _, x := range recs[0] {
		if !strings.Contains(x, "-") {
			a, err := strconv.Atoi(x)
			if err != nil {
				return nil, err
			}
			cols = append(cols, a)
			continue
		}

		if strings.Count(x, "-") != 1 {
			return nil, fmt.Errorf("too many dashes in %s", x)
		}
		xs := strings.Split(x, "-")
		a, err := strconv.Atoi(xs[0])
		if err != nil {
			return nil, err
		}
		b, err := strconv.Atoi(xs[1])
		if err != nil {
			return nil, err
		}
		if a < b {
			for i := a; i <= b; i++ {
				cols = append(cols, i)
			}
		} else {
			for i := a; i >= b; i-- {
				cols = append(cols, i)
			}
		}
	}
	return cols, nil
}

func errorBadArgs(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}

func errorOut(msg string, err error) {
	if msg != "" && err != nil {
		msg = fmt.Sprintf("error: %s: %v", msg, err)
	} else if msg != "" {
		msg = fmt.Sprintf("error: %s", msg)
	} else {
		msg = fmt.Sprintf("error: %v", err)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
