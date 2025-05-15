package stats

import (
	"math"
	"reflect"
	"testing"
)

func TestStats_boolean(t *testing.T) {
	got := booleanStats([]bool{true, true, false, true})

	want := boolCounts{true: 3, false: 1}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %v\nwant: %v\n", got, want)
	}
}

var (
	posInf = math.Inf(0)

	// equal-ish for floats
	equal = func(a, b float64) bool {
		if a == posInf || b == posInf {
			if a == posInf && b == posInf {
				return true
			}
			return false
		}
		return math.Abs(a-b) <= 1e-14
	}
)

func TestStats_numeric(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		testCases := []struct {
			nums []int

			min         int
			median      float64
			max         int
			sum         int
			mean        float64
			stdDev      float64
			uniqueCount int
			valueCounts []intCount
		}{
			{
				nums:        []int{1, 2, 3},
				min:         1,
				median:      2,
				max:         3,
				sum:         6,
				mean:        2,
				stdDev:      1,
				uniqueCount: 3,
				valueCounts: []intCount{
					{1, 1},
					{2, 1},
					{3, 1},
				},
			},
			{
				nums:        []int{1, 2, 3, 4},
				min:         1,
				median:      2.5,
				max:         4,
				sum:         10,
				mean:        2.5,
				stdDev:      1.2909944487358055997817,
				uniqueCount: 4,
				valueCounts: []intCount{
					{1, 1},
					{2, 1},
					{3, 1},
					{4, 1},
				},
			},
			{
				nums:        []int{1, 2, 3, 3, 4, 4, 4},
				min:         1,
				median:      3,
				max:         4,
				sum:         21,
				mean:        3,
				stdDev:      1.15470053837925,
				uniqueCount: 4,
				valueCounts: []intCount{
					{4, 3},
					{3, 2},
					{1, 1},
					{2, 1},
				},
			},
			// e17, biggest exponent on my machine before overflow
			{
				nums:        []int{1e17, 2e17, 3e17, 3e17, 4e17, 4e17, 4e17},
				min:         1e17,
				median:      3e17,
				max:         4e17,
				sum:         21e17,
				mean:        3e17,
				stdDev:      115_470_053_837_925_136, // WolframAlpha computed 115_470_053_837_925_168
				uniqueCount: 4,
				valueCounts: []intCount{
					{4e17, 3},
					{3e17, 2},
					{1e17, 1},
					{2e17, 1},
				},
			},
			// e18, overflows, biggest exponent on my machine the compiler allows
			{
				nums:        []int{1e18, 2e18, 3e18, 3e18, 4e18, 4e18, 4e18},
				min:         1e18,
				median:      3e18,
				max:         4e18,
				sum:         2_553_255_926_290_448_384,
				mean:        364_750_846_612_921_216,
				stdDev:      3_071_692_440_739_881_984,
				uniqueCount: 4,
				valueCounts: []intCount{
					{4e18, 3},
					{3e18, 2},
					{1e18, 1},
					{2e18, 1},
				},
			},
		}

		for _, tc := range testCases {
			assertNumericStats(`
for %v:
  min:    %d; %d
  median: %.14f; %.14f
  max:    %d; %d
  sum:    %d; %d
  mean:   %.14f; %.14f
  stdDev: %.14f; %.14f
  unique: %d; %d
  counts: %v; %v
`,

				tc.nums,
				tc.min,
				tc.median,
				tc.max,
				tc.sum,
				tc.mean,
				tc.stdDev,
				tc.uniqueCount,
				tc.valueCounts,

				t)
		}
	})

	t.Run("floats", func(t *testing.T) {
		testCases := []struct {
			nums []float64

			min         float64
			median      float64
			max         float64
			sum         float64
			mean        float64
			stdDev      float64
			uniqueCount int
			valueCounts []floatCount
		}{
			{
				nums:        []float64{1, 2, 3},
				min:         1,
				median:      2.0,
				max:         3,
				sum:         6,
				mean:        2.0,
				stdDev:      1.0,
				uniqueCount: 3,
				valueCounts: []floatCount{
					{1, 1},
					{2, 1},
					{3, 1},
				},
			},
			{
				nums:        []float64{1, 2, 3, 4},
				min:         1,
				median:      2.5,
				max:         4,
				sum:         10,
				mean:        2.5,
				stdDev:      1.2909944487358055997817,
				uniqueCount: 4,
				valueCounts: []floatCount{
					{1, 1},
					{2, 1},
					{3, 1},
					{4, 1},
				},
			},
			// e22, biggest exponent on my machine before overflow
			{
				nums:        []float64{1e22, 2e22, 3e22, 3e22, 4e22, 4e22, 4e22},
				min:         10000000000000000000000,
				median:      30000000000000000000000,
				max:         40000000000000000000000,
				sum:         210000000000000012582912,
				mean:        30000000000000000000000,
				stdDev:      11_547_005_383_792_515_874_816, // WolframAlpha computed 11_547_005_383_792_516_841_623
				uniqueCount: 4,
				valueCounts: []floatCount{
					{4e22, 3},
					{3e22, 2},
					{1e22, 1},
					{2e22, 1},
				},
			},
			// e307, overflows, biggest exponent on my machine the compiler allows
			{
				nums:        []float64{1e307, 2e307, 3e307, 3e307, 4e307, 4e307, 4e307},
				min:         9999999999999999860310597602564577717002641838126363875249660735883565852672743849064846414228960666786379280392654615393353172850252103336275952370615397010730691664689375178569039851073146339641623266071126720011020169553304018596457812688561947201171488461172921822139066929851282122002676667750021070848.0,
				median:      29999999999999998333531599348493850865774979866354987833591944435489734118926205023937106824584340884760409408280650665341030240930446782526054114365850153070209701066048487835703573958790891195463793895486513170874712543319959559957616903615141848547979922603491167024466633992206320339533970484285057007616.0,
				max:         39999999999999999441242390410258310868010567352505455500998642943534263410690975396259385656915842667145517121570618461573412691401008413345103809482461588042922766658757500714276159404292585358566493064284506880044080678213216074385831250754247788804685953844691687288556267719405128488010706671000084283392.0,
				sum:         posInf,
				mean:        posInf,
				stdDev:      posInf,
				uniqueCount: 4,
				valueCounts: []floatCount{
					{4e307, 3},
					{3e307, 2},
					{1e307, 1},
					{2e307, 1},
				},
			},
		}

		for _, tc := range testCases {
			assertNumericStats(`
for %v:
  min:    %.14f; %.14f
  median: %.14f; %.14f
  max:    %.14f; %.14f
  sum:    %.14f; %.14f
  mean:   %.14f; %.14f
  stdDev: %.14f; %.14f
  unique: %d; %d
  counts: %v; %v
`,

				tc.nums,
				tc.min,
				tc.median,
				tc.max,
				tc.sum,
				tc.mean,
				tc.stdDev,
				tc.uniqueCount,
				tc.valueCounts,

				t)
		}
	})
}

