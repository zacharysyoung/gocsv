package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"regexp"
	"text/template"

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
	fs.StringVar(&sub.Templ, "templ", "", "")
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
	var replaceFunc func(string) string
	if sub.caseInsensitive {
		sub.Regexp = "(?i)" + sub.Regexp
	}
	re, err := regexp.Compile(sub.Regexp)
	if err != nil {
		ExitWithError(err)
	}
	replaceFunc = func(field string) string {
		return re.ReplaceAllString(field, sub.repl)
	}

	if sub.Templ != "" {
		// match subgroup tokens like $0 or ${22}
		reGroupToken := regexp.MustCompile(`\$\{?(\d+)\}?`)
		templ := reGroupToken.ReplaceAllString(sub.Templ, `.MatchKey_$1`)
		// fmt.Println(templ)

		tmpl, err := template.New("template").Funcs(sprig.FuncMap()).Parse(templ)
		if err != nil {
			ExitWithError(err)
		}

		templateData := make(map[string]string)

		replaceFunc = func(field string) string {
			// Clear previous values
			for k := range templateData {
				delete(templateData, k)
			}

			submatches := re.FindAllStringSubmatch(field, -1)
			fmt.Println("re:", re, "field:", field, "submatches:", submatches)
			if len(submatches) > 0 && len(submatches[0]) > 0 {
				for i := 0; i < len(submatches[0]); i++ {
					templateData[fmt.Sprintf("MatchKey_%d", i)] = submatches[0][i]
				}
			}
			fmt.Println(templateData)

			var rendered bytes.Buffer
			err = tmpl.Execute(&rendered, templateData)
			return re.ReplaceAllString(field, rendered.String())
		}

	}

	ReplaceWithFunc(inputCsv, outputCsvWriter, columns, replaceFunc)
}

func ReplaceWithFunc(inputCsv *InputCsv, outputCsvWriter OutputCsvWriter, columns []string, replaceFunc func(string) string) {
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
