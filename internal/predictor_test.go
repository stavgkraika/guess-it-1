package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// TestNewPredictor tests the NewPredictor constructor
func TestNewPredictor(t *testing.T) {
	p := NewPredictor()

	if p == nil {
		t.Fatal("NewPredictor returned nil")
	}

	if p.k != 1.35 {
		t.Errorf("Expected k = 1.35, got %f", p.k)
	}

	if p.kMin != 1.10 {
		t.Errorf("Expected kMin = 1.10, got %f", p.kMin)
	}

	if p.kMax != 2.20 {
		t.Errorf("Expected kMax = 2.20, got %f", p.kMax)
	}

	if p.hasPrevY {
		t.Error("Expected hasPrevY to be false")
	}

	if p.hasPrevIV {
		t.Error("Expected hasPrevIV to be false")
	}

	if cap(p.tmpVals) != WindowN {
		t.Errorf("Expected tmpVals capacity %d, got %d", WindowN, cap(p.tmpVals))
	}

	if cap(p.tmpDiffs) != WindowN {
		t.Errorf("Expected tmpDiffs capacity %d, got %d", WindowN, cap(p.tmpDiffs))
	}
}

// TestPredictorRunSimple tests basic prediction with simple input
func TestPredictorRunSimple(t *testing.T) {
	p := NewPredictor()

	input := "10\n20\n30\n40\n50\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}

	// Each line should have two integers
	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > upper {
			t.Errorf("Line %d: lower bound %d > upper bound %d", i, lower, upper)
		}
	}
}

// TestPredictorRunEmpty tests with empty input
func TestPredictorRunEmpty(t *testing.T) {
	p := NewPredictor()

	input := ""
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

// TestPredictorRunSingleValue tests with single input value
func TestPredictorRunSingleValue(t *testing.T) {
	p := NewPredictor()

	input := "100\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 1 {
		t.Errorf("Expected 1 output line, got %d", len(lines))
	}

	var lower, upper int64
	_, err := fmt.Sscanf(lines[0], "%d %d", &lower, &upper)
	if err != nil {
		t.Errorf("Failed to parse output: %v", err)
	}

	// Should contain the value 100
	if lower > 100 || upper < 100 {
		t.Errorf("Interval [%d, %d] does not contain input value 100", lower, upper)
	}
}

// TestPredictorRunNegativeValues tests with negative values
func TestPredictorRunNegativeValues(t *testing.T) {
	p := NewPredictor()

	input := "-50\n-40\n-30\n-20\n-10\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}

	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > upper {
			t.Errorf("Line %d: lower bound %d > upper bound %d", i, lower, upper)
		}
	}
}

// TestPredictorRunConstantValues tests with constant input
func TestPredictorRunConstantValues(t *testing.T) {
	p := NewPredictor()

	input := "50\n50\n50\n50\n50\n50\n50\n50\n50\n50\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 10 {
		t.Errorf("Expected 10 output lines, got %d", len(lines))
	}

	// All intervals should contain 50
	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > 50 || upper < 50 {
			t.Errorf("Line %d: interval [%d, %d] does not contain 50", i, lower, upper)
		}
	}
}

// TestPredictorRunLargeValues tests with large values
func TestPredictorRunLargeValues(t *testing.T) {
	p := NewPredictor()

	input := "1000000\n1000100\n1000200\n1000300\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 output lines, got %d", len(lines))
	}

	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > upper {
			t.Errorf("Line %d: lower bound %d > upper bound %d", i, lower, upper)
		}
	}
}

// TestPredictorRunVolatileData tests with volatile/random data
func TestPredictorRunVolatileData(t *testing.T) {
	p := NewPredictor()

	input := "10\n100\n20\n90\n30\n80\n40\n70\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 8 {
		t.Errorf("Expected 8 output lines, got %d", len(lines))
	}

	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > upper {
			t.Errorf("Line %d: lower bound %d > upper bound %d", i, lower, upper)
		}
	}
}

// TestPredictorRunIncreasingSequence tests with steadily increasing values
func TestPredictorRunIncreasingSequence(t *testing.T) {
	p := NewPredictor()

	var inputBuilder strings.Builder
	for i := 0; i < 20; i++ {
		inputBuilder.WriteString(fmt.Sprintf("%d\n", i*10))
	}

	in := bufio.NewReader(strings.NewReader(inputBuilder.String()))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 20 {
		t.Errorf("Expected 20 output lines, got %d", len(lines))
	}
}

// TestPredictorRunDecreasingSequence tests with steadily decreasing values
func TestPredictorRunDecreasingSequence(t *testing.T) {
	p := NewPredictor()

	var inputBuilder strings.Builder
	for i := 20; i > 0; i-- {
		inputBuilder.WriteString(fmt.Sprintf("%d\n", i*10))
	}

	in := bufio.NewReader(strings.NewReader(inputBuilder.String()))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 20 {
		t.Errorf("Expected 20 output lines, got %d", len(lines))
	}
}

// TestPredictorRunZeroValues tests with zero values
func TestPredictorRunZeroValues(t *testing.T) {
	p := NewPredictor()

	input := "0\n0\n0\n0\n0\n"
	in := bufio.NewReader(strings.NewReader(input))

	var out bytes.Buffer
	outWriter := bufio.NewWriter(&out)

	p.Run(in, outWriter)
	outWriter.Flush()

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}

	// All intervals should contain 0
	for i, line := range lines {
		var lower, upper int64
		_, err := fmt.Sscanf(line, "%d %d", &lower, &upper)
		if err != nil {
			t.Errorf("Line %d: failed to parse output: %v", i, err)
		}
		if lower > 0 || upper < 0 {
			t.Errorf("Line %d: interval [%d, %d] does not contain 0", i, lower, upper)
		}
	}
}
