package stats

import (
	"cmp"
	"math"
	"slices"
)

type boolCounts struct {
	true  int
	false int
}

func booleanStats(bools []bool) (valueCounts boolCounts) {
	for _, b := range bools {
		switch b {
		case true:
			valueCounts.true++
		case false:
			valueCounts.false++
		}
	}
	return valueCounts
}

type floatCount = struct {
	value float64
	count int
}

type floatStats struct {
	min         float64
	median      float64
	max         float64
	sum         float64
	mean        float64
	stdDev      float64
	uniqueCount int
	valueCounts []floatCount
}

type intCount = struct {
	value int
	count int
}

type intStats struct {
	min         int
	median      float64
	max         int
	sum         int
	mean        float64
	stdDev      float64
	uniqueCount int
	valueCounts []intCount
}

type numeric interface{ float64 | int }

func numericStats[Num numeric](nums []Num) (
	minN Num,
	medianN float64,
	maxN Num,
	sum Num,
	mean float64,
	stdDev float64,
	uniqueCount int,
	valueCounts []struct {
		value Num
		count int
	}) {

	// internal type alias
	type _valueCount = struct {
		value Num
		count int
	}

	n := len(nums)

	// min, median, max
	slices.Sort(nums)
	minN, maxN = nums[0], nums[n-1]
	switch {
	case n%2 == 0:
		i, j := n/2-1, n/2
		medianN = float64(nums[i]+nums[j]) / 2
	default:
		i := (n - 1) / 2
		medianN = float64(nums[i])
	}

	// sum and raw value counts
	m := make(map[Num]int)
	for _, x := range nums {
		sum += x
		m[x]++
	}

	// mean
	mean = float64(sum) / float64(n)

	// stdDev
	acc := 0.0
	for _, x := range nums {
		y := float64(x) - mean
		acc += (y * y)
	}
	stdDev = math.Sqrt(acc / float64(n-1))

	// uniqueCount
	uniqueCount = len(m)

	// convert raw map to properly sorted slice of value counts
	valueCounts = make([]_valueCount, 0, len(m))
	for val, ct := range m {
		valueCounts = append(valueCounts, _valueCount{val, ct})
	}
	slices.SortFunc(valueCounts, func(a, b _valueCount) int {
		// count (descending); break tie on value (ascending)
		if a.count == b.count {
			return cmp.Compare(a.value, b.value)
		}
		return b.count - a.count
	})

	return minN,
		medianN,
		maxN,
		sum,
		mean,
		stdDev,
		uniqueCount,
		valueCounts
}

type stringCount struct {
	s     string
	count int
}

func stringStats(strings []string) (
	maxLen int,
	uniqueCount int,
	valueCounts []stringCount,
) {
	m := make(map[string]int)
	for _, s := range strings {
		maxLen = max(len(s), maxLen)
		m[s]++
	}

	// uniqueCount
	uniqueCount = len(m)

	// convert raw map to properly sorted slice of value counts
	valueCounts = make([]stringCount, 0, len(m))
	for s, ct := range m {
		valueCounts = append(valueCounts, stringCount{s, ct})
	}
	slices.SortFunc(valueCounts, func(a, b stringCount) int {
		// count (descending); break tie on value (ascending)
		if a.count == b.count {
			return cmp.Compare(a.s, b.s)
		}
		return b.count - a.count
	})

	return maxLen,
		uniqueCount,
		valueCounts
}
