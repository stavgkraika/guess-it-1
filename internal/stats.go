package internal

import (
	"math"
	"sort"
)

// medianSorted computes the median of a sorted slice of int64 values.
// For odd-length slices, it returns the middle element.
// For even-length slices, it returns the average of the two middle elements.
func medianSorted(sorted []int64) int64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	
	mid := n / 2
	
	// Odd number of elements: return the middle one
	if n%2 == 1 {
		return sorted[mid]
	}
	
	// Even number of elements: return the average of the two middle ones
	return (sorted[mid-1] + sorted[mid]) / 2
}

// MeanStd computes mean and population stddev from values.
// The mean is the average of all values.
// The standard deviation measures how spread out the values are from the mean.
// This uses the population formula (dividing by n, not n-1).
func MeanStd(vals []int64) (mean float64, std float64) {
	if len(vals) == 0 {
		return 0, 0
	}
	
	// Calculate the sum of all values
	var sum int64
	for _, v := range vals {
		sum += v
	}
	
	// Compute the mean (average)
	n := float64(len(vals))
	mean = float64(sum) / n

	// Calculate the sum of squared deviations from the mean
	var ss float64
	for _, v := range vals {
		d := float64(v) - mean
		ss += d * d
	}
	
	// Standard deviation is the square root of the variance
	std = math.Sqrt(ss / n)
	return mean, std
}

// MedianMAD computes median and MAD from vals.
// MAD (Median Absolute Deviation) is a robust measure of statistical dispersion.
// It's less sensitive to outliers than standard deviation.
// It allocates copies for sorting (WindowN is small, so this is fine).
func MedianMAD(vals []int64) (median int64, mad float64) {
	if len(vals) == 0 {
		return 0, 0
	}
	
	// Create a sorted copy of the values
	sorted := make([]int64, len(vals))
	copy(sorted, vals)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	
	// Compute the median of the values
	median = medianSorted(sorted)

	// Compute the absolute deviations from the median
	dev := make([]int64, len(sorted))
	for i, v := range sorted {
		dev[i] = abs64(v - median)
	}
	
	// Sort the deviations
	sort.Slice(dev, func(i, j int) bool { return dev[i] < dev[j] })
	
	// MAD is the median of the absolute deviations
	mad = float64(medianSorted(dev))
	return median, mad
}

// TypicalStep returns median of diffs (absolute successive differences).
// This gives us a robust estimate of the typical change between consecutive values.
// diffs should already be absolute values. Returns 0 if unavailable.
func TypicalStep(diffs []int64) float64 {
	if len(diffs) == 0 {
		return 0
	}
	
	// Create a sorted copy of the diffs
	cp := make([]int64, len(diffs))
	copy(cp, diffs)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	
	// Return the median diff as the typical step size
	return float64(medianSorted(cp))
}
