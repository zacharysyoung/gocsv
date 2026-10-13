package cmd

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"maps"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"
)

type InMemoryCsv struct {
	header []string
	rows   [][]string

	// index of column
	isIndexed bool
	index     map[string][]int
}

func NewInMemoryCsvFromInputCsv(inputCsv *InputCsv) *InMemoryCsv {
	imc := new(InMemoryCsv)
	rows, err := inputCsv.ReadAll()
	if err != nil {
		ExitWithError(err)
	}
	imc.header = rows[0]
	imc.rows = rows[1:]
	imc.isIndexed = false
	return imc
}

func (imc *InMemoryCsv) Index(columnIndex int) {
	imc.index = make(map[string][]int)
	for i, row := range imc.rows {
		rowval := row[columnIndex]
		group, ok := imc.index[rowval]
		if ok {
			group = append(group, i)
		} else {
			group = make([]int, 1)
			group[0] = i
		}
		imc.index[rowval] = group
	}
}

func (imc *InMemoryCsv) NumRows() int {
	return len(imc.rows)
}

func (imc *InMemoryCsv) NumColumns() int {
	return len(imc.header)
}

func (imc *InMemoryCsv) GetRowIndicesMatchingIndexedColumn(value string) []int {
	indices, ok := imc.index[value]
	if ok {
		return indices
	} else {
		return make([]int, 0)
	}
}

func (imc *InMemoryCsv) GetRowsMatchingIndexedColumn(value string) [][]string {
	indices := imc.GetRowIndicesMatchingIndexedColumn(value)
	rows := make([][]string, 0)
	for _, idx := range indices {
		rows = append(rows, imc.rows[idx])
	}
	return rows
}

type ColumnValueIterator struct {
	imc         *InMemoryCsv
	columnIndex int
	rowIndex    int
}

func NewColumnValueIterator(imc *InMemoryCsv, columnIndex int) *ColumnValueIterator {
	return &ColumnValueIterator{
		imc:         imc,
		columnIndex: columnIndex,
		rowIndex:    0,
	}
}
func (cvi *ColumnValueIterator) Next() (string, bool) {
	if cvi.rowIndex >= len(cvi.imc.rows) {
		return "", false
	}
	retval := cvi.imc.rows[cvi.rowIndex][cvi.columnIndex]
	cvi.rowIndex++
	return retval, true
}

func (imc *InMemoryCsv) InferType(columnIndex int) ColumnType {
	cvi := NewColumnValueIterator(imc, columnIndex)
	return InferTypeFromStringIterator(cvi)
}

func (imc *InMemoryCsv) SortRows(columnIndices []int, columnTypes []ColumnType, stable bool, reverse bool) {
	isLessFunc := func(row1Ptr, row2Ptr *[]string) bool {
		row1 := *row1Ptr
		row2 := *row2Ptr
		for i, columnIndex := range columnIndices {
			isElem1Null := IsNullType(row1[columnIndex])
			isElem2Null := IsNullType(row2[columnIndex])
			if isElem1Null && isElem2Null {
				continue
			}
			if isElem1Null && !isElem2Null {
				return true
			}
			if !isElem1Null && isElem2Null {
				return false
			}
			columnType := columnTypes[i]
			if columnType == FLOAT_TYPE {
				row1Val := ParseFloat64OrPanic(row1[columnIndex])
				row2Val := ParseFloat64OrPanic(row2[columnIndex])
				if row1Val < row2Val {
					return true
				} else if row1Val > row2Val {
					return false
				}
			} else if columnType == INT_TYPE {
				row1Val := ParseInt64OrPanic(row1[columnIndex])
				row2Val := ParseInt64OrPanic(row2[columnIndex])
				if row1Val < row2Val {
					return true
				} else if row1Val > row2Val {
					return false
				}
			} else if columnType == DATETIME_TYPE {
				row1Val := ParseDatetimeOrPanic(row1[columnIndex])
				row2Val := ParseDatetimeOrPanic(row2[columnIndex])
				if row1Val.Before(row2Val) {
					return true
				} else if row1Val.After(row2Val) {
					return false
				}
			} else if columnType == DATE_TYPE {
				row1Val := ParseDateOrPanic(row1[columnIndex])
				row2Val := ParseDateOrPanic(row2[columnIndex])
				if row1Val.Before(row2Val) {
					return true
				} else if row1Val.After(row2Val) {
					return false
				}
			} else {
				row1Val := row1[columnIndex]
				row2Val := row2[columnIndex]
				if row1Val < row2Val {
					return true
				} else if row1Val > row2Val {
					return false
				}
			}
		}
		return false // all values to sort by are equal (hence, not 'less')
	}

	SortRowsBy(isLessFunc).Sort(imc.rows, stable, reverse)
}

