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

type scMaker func(...string) (cmd.SubCommander, []string, error)

var streamers = map[string]scMaker{
	"filter": newFilter,
	"select": newSelect,
	"sort":   newSort,
	"view":   newView,
}

func main() {
	name := os.Args[1]

	if newfunc, ok := streamers[name]; ok {
		sc, tailArgs, err := newfunc(os.Args[2:]...)
		if err != nil {
			errorBadArgs(err)
		}

		var r io.Reader
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

		if err := sc.CheckConfig(); err != nil {
			errorBadArgs(err)
		}

		if err := sc.Run(r, os.Stdout); err != nil {
			errorOut("", err)
		}

		return
	}
}

func newFilter(args ...string) (cmd.SubCommander, []string, error) {
	var (
		fs = flag.NewFlagSet("filter", flag.ExitOnError)

		colFlag = fs.Int("col", 1, "the column to compare")

		neFlag  = fs.String("ne", "", "the value to not equal")
		eqFlag  = fs.String("eq", "", "the value to equal")
		gtFlag  = fs.String("gt", "", "the value to be greater than")
		gteFlag = fs.String("gte", "", "the value to be greater than or equal to")
		ltFlag  = fs.String("lt", "", "the value to be less than")
		lteFlag = fs.String("lte", "", "the value to be less than or equal to")
		reFlag  = fs.String("re", "", "the regular expression value to match")

		iFlag  = fs.Bool("i", false, "make any string comparison case-insensitive")
		exFlag = fs.Bool("exclude", false, "print non-matching rows")
	)
	fs.Parse(args)

	var (
		op  cmd.Operator
		val string
	)
	switch {
	case *neFlag != "":
		op = cmd.Ne
		val = *neFlag
	case *eqFlag != "":
		op = cmd.Eq
		val = *eqFlag
	case *gtFlag != "":
		op = cmd.Gt
		val = *gtFlag
	case *gteFlag != "":
		op = cmd.Gte
		val = *gteFlag
	case *ltFlag != "":
		op = cmd.Lt
		val = *ltFlag
	case *lteFlag != "":
		op = cmd.Lt
		val = *lteFlag

	case *reFlag != "":
		op = cmd.Re
		val = *reFlag
	}

	return cmd.NewFilter(*colFlag, op, val, *iFlag, *exFlag), fs.Args(), nil
}

func newSelect(args ...string) (cmd.SubCommander, []string, error) {
	var (
		fs       = flag.NewFlagSet("select", flag.ExitOnError)
		colsflag = fs.String("cols", "", "a range of columns to select, e.g., 1,3-5,2")
		exflag   = fs.Bool("exclude", false, "invert the cols selection to exclude those columns")
	)
	fs.Parse(args)
	cols, err := parseCols(*colsflag)
	if err != nil {
		return nil, nil, err
	}
	return cmd.NewSelect(cols, *exflag), fs.Args(), nil
}

func newSort(args ...string) (cmd.SubCommander, []string, error) {
	var (
		fs       = flag.NewFlagSet("sort", flag.ExitOnError)
		colsflag = fs.String("cols", "", "a range of columns to use as the sort key, e.g., 1,3-5,2")
		revflag  = fs.Bool("reversed", false, "sort descending")
	)
	fs.Parse(args)
	cols, err := parseCols(*colsflag)
	if err != nil {
		return nil, nil, err
	}
	return cmd.NewSort(cols, *revflag, false), fs.Args(), nil
}

func newView(args ...string) (cmd.SubCommander, []string, error) {
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
			return errors.New("-maxh must be preceded by -box")
		}
		if maxhflag < 1 {
			return errors.New("minimum of 1")
		}
		return nil
	})
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: view [-box [-maxh] | -md] [-maxw]")
		fs.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)
	return &View{box: *boxflag, md: *mdflag, maxh: maxhflag, maxw: maxwflag}, fs.Args(), nil
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
		switch strings.Count(x, "-") {
		case 0:
			a, err := strconv.Atoi(x)
			if err != nil {
				return nil, err
			}
			cols = append(cols, a)
			continue
		case 1:
			xcols, err := splitRange(x)
			if err != nil {
				return nil, err
			}
			cols = append(cols, xcols...)
		default:
			return nil, fmt.Errorf("too many dashes in %s", x)

		}
	}
	return cols, nil
}

// splitRange splits a range of column indexes into a
// complete slice of the represented indexes, e.g.,
// "1-4" to [1 2 3 4], or "9-7" to [9 8 7].
func splitRange(x string) (cols []int, err error) {
	s := strings.Split(x, "-")

	var a, b int
	if a, err = strconv.Atoi(s[0]); err != nil {
		return
	}
	if b, err = strconv.Atoi(s[1]); err != nil {
		return
	}

	switch a < b {
	case true:
		for i := a; i <= b; i++ {
			cols = append(cols, i)
		}
	case false:
		for i := a; i >= b; i-- {
			cols = append(cols, i)
		}
	}

	return
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
