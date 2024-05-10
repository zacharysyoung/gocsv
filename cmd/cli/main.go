package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/zacharysyoung/gocsv/pkg/subcmd"
	_select "github.com/zacharysyoung/gocsv/pkg/subcmd/select"
)

const usage = `Usage: csv [-v | -h] <command> <args>

Commands:
clean   Prepare input CSV for further processing
conv    Convert non-CSV formats, like Markdown table, to CSV
filter  Filter rows of input CSV based on values in a column
head    Print beggining rows of input CSV
rename  Rename CSV's columns
select  Select (or omit) certain columns of input CSV
sort    Sort rows of input CSV based on a column's values
tail    Print ending rows of input CSV
view    Print input CSV in nicer-to-look-at formats
`

var streamers = map[string]scMaker{
	"clean":  newClean,
	"conv":   newConvert,
	"filter": newFilter,
	"head":   newHead,
	"rename": newRename,
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
	const usage = "[-h] -fields | -md [file]"

	var (
		fs           = flag.NewFlagSet("conv", flag.ExitOnError)
		fieldsFlag   = fs.Bool("fields", false, "convert whitespace-delimited fields to CSV")
		markdownFlag = fs.Bool("md", false, "converts first (Github Flavored) Markdown table to CSV; all other Markdown is discarded")
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

func newHead(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] [-n] [file]"
	var (
		fs = flag.NewFlagSet("head", flag.ExitOnError)

		n          = 10
		fromBottom = false
	)
	fs.Func(
		"n",
		"the number of rows to print relative to the beginning of the input (default 10); a leading minus sign ('-') makes the terminal row relative to the end of the input, e.g., '-2' will print up to the second-from-last row",
		func(s string) error {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "-") {
				fromBottom = true
				s = strings.TrimPrefix(s, "-")
			}
			var err error
			n, err = strconv.Atoi(s)
			if err != nil || n == 0 {
				return errors.New("n must be a non-zero integer")
			}
			return nil
		})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of head: %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)
	return subcmd.NewHead(n, fromBottom), fs.Args(), nil
}

func newRename(args ...string) (subcmd.SubCommander, []string, error) {
	const usage = "[-h] [-cols] [-names | -regexp [-repl]] [file]"
	var (
		fs         = flag.NewFlagSet("select", flag.ExitOnError)
		colsflag   = fs.String("cols", "", "a range of columns to select, e.g., 1,3-5,2")
		namesflag  = fs.String("names", "", "list of new names, matches the count and order of represented columns in -cols")
		regexpflag = fs.String("regexp", "", "regexp to match names in -cols")
		replflag   = fs.String("repl", "", "replacement string; can only be used with -regexp")
	)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of filter: %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)
	groups, err := parseCols(*colsflag)
	if err != nil {
		return nil, nil, err
	}
	names := strings.Split(*namesflag, ",")
	if len(names) == 1 && names[0] == "" {
		names = names[:0]
	}
	return subcmd.NewRename(groups, names, *regexpflag, *replflag), fs.Args(), nil
}

func newSelect(args ...string) (subcmd.SubCommander, []string, error) {
	var (
		fs       = flag.NewFlagSet("select", flag.ExitOnError)
		colsflag = fs.String("cols", "", "a range of columns to select, e.g., 1,3-5,2")
		exflag   = fs.Bool("exclude", false, "invert the cols selection to exclude those columns")
	)
	fs.Parse(args)
	groups, err := parseCols(*colsflag)
	if err != nil {
		return nil, nil, err
	}
	return _select.NewSelect(groups, *exflag), fs.Args(), nil
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
	const usage = "[-h] [-n] [file]"
	var (
		fs = flag.NewFlagSet("tail", flag.ExitOnError)

		n          = 10
		fromBottom = false
	)
	fs.Func(
		"n",
		"the number of rows to print relative to the end of the input (default 10); a leading plus sign ('+') makes the offset relative to the beginning of the input, e.g., '+2' will start printing at the second row",
		func(s string) error {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "+") {
				fromBottom = true
				s = strings.TrimPrefix(s, "+")
			}
			var err error
			n, err = strconv.Atoi(s)
			if err != nil || n < 1 {
				return errors.New("n must be a positive integer")
			}
			return nil
		})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of tail:  %s\n", usage)
		fs.PrintDefaults()
		os.Exit(2)
	}
	fs.Parse(args)

	return subcmd.NewTail(n, fromBottom), fs.Args(), nil
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

// parseCols parses a comma-delimited list of 1-based columns and
// column ranges into column groups, e.g.,
//
//	1,2,3         → [[1],[2],[3]]
//	4-6,8,-3,9-,7 → [[4,6],[8],[-1,3],[9,-1],[7]].
//
// Once the calling subcmd has the header it will send the groups
// to [subcmd.FinalizeCols] to finalize the indexes.
func parseCols(s string) ([]subcmd.ColGroup, error) {
	if s == "" {
		return nil, nil
	}
	if strings.Contains(s, " ") {
		return nil, fmt.Errorf("cols cannot have a space: %s", s)
	}

	ss := strings.Split(s, ",")

	groups := make([]subcmd.ColGroup, 0)
	for _, x := range ss {
		switch strings.Contains(x, "-") {
		case false:
			a, err := strconv.Atoi(x)
			if err != nil {
				return nil, err
			}
			groups = append(groups, subcmd.ColGroup{a})
			continue
		case true:
			group, err := splitRange(x)
			if err != nil {
				return nil, err
			}
			groups = append(groups, group)
		}
	}
	return groups, nil
}

// splitRange parses a range string into a ColGroup.
func splitRange(s string) (group subcmd.ColGroup, err error) {
	const dash = "-"

	a, b := -1, -1
	switch strings.Count(s, dash) {
	case 0:
		if a, err = strconv.Atoi(s); err != nil {
			return
		}
		group = subcmd.ColGroup{a}
	case 1:
		ss := strings.Split(s, dash)
		if ss[0] != "" {
			if a, err = strconv.Atoi(ss[0]); err != nil {
				return
			}
		}
		if ss[1] != "" {
			if b, err = strconv.Atoi(ss[1]); err != nil {
				return
			}
		}
		group = subcmd.ColGroup{a, b}
	default:
		err = errors.New("wrong number of dashes")
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