func (imc *InMemoryCsv) SampleRowIndicesWithReplacement(numRows, seed int) []int {
	totalRows := imc.NumRows()
	retval := make([]int, numRows)
	for i := 0; i < numRows; i++ {
		retval[i] = rand.Intn(totalRows)
	}
	return retval
}

func (imc *InMemoryCsv) SampleRowIndicesWithoutReplacement(numRows, seed int) []int {
	totalRows := imc.NumRows()
	permuted := rand.Perm(totalRows)
	retval := make([]int, numRows)
	for i := 0; i < numRows; i++ {
		retval[i] = permuted[i]
	}
	return retval
}

func (imc *InMemoryCsv) SampleRowIndices(numRows int, replace bool, seed int) []int {
	// NOTE: Updating global `rand` variable for the life of the proces...
	if seed != 0 {
		rand.Seed(int64(seed))
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
	}
	if replace {
		return imc.SampleRowIndicesWithReplacement(numRows, seed)
	} else {
		return imc.SampleRowIndicesWithoutReplacement(numRows, seed)
	}
}

func (imc *InMemoryCsv) PrintStats() {
	numColumns := imc.NumColumns()
	buffers := make([]bytes.Buffer, numColumns)

	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(numColumns)

	// Process each column concurrently
	for i := 0; i < numColumns; i++ {
		go func(columnIndex int) {
			defer wg.Done()
			bufferedWriter := bufio.NewWriter(&buffers[columnIndex])
			imc.FprintStatsForColumn(bufferedWriter, columnIndex)
			bufferedWriter.Flush()
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Print all column stats in order
	for i := 0; i < numColumns; i++ {
		fmt.Print(buffers[i].String())
	}

	fmt.Printf("Number of rows: %d\n", imc.NumRows())
}

func (imc *InMemoryCsv) GetPrintStatsForColumn(columnIndex int) bytes.Buffer {
	var buf bytes.Buffer
	bufferedWriter := bufio.NewWriter(&buf)
	imc.FprintStatsForColumn(bufferedWriter, columnIndex)
	bufferedWriter.Flush()
	return buf
}

func (imc *InMemoryCsv) FprintStatsForColumn(w io.Writer, columnIndex int) {
	fmt.Fprintf(w, "%d. %s\n", columnIndex+1, imc.header[columnIndex])
	columnType := imc.InferType(columnIndex)
	fmt.Fprintf(w, "  Type: %s\n", ColumnTypeToString(columnType))
	imc.FprintColumnNumberNulls(w, columnIndex)
	if columnType == NULL_TYPE {
		// continue
	} else if columnType == INT_TYPE {
		imc.FprintStatsForColumnAsInt(w, columnIndex)
	} else if columnType == FLOAT_TYPE {
		imc.FprintStatsForColumnAsFloat(w, columnIndex)
	} else if columnType == BOOLEAN_TYPE {
		imc.FprintStatsForColumnAsBoolean(w, columnIndex)
	} else if columnType == DATE_TYPE {
		imc.FprintStatsForColumnAsDatetimeWithFormat(w, columnIndex, time.DateOnly)
	} else if columnType == DATETIME_TYPE {
		imc.FprintStatsForColumnAsDatetimeWithFormat(w, columnIndex, time.RFC3339)
	} else if columnType == STRING_TYPE {
		imc.FprintStatsForColumnAsString(w, columnIndex)
	}
}

func (imc *InMemoryCsv) FprintColumnNumberNulls(w io.Writer, columnIndex int) {
	numNulls := imc.CountNullsInColumn(columnIndex)
	fmt.Fprintf(w, "  Number NULL: %d\n", numNulls)
}

func (imc *InMemoryCsv) CountNullsInColumn(columnIndex int) int {
	numNulls := 0
	for _, row := range imc.rows {
		cell := row[columnIndex]
		if IsNullType(cell) {
			numNulls += 1
		}
	}
	return numNulls
}

func (imc *InMemoryCsv) FprintStatsForColumnAsInt(w io.Writer, columnIndex int) {
	numNulls := imc.CountNullsInColumn(columnIndex)
	intArray := make([]int64, imc.NumRows()-numNulls)
	i := 0
	for _, row := range imc.rows {
		if !IsNullType(row[columnIndex]) {
			intArray[i] = ParseInt64OrPanic(row[columnIndex])
			i++
		}
	}
	ics := NewIntColumnsStats(intArray)
	ics.CalculateAllStats()

	fmt.Fprintf(w, "  Min: %d\n", ics.min)
	fmt.Fprintf(w, "  Max: %d\n", ics.max)
	fmt.Fprintf(w, "  Sum: %d\n", ics.sum)
	fmt.Fprintf(w, "  Mean: %f\n", ics.mean)
	fmt.Fprintf(w, "  Median: %f\n", ics.median)
	fmt.Fprintf(w, "  Standard Deviation: %f\n", ics.stdev)
	fmt.Fprintf(w, "  Unique values: %d\n", len(ics.valueCounts))
	numFrequent := 5
	if numFrequent > len(ics.valueCounts) {
		numFrequent = len(ics.valueCounts)
	}
	fmt.Fprintf(w, "  %d most frequent values:\n", numFrequent)
	for i := 0; i < numFrequent; i++ {
		fmt.Fprintf(w, "      %d: %d\n", ics.valueCounts[i].value, ics.valueCounts[i].count)
	}
}

type IntColumnStats struct {
	array               []int64
	min, max, sum       int64
	mean, median, stdev float64
	valueCounts         []valueCount[int64]
}

func NewIntColumnsStats(intArray []int64) *IntColumnStats {
	ics := new(IntColumnStats)
	ics.array = intArray
	return ics
}

func (ics *IntColumnStats) CalculateAllStats() {
	ics.min = calculateMin(ics.array)
	ics.max = calculateMax(ics.array)
	ics.sum = calculateSum(ics.array)
	ics.mean = calculateMean(ics.array)
	ics.median = calculateMedian(ics.array)
	ics.stdev = calculateStdDev(ics.array)
	ics.valueCounts = getValueCounts(ics.array, cmp.Compare)
}

func (imc *InMemoryCsv) FprintStatsForColumnAsFloat(w io.Writer, columnIndex int) {
	numNulls := imc.CountNullsInColumn(columnIndex)
	floatArray := make([]float64, imc.NumRows()-numNulls)
	i := 0
	for _, row := range imc.rows {
		if !IsNullType(row[columnIndex]) {
			floatArray[i] = ParseFloat64OrPanic(row[columnIndex])
			i++
		}
	}
	fcs := NewFloatColumnsStats(floatArray)
	fcs.CalculateAllStats()

	fmt.Fprintf(w, "  Min: %f\n", fcs.min)
	fmt.Fprintf(w, "  Max: %f\n", fcs.max)
	fmt.Fprintf(w, "  Sum: %f\n", fcs.sum)
	fmt.Fprintf(w, "  Mean: %f\n", fcs.mean)
	fmt.Fprintf(w, "  Median: %f\n", fcs.median)
	fmt.Fprintf(w, "  Standard Deviation: %f\n", fcs.stdev)
	fmt.Fprintf(w, "  Unique values: %d\n", len(fcs.valueCounts))
	numFrequent := 5
	if numFrequent > len(fcs.valueCounts) {
		numFrequent = len(fcs.valueCounts)
	}
	fmt.Fprintf(w, "  %d most frequent values:\n", numFrequent)
	for i := 0; i < numFrequent; i++ {
		fmt.Fprintf(w, "      %f: %d\n", fcs.valueCounts[i].value, fcs.valueCounts[i].count)
	}
}

type FloatColumnStats struct {
	array               []float64
	min, max, sum       float64
	mean, median, stdev float64
	valueCounts         []valueCount[float64]
}

func NewFloatColumnsStats(floatArray []float64) *FloatColumnStats {
	fcs := new(FloatColumnStats)
	fcs.array = floatArray
	return fcs
}

func (fcs *FloatColumnStats) CalculateAllStats() {
	fcs.min = calculateMin(fcs.array)
	fcs.max = calculateMax(fcs.array)
	fcs.sum = calculateSum(fcs.array)
	fcs.mean = calculateMean(fcs.array)
	fcs.median = calculateMedian(fcs.array)
	fcs.stdev = calculateStdDev(fcs.array)
	fcs.valueCounts = getValueCounts(fcs.array, cmp.Compare)
}

func (imc *InMemoryCsv) FprintStatsForColumnAsBoolean(w io.Writer, columnIndex int) {
	numTrue := 0
	numFalse := 0
	for _, row := range imc.rows {
		value := strings.Trim(row[columnIndex], " ")
		if !IsNullType(value) {
			bval := ParseBooleanOrPanic(value)
			if bval {
				numTrue++
			} else {
				numFalse++
			}
		}
	}
	fmt.Fprintf(w, "  Number TRUE: %d\n", numTrue)
	fmt.Fprintf(w, "  Number FALSE: %d\n", numFalse)
}

func (imc *InMemoryCsv) FprintStatsForColumnAsDatetimeWithFormat(w io.Writer, columnIndex int, format string) {
	numNulls := imc.CountNullsInColumn(columnIndex)
	dateArray := make([]time.Time, imc.NumRows()-numNulls)
	i := 0
	for _, row := range imc.rows {
		if !IsNullType(row[columnIndex]) {
			dateArray[i] = ParseDatetimeOrPanic(row[columnIndex])
			i++
		}
	}
	dcs := NewDateColumnsStats(dateArray, format)
	dcs.CalculateAllStats()

	fmt.Fprintf(w, "  Min: %s\n", dcs.min.Format(format))
	fmt.Fprintf(w, "  Max: %s\n", dcs.max.Format(format))
	fmt.Fprintf(w, "  Unique values: %d\n", len(dcs.valueCounts))
	numFrequent := 5
	if numFrequent > len(dcs.valueCounts) {
		numFrequent = len(dcs.valueCounts)
	}
	fmt.Fprintf(w, "  %d most frequent values:\n", numFrequent)
	for i := 0; i < numFrequent; i++ {
		fmt.Fprintf(w, "      %s: %d\n", dcs.valueCounts[i].value, dcs.valueCounts[i].count)
	}
}

type DateColumnStats struct {
	array       []time.Time
	min, max    time.Time
	valueCounts []valueCount[time.Time]
	format      string
}

func NewDateColumnsStats(dateArray []time.Time, format string) *DateColumnStats {
	dcs := new(DateColumnStats)
	dcs.array = dateArray
	dcs.format = format
	return dcs
}

func (dcs *DateColumnStats) CalculateAllStats() {
	dcs.CalculateMin()
	dcs.CalculateMax()
	dcs.valueCounts = getValueCounts(dcs.array, time.Time.Compare)
}

func (dcs *DateColumnStats) CalculateMin() {
	for i, dateVal := range dcs.array {
		if i == 0 || dateVal.Before(dcs.min) {
			dcs.min = dateVal
		}
	}
}

func (dcs *DateColumnStats) CalculateMax() {
	for i, dateVal := range dcs.array {
		if i == 0 || dateVal.After(dcs.max) {
			dcs.max = dateVal
		}
	}
}

func (imc *InMemoryCsv) FprintStatsForColumnAsString(w io.Writer, columnIndex int) {
	numNulls := imc.CountNullsInColumn(columnIndex)
	stringArray := make([]string, imc.NumRows()-numNulls)
	i := 0
	for _, row := range imc.rows {
		if !IsNullType(row[columnIndex]) {
			stringArray[i] = row[columnIndex]
			i++
		}
	}
	scs := NewStringColumnsStats(stringArray)
	scs.CalculateAllStats()

	fmt.Fprintf(w, "  Unique values: %d\n", len(scs.valueCounts))
	fmt.Fprintf(w, "  Max length: %d\n", scs.maxLength)
	numFrequent := 5
	if numFrequent > len(scs.valueCounts) {
		numFrequent = len(scs.valueCounts)
	}
	fmt.Fprintf(w, "  %d most frequent values:\n", numFrequent)
	for i := 0; i < numFrequent; i++ {
		fmt.Fprintf(w, "      %s: %d\n", scs.valueCounts[i].value, scs.valueCounts[i].count)
	}
}

type StringColumnStats struct {
	array       []string
	valueCounts []valueCount[string]
	maxLength   int
}

func NewStringColumnsStats(stringArray []string) *StringColumnStats {
	scs := new(StringColumnStats)
	scs.array = stringArray
	return scs
}

func (scs *StringColumnStats) CalculateAllStats() {
	scs.CalculateMaxLength()
	scs.valueCounts = getValueCounts(scs.array, cmp.Compare)
}

func (scs *StringColumnStats) CalculateMaxLength() {
	scs.maxLength = -1
	for _, elem := range scs.array {
		if len(elem) > scs.maxLength {
			scs.maxLength = len(elem)
		}
	}
}

func calculateMin[E int64 | float64](values []E) E {
	switch len(values) {
	case 0, 1:
		panic("need more than 1 value for comparison")
	case 2:
		return min(values[0], values[1])
	default:
		minX := min(values[0], values[1])
		for _, x := range values[2:] {
			minX = min(minX, x)
		}
		return minX
	}
}

func calculateMax[E float64 | int64](values []E) E {
	switch len(values) {
	case 0, 1:
		panic("need more than 1 value for comparison")
	case 2:
		return max(values[0], values[1])
	default:
		maxX := max(values[0], values[1])
		for _, x := range values[2:] {
			maxX = max(maxX, x)
		}
		return maxX
	}
}

func calculateSum[E int64 | float64](values []E) E {
	var sum E
	for _, x := range values {
		sum += x
	}
	return sum
}

func calculateMean[E int64 | float64](values []E) float64 {
	return float64(calculateSum(values)) / float64(len(values))
}

func calculateMedian[E int64 | float64](values []E) float64 {
	sorted := slices.Sorted(slices.Values(values))
	switch n := len(sorted); {
	default:
		return float64(sorted[n/2])
	case n%2 == 0:
		return calculateMean(sorted[n/2-1 : n/2+1])
	}
}

// Calculate the sample standard deviation (n-1); as oppossed to
// standard deviation (n).
func calculateStdDev[E int64 | float64](values []E) float64 {
	mean := calculateMean(values)
	n := len(values)

	sum := 0.0
	for _, x := range values {
		deviation := float64(x) - mean
		sum += deviation * deviation
	}

	variance := sum / float64(n-1)

	return math.Sqrt(variance)
}

// Inferred Types
type IT interface {
	bool | float64 | int64 | string | time.Time
}

type valueCount[K IT] struct {
	value K
	count int
}

// getValueCounts iterates and counts occurences of values,
// returning a slice of valueCount[E] sorted with the
// highest counts first.
//
// In the case of different values with the same count,
// tieBreaker enforces a consistent sort of the same-count
// values.
func getValueCounts[E IT](values []E, tieBreaker func(a, b E) int) []valueCount[E] {
	counts := map[E]int{}
	for _, x := range values {
		counts[x] += 1
	}

	var valueCounts []valueCount[E]
	for k, v := range maps.All(counts) {
		valueCounts = append(valueCounts, valueCount[E]{k, v})
	}

	slices.SortFunc(valueCounts, func(a, b valueCount[E]) int {
		if n := cmp.Compare(b.count, a.count); n != 0 {
			return n
		}
		return tieBreaker(a.value, b.value)
	})

	return valueCounts
}

// boolCompare produces a consistent sort of
// valueCount[bool] items: true, then false; use as
// a tieBreaker for [getValueCounts].
//
// Currently not used.
func boolCompare(a, b bool) int {
	if a {
		return -1
	}
	return 1
}
