package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestRunSql(t *testing.T) {
	testCases := []struct {
		queryString string
		rows        [][]string
	}{
		{"SELECT * FROM [simple-sort] WHERE [Number] > 0", [][]string{
			{"Number", "String"},
			{"1", "One"},
			{"2", "Two"},
			{"2", "Another Two"},
		}},
		{"SELECT SUM([Number]) AS Total FROM [simple-sort]", [][]string{
			{"Total"},
			{"4"},
		}},
		{"SELECT [Number], COUNT(*) AS Count FROM [simple-sort] GROUP BY [Number] ORDER BY [Number] ASC", [][]string{
			{"Number", "Count"},
			{"-1", "1"},
			{"1", "1"},
			{"2", "2"},
		}},
	}
	for i, tt := range testCases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			ic, err := NewInputCsv("../test-files/simple-sort.csv")
			if err != nil {
				t.Error("Unexpected error", err)
			}
			toc := new(testOutputCsv)
			sub := new(SqlSubcommand)
			sub.queryString = tt.queryString
			sub.RunSql([]*InputCsv{ic}, toc)
			err = assertRowsEqual(tt.rows, toc.rows)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEscapeSqlName(t *testing.T) {
	testCases := []struct {
		inputName  string
		outputName string
	}{
		{"basic", "\"basic\""},
		{"single space", "\"single space\""},
		{"single'quote", "\"single'quote\""},
		{"square[]brackets", "\"square[]brackets\""},
		{"\"alreadyquoted\"", "\"\"\"alreadyquoted\"\"\""},
		{"middle\"quote", "\"middle\"\"quote\""},
	}
	for i, tt := range testCases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			output := escapeSqlName(tt.inputName)
			if output != tt.outputName {
				t.Errorf("Expected %s but got %s", tt.outputName, output)
			}
		})
	}
}

func TestSql_RunSql(t *testing.T) {
	archive := must(txtar.ParseFile("./testdata/sql.txt"))

	// create tempdir to read input CSVs and SQL scripts
	t.Chdir(t.TempDir())

	// save fs files (normalizing CSVs); separate
	// non-fs files (should be test or want file)
	const fsPrefix = "fs: "
	otherFiles := []txtar.File{}
	for _, f := range archive.Files {
		if !strings.HasPrefix(f.Name, fsPrefix) {
			otherFiles = append(otherFiles, f)
			continue
		}
		if strings.HasSuffix(f.Name, ".csv") {
			f.Data = normalize(f.Data)
		}
		f.Name = strings.TrimPrefix(f.Name, fsPrefix)

		_must(os.WriteFile(f.Name, f.Data, 0444))
	}

	// iterate test/want pairs
	for i := 0; i < len(otherFiles); i += 2 {
		a, b := otherFiles[i], otherFiles[i+1]
		if b.Name != "want" {
			t.Fatalf("after test file %s got want file %q; want \"want\"",
				a.Name, b.Name)
		}
		ftest, fwant := a, b

		t.Run(ftest.Name, func(t *testing.T) {
			subcmd := SqlSubcommand{}
			fs := flag.NewFlagSet("", 0)
			subcmd.SetFlags(fs)
			_must(fs.Parse(getArgs(ftest.Data)))

			inputCsvs := GetInputCsvsOrPanic(fs.Args(), -1)
			outputBuf := bytes.Buffer{}
			outputCsv := &OutputCsv{csvWriter: csv.NewWriter(&outputBuf)}

			subcmd.RunSql(inputCsvs, outputCsv)

			got := prettify(outputBuf.Bytes())
			want := prettify(fwant.Data)
			if !slices.Equal(got, want) {
				t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
			}
		})
	}
}

// normalize parses csvData with [csv.Reader]'s
// TrimLeadingSpace option.
func normalize(csvData []byte) []byte {
	reader := csv.NewReader(bytes.NewReader(csvData))
	reader.TrimLeadingSpace = true
	records := must(reader.ReadAll())

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	_must(writer.WriteAll(records))
	writer.Flush()
	_must(writer.Error())

	return buf.Bytes()
}

// prettify justifies columns (right-justifies numbers)
// in a way which [normalize] can still parse.
//
// Only fit for simple, toy CSV data; does not properly
// encode chars that require quoting.
func prettify(csvData []byte) []byte {
	csvData = normalize(csvData)

	reader := csv.NewReader(bytes.NewReader(csvData))
	records := must(reader.ReadAll())

	// get max width (field length) of each column
	nFields := len(records[0])
	widths := make([]int, nFields)
	for _, record := range records {
		for i, field := range record {
			widths[i] = max(len(field), widths[i])
		}
	}

	// write CSV-like output
	const sep = ", "
	var (
		buf   = &bytes.Buffer{}
		write = func(s string) { must(buf.WriteString(s)) }

		lastField = nFields - 1
	)
	for _, record := range records {
		for i, field := range record {
			pad := strings.Repeat(" ", widths[i]-len(field))
			switch isNum(field) {
			case true:
				// pad num [, ]
				write(pad + field)
				if i != lastField {
					write(sep)
				}
			default:
				// text [, pad]
				write(field)
				if i != lastField {
					write(sep + pad)
				}
			}
		}
		write("\n")
	}

	return bytes.TrimSpace(buf.Bytes())
}

// getArgs tries to interpret data as a JSON array of strings.
func getArgs(data []byte) []string {
	var args []string
	_must(json.Unmarshal(data, &args))
	return args
}

// isNum checks if s represents a number.
func isNum(s string) bool { _, err := strconv.ParseFloat(s, 64); return err == nil }

// must returns obj if err==nil.
func must[T any](obj T, err error) T { _must(err); return obj }

// _must panics if err!=nil.
func _must(err error) {
	if err != nil {
		panic(err)
	}
}
