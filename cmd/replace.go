package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/Masterminds/sprig/v3"
)

type ReplaceSubcommand struct {
	Col             string
	Regexp          string
	repl            string
	caseInsensitive bool
	Templ           string
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
	fs.StringVar(&sub.Col, "columns", "", "Columns to replace cells")
	fs.StringVar(&sub.Col, "c", "", "Columns to replace cells (shorthand)")
	fs.StringVar(&sub.Regexp, "regex", "", "Regular expression to match for replacement")
	fs.StringVar(&sub.repl, "repl", "", "Replacement string")
	fs.BoolVar(&sub.caseInsensitive, "case-insensitive", false, "Make regex case insensitive")
	fs.BoolVar(&sub.caseInsensitive, "i", false, "Make regex case insensitive (shorthand)")
	fs.StringVar(&sub.Templ, "template", "", "")
	fs.StringVar(&sub.Templ, "t", "", "")
}

func (sub *ReplaceSubcommand) Run(args []string) {
	inputCsvs := GetInputCsvsOrPanic(args, 1)
	outputCsv := NewOutputCsvFromInputCsv(inputCsvs[0])
	sub.RunReplace(inputCsvs[0], outputCsv)
}

func (sub *ReplaceSubcommand) RunReplace(inputCsv *InputCsv, outputCsvWriter OutputCsvWriter) {
	// Get columns to compare against
	var columns []string
	if sub.Col == "" {
		columns = make([]string, 0)
	} else {
		columns = GetArrayFromCsvString(sub.Col)
	}

	// Get replace function
	if sub.caseInsensitive {
		sub.Regexp = "(?i)" + sub.Regexp
	}
	re, err := regexp.Compile(sub.Regexp)
	if err != nil {
		ExitWithError(err)
	}

	var fReplacer replacerFunc
	switch sub.Templ {
	default:
		fReplacer = templateReplacerFunc(re, sub.Templ)
	case "":
		fReplacer = regexpReplacerFunc(re, sub.repl)
	}

	ReplaceWithFunc(inputCsv, outputCsvWriter, columns, fReplacer)
}

func ReplaceWithFunc(inputCsv *InputCsv, outputCsvWriter OutputCsvWriter, columns []string, fReplacer replacerFunc) {
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
			rowToWrite[columnIndex] = fReplacer(rowToWrite[columnIndex])
		}
		outputCsvWriter.Write(rowToWrite)
	}
}

// A replacerFunc takes a field and returns it with some
// text replaced.
type replacerFunc func(field string) string

// regexpReplacerFunc returns a replacerFunc that replaces
// all matches and submatches in field with repl.
func regexpReplacerFunc(re *regexp.Regexp, repl string) replacerFunc {
	return func(field string) string {
		return re.ReplaceAllString(field, repl)
	}
}

// templateReplacerFunc returns a replacerFunc that replaces
// the field with the rendered templ for each submatch of re.
func templateReplacerFunc(re *regexp.Regexp, templ string) replacerFunc {
	newTempl := convertTemplNames(templ)
	debugln(newTempl)

	t, err := template.New("template").Funcs(sprig.FuncMap()).Parse(newTempl)
	if err != nil {
		ExitWithError(err)
	}

	var (
		// for the field "foo987" and the re `([a-z]+)(\d*(\d))`, and
		// using [namePrefix]:
		//   {Submatch_0: foo987 Submatch_1:foo Submatch_2:987 Submatch_3:7}
		submatchData = make(map[string]string)

		namePrefix = strings.TrimPrefix(templDataPrefix, ".")
		buf        = &bytes.Buffer{}
	)

	return func(field string) string {
		matches := re.FindAllStringSubmatch(field, -1)
		debugln("matches:", matches)
		for _, match := range matches {
			for k := range submatchData {
				submatchData[k] = "<no-value>"
			}
			for i, value := range match {
				name := fmt.Sprintf("%s%d", namePrefix, i)
				submatchData[name] = value
			}
			debugln("submatchData:", submatchData)
			buf.Reset()
			err = t.Execute(buf, submatchData)
			if err != nil {
				ExitWithError(err)
			}
			field = strings.Replace(field, match[0], buf.String(), 1)
		}
		return field
	}
}

// match submatch-specifiers like $0, or its escaped
// equivalent ${0}, submatching the number
var reSubmatchToken = regexp.MustCompile(`\$\{?(\d+)\}?`)

const templDataPrefix = ".Submatch_"

// convertTemplNames replaces submatch specifiers with names
// that can be called while executing a template,
// e.g.:, $0 → .Submatch_0, ${4} → .Submatch_4.
func convertTemplNames(templ string) (newTempl string) {
	return reSubmatchToken.ReplaceAllString(templ, templDataPrefix+"$1")
}
