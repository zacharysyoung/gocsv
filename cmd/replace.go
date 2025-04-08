package cmd

import (
	"flag"
	"io"
	"regexp"
	"strconv"
)

type ReplaceSubcommand struct {
	columnsString   string
	regex           string
	repl            string
	caseInsensitive bool
}

func (sub *ReplaceSubcommand) Name() string {
	return "replace"
}
func (sub *ReplaceSubcommand) Aliases() []string {
	return []string{}
}
func (sub *ReplaceSubcommand) Description() string {
	return "Replace values in cells by regular expression."
}
func (sub *ReplaceSubcommand) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&sub.columnsString, "columns", "", "Columns to replace cells")
	fs.StringVar(&sub.columnsString, "c", "", "Columns to replace cells (shorthand)")
	fs.StringVar(&sub.regex, "regex", "", "Regular expression to match for replacement")
	fs.StringVar(&sub.repl, "repl", "", "Replacement string")
	fs.BoolVar(&sub.caseInsensitive, "case-insensitive", false, "Make regex case insensitive")
	fs.BoolVar(&sub.caseInsensitive, "i", false, "Make regex case insensitive (shorthand)")
}

func (sub *ReplaceSubcommand) Run(args []string) {
	inputCsvs := GetInputCsvsOrPanic(args, 1)
	outputCsv := NewOutputCsvFromInputCsv(inputCsvs[0])
	sub.RunReplace(inputCsvs[0], outputCsv)
}

func (sub *ReplaceSubcommand) RunReplace(inputCsv *InputCsv, outputCsvWriter OutputCsvWriter) {
	// Get columns to compare against
	var columns []string
	if sub.columnsString == "" {
		columns = make([]string, 0)
	} else {
		columns = GetArrayFromCsvString(sub.columnsString)
	}

	// Get replace function
	if sub.caseInsensitive {
		sub.regex = "(?i)" + sub.regex
	}
	replaceFunc, err := regexReplaceFunc(sub.regex, sub.repl)
	if err != nil {
		ExitWithError(err)
	}

	ReplaceWithFunc(inputCsv, outputCsvWriter, columns, replaceFunc)
}

// replacerFunc is a func that returns elem modified by some
// internal replacement process.
type replacerFunc func(elem string) string

func regexReplaceFunc(regex, repl string) (replacerFunc, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	repl, err = strconv.Unquote(`"` + repl + `"`)
	if err != nil {
		return nil, err
	}

	f := func(elem string) string {
		return re.ReplaceAllString(elem, repl)
	}

	return f, nil
}

func ReplaceWithFunc(inputCsv *InputCsv, outputCsvWriter OutputCsvWriter, columns []string, replaceFunc replacerFunc) {
	// Read header to get column index and write.
	header, err := inputCsv.Read()
	if err != nil {
		ExitWithError(err)
	}

	columnIndices := GetIndicesForColumnsOrPanic(header, columns)

	outputCsvWriter.Write(header)

	// Write replaced rows
	rowToWrite := make([]string, len(header))
	for {
		row, err := inputCsv.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				ExitWithError(err)
			}
		}
		copy(rowToWrite, row)
		for _, columnIndex := range columnIndices {
			rowToWrite[columnIndex] = replaceFunc(rowToWrite[columnIndex])
		}
		outputCsvWriter.Write(rowToWrite)
	}
}