func assertNumericStats[Num numeric](
	format string,

	nums []Num,
	min Num,
	median float64,
	max Num,
	sum Num,
	mean float64,
	stdDev float64,
	uniqueCount int,
	valueCounts []struct {
		value Num
		count int
	},

	t *testing.T) {

	t.Helper()

	_min, _median, _max, _sum, _mean, _stdDev, _uniqueCount, _valueCounts := numericStats(nums)

	if _min != min ||
		!equal(_median, median) ||
		_max != max ||
		_sum != sum ||
		!equal(_mean, mean) ||
		!equal(_stdDev, stdDev) ||
		_uniqueCount != uniqueCount ||
		!reflect.DeepEqual(_valueCounts, valueCounts) {
		t.Errorf(
			format,

			nums,

			_min, min,
			_median, median,
			_max, max,
			_sum, sum,
			_mean, mean,
			_stdDev, stdDev,
			_uniqueCount, uniqueCount,
			_valueCounts, valueCounts,
		)
	}
}

func TestStats_string(t *testing.T) {
	strings := []string{"foo", "baz", "baker", "bar", "foo"}

	_maxLen, _uniqueCount, _valueCounts := stringStats(strings)

	// wants
	var (
		maxLen      = 5
		uniqueCount = 4
		valueCounts = []stringCount{
			{s: "foo", count: 2},
			{s: "baker", count: 1},
			{s: "bar", count: 1},
			{s: "baz", count: 1},
		}
	)

	if _maxLen != maxLen ||
		_uniqueCount != uniqueCount ||
		!reflect.DeepEqual(_valueCounts, valueCounts) {
		t.Errorf(`
            for %v:
              maxLen:      %d; %d
              uniqueCount: %d; %d
              valueCounts: %v; %v`,

			strings,

			_maxLen, maxLen,
			_uniqueCount, uniqueCount,
			_valueCounts, valueCounts)
	}
}
