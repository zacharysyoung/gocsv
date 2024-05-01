package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
)

const usage = `Usage: csv [-v | -h] <command> <args>

Commands:
clean   Prepare input CSV for further processing
conv    Convert non-CSV formats, like Markdown table, to CSV
filter  Filter rows of input CSV based on a columns's values
select  Select (or omit) certain columns of the input CSV
sort    Sort rows of input CSV based on a column's values
tail    Print the end of the input CSV
view    Print the input CSV in nicer-to-look-at formats
`

var streamers = map[string]scMaker{
	"clean":  newClean,
	"conv":   newConvert,
	"filter": newFilter,
	"select": newSelect,
	"sort":   newSort,
	"tail":   newTail,
	"view":   newView,
}

type scMaker func(...string) (subcmd.SubCommander, []string, error)

func main() {
	if len(os.Args) < 2 {
		printHelp()
	}

	switch os.Args[1] {
	case "-v":
		printVersion()
	case "-h":
		printHelp()
	}

	name := os.Args[1]

	newfunc, ok := streamers[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: no command %s\n", name)
		printHelp()
	}

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

func newClean(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] [-trim]"

	var (
		fs       = flag.NewFlagSet("clean", flag.ExitOnError)
		trimFlag = fs.Bool("trim", false, "trim leading spaces from fields")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of clean: %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}

	fs.Parse(args)

	sc := subcmd.NewClean(*trimFlag)
	return sc, fs.Args(), nil
}

func newConvert(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] -fields | -md"

	var (
		fs           = flag.NewFlagSet("conv", flag.ExitOnError)
		fieldsFlag   = fs.Bool("fields", false, "convert fields to CSV")
		markdownFlag = fs.Bool("md", false, "convert Markdown table to CSV")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of convert: %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}

	fs.Parse(args)

	switch {
	case *fieldsFlag, *markdownFlag:
	default:
		return nil, nil, errors.New("conv: must specify either -fields or -md")
	}

	sc := subcmd.NewConvert(*fieldsFlag, *markdownFlag)
	return sc, fs.Args(), nil
}

func newFilter(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h]  -col col_num -eq|-ne|-lt|-lte|-gt|-gte|-re value  [-i] [-exclude] [-no-infer] [file]"

	var (
		fs        = flag.NewFlagSet("filter", flag.ExitOnError)
		colFlag   = fs.Int("col", 1, "the column number with values to compare and filter")
		iFlag     = fs.Bool("i", false, "make any string comparison case-insensitive")
		exFlag    = fs.Bool("exclude", false, "print non-matching rows")
		noInfFlag = fs.Bool("no-infer", false, "treat all values as strings")
	)
	fs.String("ne", "", "filter if not equal to value")
	fs.String("eq", "", "filter if equal to value")
	fs.String("gt", "", "filter if greater than value")
	fs.String("gte", "", "filter if greater than or equal to value")
	fs.String("lt", "", "filter if less than value")
	fs.String("lte", "", "filter if less than or equal to value")
	fs.String("re", "", "filter if matches regular expression")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of filter: %s
  -col int
    	the column number with values to compare and filter (default 1)

  -eq string
    	filter if equal to value
  -ne string
    	filter if not equal to value
  -gt string
    	filter if greater than value
  -gte string
    	filter if greater than or equal to value
  -lt string
    	filter if less than value
  -lte string
    	filter if less than or equal to value

  -re string
    	filter if value matches regular expression

  -i	make any string comparison case-insensitive
  -exclude
    	print non-matching rows
  -no-infer
    	treat all values as strings
`, usage)
		// fs.PrintDefaults()
		os.Exit(2)
	}

	fs.Parse(args)

	var (
		ops []subcmd.Operator
		val string
	)
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "ne", "eq", "gt", "gte", "lt", "lte", "re":
			ops = append(ops, subcmd.Operator(f.Name))
			val = f.Value.String()
		}
	})
	if len(ops) > 1 {
		return nil, nil, fmt.Errorf("filter: cannot combine operators %v; one and only one operator can be set\n%s", ops, usage)
	}
	if *colFlag < 1 {
		return nil, nil, errors.New("filter: -col must be a positive integer")
	}

	sc := subcmd.NewFilter(*colFlag, ops[0], val)
	sc.CaseInsensitive = *iFlag
	sc.Exclude = *exFlag
	sc.NoInference = *noInfFlag

	return sc, fs.Args(), nil
}

func newSelect(args ...string) (subcmd.SubCommander, []string, error) {
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
	return subcmd.NewSelect(cols, *exflag), fs.Args(), nil
}

func newSort(args ...string) (subcmd.SubCommander, []string, error) {
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
	return subcmd.NewSort(cols, *revflag, false), fs.Args(), nil
}

func newTail(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] [-n]"
	var (
		fs = flag.NewFlagSet("tail", flag.ExitOnError)

		n       int
		fromTop bool
	)
	fs.Func(
		"n",
		"the number of rows to print relative to the end of the input (default 10); a leading plus sign ('+') makes the offset relative to the beginning of the input, e.g., '+2' will start printing at the second row",
		func(s string) error {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "+") {
				fromTop = true
				s = strings.TrimPrefix(s, "+")
			}
			var err error
			n, err = strconv.Atoi(s)
			return err
		})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of tail:  %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)

	return subcmd.NewTail(n, fromTop), fs.Args(), nil
}

func newView(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] [-box [-maxh] | -fields | -md] [-maxw]"
	var (
		fs         = flag.NewFlagSet("view", flag.ExitOnError)
		mdflag     = fs.Bool("md", false, "print as (extended) Markdown table")
		fieldsflag = fs.Bool("fields", false, "print as space-delimited cells")
		boxflag    = fs.Bool("box", false, "print complete cells in simple ascii boxes")

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
		fmt.Fprintf(os.Stderr, `Usage of view: %s
  -box
    	print complete cells in simple ascii boxes
  -maxh value
    	cap the height of printed multiline cells; must be preceded
    	by -box; minimum of 1; defaults to "all"

  -md
      	print as (extended) Markdown table
  -fields
    	print as space-delimited cells

  -maxw value
    	cap the width of printed cells; minimum of 3
`, usage)
		// fs.PrintDefaults()
		os.Exit(2)
	}

	fs.Parse(args)

	if maxhflag < 1 && !*boxflag {
		maxhflag = 1
	}

	return &View{box: *boxflag, markdown: *mdflag, fields: *fieldsflag, maxh: maxhflag, maxw: maxwflag}, fs.Args(), nil
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

func printHelp() {
	fmt.Fprint(os.Stderr, usage)
	os.Exit(2)
}

func printVersion() {
	s := "gocsv2"
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, x := range bi.Settings {
			if x.Key == "vcs.revision" {
				s += ":" + x.Value[:7] // short hash
				break
			}
		}
		s += ":" + bi.GoVersion
	}
	fmt.Fprintln(os.Stderr, s)
	os.Exit(2)
}

func errorBadArgs(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
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
