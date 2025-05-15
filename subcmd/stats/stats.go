package stats

import (
	"math"
	"slices"
)

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
	valueCounts = make([]struct {
		value Num
		count int
	}, 0, len(m))
	for val, ct := range m {
		valueCounts = append(valueCounts, struct {
			value Num
			count int
		}{val, ct})
	}
	slices.SortFunc(valueCounts, func(a, b struct {
		value Num
		count int
	}) int {
		// count (descending); break tie on value (ascending)
		if a.count == b.count {
			return int(a.value - b.value)
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
