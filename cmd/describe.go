package cmd

import (
	"flag"
	"fmt"
	"strconv"
)

type DescribeSubcommand struct {
	asCsv bool
}

func (sub *DescribeSubcommand) Name() string {
	return "describe"
}
func (sub *DescribeSubcommand) Aliases() []string {
	return []string{}
}
func (sub *DescribeSubcommand) Description() string {
	return "Get basic information about a CSV."
}
func (sub *DescribeSubcommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&sub.asCsv, "csv", false, "Output results as CSV")
}

func (sub *DescribeSubcommand) Run(args []string) {
	inputCsvs := GetInputCsvsOrPanic(args, 1)
	DescribeCsv(inputCsvs[0], sub.asCsv)
}

func DescribeCsv(inputCsv *InputCsv, asCsv bool) {
	imc := NewInMemoryCsvFromInputCsv(inputCsv)

	numColumns := imc.NumColumns()

	var descriptions [][]string
	for i := 0; i < numColumns; i++ {
		descriptions = append(descriptions, []string{
			strconv.Itoa(i + 1),
			imc.header[i],
			ColumnTypeToString(imc.InferType(i)),
		})
	}

	var outputCsv *OutputCsv
	if asCsv {
		outputCsv = NewOutputCsvFromInputCsv(inputCsv)
		outputCsv.Write([]string{"Column", "Name", "Type"})
	} else {
		fmt.Println("Columns:")
	}

	for _, x := range descriptions {
		if asCsv {
			outputCsv.Write(x)
		} else {
			fmt.Printf("  %s: %s\n", x[0], x[1])
			fmt.Printf("    Type: %s\n", x[2])
		}
	}
}
