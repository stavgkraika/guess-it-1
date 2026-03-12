package internal

import (
	"math"
	"testing"
)

// TestMedianSorted tests the medianSorted function
func TestMedianSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    []int64
		expected int64
	}{
		{"empty slice", []int64{}, 0},
		{"single element", []int64{5}, 5},
		{"odd length", []int64{1, 2, 3, 4, 5}, 3},
		{"even length", []int64{1, 2, 3, 4}, 2},
		{"two elements", []int64{10, 20}, 15},
		{"negative values", []int64{-5, -3, -1, 0, 2}, -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := medianSorted(test.input)
			if result != test.expected {
				t.Errorf("medianSorted(%v) = %d, expected %d", test.input, result, test.expected)
			}
		})
	}
}

// TestMeanStd tests the MeanStd function
func TestMeanStd(t *testing.T) {
	tests := []struct {
		name        string
		input       []int64
		expectedMu  float64
		expectedStd float64
	}{
		{"empty slice", []int64{}, 0.0, 0.0},
		{"single element", []int64{5}, 5.0, 0.0},
		{"uniform values", []int64{10, 10, 10, 10}, 10.0, 0.0},
		{"simple sequence", []int64{1, 2, 3, 4, 5}, 3.0, math.Sqrt(2.0)},
		{"negative values", []int64{-10, -5, 0, 5, 10}, 0.0, math.Sqrt(50.0)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mu, std := MeanStd(test.input)
			if math.Abs(mu-test.expectedMu) > 0.0001 {
				t.Errorf("MeanStd(%v) mean = %f, expected %f", test.input, mu, test.expectedMu)
			}
			if math.Abs(std-test.expectedStd) > 0.0001 {
				t.Errorf("MeanStd(%v) std = %f, expected %f", test.input, std, test.expectedStd)
			}
		})
	}
}

// TestMedianMAD tests the MedianMAD function
func TestMedianMAD(t *testing.T) {
	tests := []struct {
		name           string
		input          []int64
		expectedMedian int64
		expectedMAD    float64
	}{
		{"empty slice", []int64{}, 0, 0.0},
		{"single element", []int64{5}, 5, 0.0},
		{"uniform values", []int64{10, 10, 10, 10}, 10, 0.0},
		{"simple sequence", []int64{1, 2, 3, 4, 5}, 3, 1.0},
		{"with outlier", []int64{1, 2, 3, 4, 100}, 3, 1.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			median, mad := MedianMAD(test.input)
			if median != test.expectedMedian {
				t.Errorf("MedianMAD(%v) median = %d, expected %d", test.input, median, test.expectedMedian)
			}
			if math.Abs(mad-test.expectedMAD) > 0.0001 {
				t.Errorf("MedianMAD(%v) MAD = %f, expected %f", test.input, mad, test.expectedMAD)
			}
		})
	}
}

// TestTypicalStep tests the TypicalStep function
func TestTypicalStep(t *testing.T) {
	tests := []struct {
		name     string
		input    []int64
		expected float64
	}{
		{"empty slice", []int64{}, 0.0},
		{"single element", []int64{5}, 5.0},
		{"uniform diffs", []int64{10, 10, 10, 10}, 10.0},
		{"varied diffs", []int64{1, 2, 3, 4, 5}, 3.0},
		{"two elements", []int64{5, 15}, 10.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := TypicalStep(test.input)
			if math.Abs(result-test.expected) > 0.0001 {
				t.Errorf("TypicalStep(%v) = %f, expected %f", test.input, result, test.expected)
			}
		})
	}
}

// TestMeanStdLargeDataset tests MeanStd with larger dataset
func TestMeanStdLargeDataset(t *testing.T) {
	// Create a dataset with known mean and std
	data := make([]int64, 100)
	for i := 0; i < 100; i++ {
		data[i] = int64(i)
	}

	mu, std := MeanStd(data)

	// Mean should be 49.5
	expectedMu := 49.5
	if math.Abs(mu-expectedMu) > 0.1 {
		t.Errorf("MeanStd large dataset mean = %f, expected %f", mu, expectedMu)
	}

	// Std should be approximately 28.87
	if std < 28.0 || std > 30.0 {
		t.Errorf("MeanStd large dataset std = %f, expected around 28.87", std)
	}
}

// TestMedianMADWithUnsortedData tests MedianMAD with unsorted input
func TestMedianMADWithUnsortedData(t *testing.T) {
	// MedianMAD should handle unsorted data
	unsorted := []int64{5, 1, 4, 2, 3}
	median, mad := MedianMAD(unsorted)

	if median != 3 {
		t.Errorf("MedianMAD unsorted median = %d, expected 3", median)
	}

	if math.Abs(mad-1.0) > 0.0001 {
		t.Errorf("MedianMAD unsorted MAD = %f, expected 1.0", mad)
	}
}

// TestTypicalStepWithUnsortedData tests TypicalStep with unsorted input
func TestTypicalStepWithUnsortedData(t *testing.T) {
	// TypicalStep should handle unsorted data
	unsorted := []int64{10, 5, 20, 15, 25}
	result := TypicalStep(unsorted)

	// Median of [5, 10, 15, 20, 25] is 15
	if math.Abs(result-15.0) > 0.0001 {
		t.Errorf("TypicalStep unsorted = %f, expected 15.0", result)
	}
}

// TestMeanStdWithNegativeValues tests MeanStd with all negative values
func TestMeanStdWithNegativeValues(t *testing.T) {
	data := []int64{-100, -50, -25, -10, -5}
	mu, std := MeanStd(data)

	expectedMu := -38.0
	if math.Abs(mu-expectedMu) > 0.1 {
		t.Errorf("MeanStd negative values mean = %f, expected %f", mu, expectedMu)
	}

	if std <= 0 {
		t.Errorf("MeanStd negative values std = %f, expected positive value", std)
	}
}

// TestMedianMADEvenLength tests MedianMAD with even-length slice
func TestMedianMADEvenLength(t *testing.T) {
	data := []int64{1, 2, 3, 4, 5, 6}
	median, mad := MedianMAD(data)

	// Median of even-length should be average of middle two: (3+4)/2 = 3
	if median != 3 {
		t.Errorf("MedianMAD even length median = %d, expected 3", median)
	}

	// MAD should be calculated correctly
	if mad < 0 {
		t.Errorf("MedianMAD even length MAD = %f, expected non-negative", mad)
	}
}
