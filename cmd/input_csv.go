package cmd

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/aotimme/gocsv/csv"
)

type InputCsvReader interface {
	Read() ([]string, error)
	ReadAll() ([][]string, error)
}

type InputCsv struct {
	file      *os.File
	filename  string
	reader    *csv.Reader
	bufReader *bufio.Reader
	hasBom    bool
}

// NewInputCsv creates an InputCsv from the named file,
// or stdin if filename is "-".
// The presence of a UTF-8 BOM is recorded, and can be
// passed to the OutputCsv writer to preserve the BOM in
// the output stream.
func NewInputCsv(filename string) (ic *InputCsv, err error) {
	var f *os.File
	if filename == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(filename)
		if err != nil {
			return
		}
	}

	ic, err = fromReader(f)
	if err != nil {
		return
	}

	ic.file = f
	ic.filename = filename
	return
}

func fromReader(r io.Reader) (ic *InputCsv, err error) {
	ic = new(InputCsv)

	ic.bufReader = bufio.NewReader(r)
	ic.reader = csv.NewReader(ic.bufReader)
	delimiter := os.Getenv("GOCSV_DELIMITER")
	if delimiter != "" {
		ic.reader.Comma = GetDelimiterFromStringOrPanic(delimiter)
	}
	err = ic.handleBom()
	return
}

func (ic *InputCsv) handleBom() error {
	bomRune, _, err := ic.bufReader.ReadRune()
	if err != nil && err != io.EOF {
		return err
	}
	if err != io.EOF && bomRune == BOM_RUNE {
		ic.hasBom = true
	} else {
		ic.bufReader.UnreadRune()
	}
	return nil
}

func (ic *InputCsv) Close() error {
	return ic.file.Close()
}

func (ic *InputCsv) SetFieldsPerRecord(fieldsPerRecord int) {
	ic.reader.FieldsPerRecord = fieldsPerRecord
}

func (ic *InputCsv) SetLazyQuotes(lazyQuotes bool) {
	ic.reader.LazyQuotes = lazyQuotes
}

func (ic *InputCsv) SetDelimiter(delimiter rune) {
	ic.reader.Comma = delimiter
}

func (ic *InputCsv) Reader() *csv.Reader {
	return ic.reader
}

func (ic *InputCsv) Read() (row []string, err error) {
	return ic.reader.Read()
}

func (ic *InputCsv) ReadAll() (rows [][]string, err error) {
	return ic.reader.ReadAll()
}

func (ic *InputCsv) Name() string {
	if ic.filename == "-" {
		return "stdin"
	} else {
		return GetBaseFilenameWithoutExtension(ic.filename)
	}
}

func (ic *InputCsv) Filename() string {
	return ic.filename
}

func GetInputCsvsOrPanic(filenames []string, numInputCsvs int) (csvs []*InputCsv) {
	csvs, err := GetInputCsvs(filenames, numInputCsvs)
	if err != nil {
		ExitWithError(err)
	}
	return
}

func GetInputCsvs(filenames []string, numInputCsvs int) (csvs []*InputCsv, err error) {
	hasDash := false
	for _, filename := range filenames {
		if filename == "-" {
			hasDash = true
			break
		}
	}
	if numInputCsvs == -1 {
		if len(filenames) == 0 {
			csvs = make([]*InputCsv, 1)
			csvs[0], err = NewInputCsv("-")
			return
		} else {
			csvs = make([]*InputCsv, len(filenames))
			for i, filename := range filenames {
				csvs[i], err = NewInputCsv(filename)
				if err != nil {
					return
				}
			}
			return
		}
	} else {
		csvs = make([]*InputCsv, numInputCsvs)
		if len(filenames) > numInputCsvs {
			err = errors.New("too many files for command")
			return
		}
		if len(filenames) == numInputCsvs {
			for i, filename := range filenames {
				csvs[i], err = NewInputCsv(filename)
				if err != nil {
					return
				}
			}
			return
		}
		if len(filenames) == numInputCsvs-1 {
			if hasDash {
				err = errors.New("too few inputs specified")
				return
			}
			csvs[0], err = NewInputCsv("-")
			if err != nil {
				return
			}
			for i, filename := range filenames {
				csvs[i+1], err = NewInputCsv(filename)
				if err != nil {
					return
				}
			}
			return
		}
		err = errors.New("too few inputs specified")
		return
	}
}
