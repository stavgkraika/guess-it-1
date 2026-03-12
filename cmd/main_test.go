package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRunSimple tests the run function with simple input
func TestRunSimple(t *testing.T) {
	input := "10\n20\n30\n40\n50\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}

	// Each line should have content
	for i, line := range lines {
		if line == "" {
			t.Errorf("Line %d is empty", i)
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			t.Errorf("Line %d: expected 2 values, got %d", i, len(parts))
		}
	}
}

// TestRunEmpty tests with empty input
func TestRunEmpty(t *testing.T) {
	input := ""
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

// TestRunSingleValue tests with single input value
func TestRunSingleValue(t *testing.T) {
	input := "100\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 1 {
		t.Errorf("Expected 1 output line, got %d", len(lines))
	}

	if lines[0] == "" {
		t.Error("Output line is empty")
	}
}

// TestRunNegativeValues tests with negative values
func TestRunNegativeValues(t *testing.T) {
	input := "-50\n-40\n-30\n-20\n-10\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}
}

// TestRunConstantValues tests with constant input
func TestRunConstantValues(t *testing.T) {
	input := "50\n50\n50\n50\n50\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}
}

// TestRunLargeInput tests with larger input
func TestRunLargeInput(t *testing.T) {
	var inputBuilder strings.Builder
	for i := 0; i < 100; i++ {
		inputBuilder.WriteString("100\n")
	}

	in := strings.NewReader(inputBuilder.String())

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 100 {
		t.Errorf("Expected 100 output lines, got %d", len(lines))
	}
}

// TestRunIncreasingSequence tests with increasing values
func TestRunIncreasingSequence(t *testing.T) {
	var inputBuilder strings.Builder
	for i := 0; i < 20; i++ {
		inputBuilder.WriteString(strings.Trim(strings.Repeat(" ", i*10), " "))
		if i > 0 {
			inputBuilder.WriteString("\n")
		}
	}

	// Use simpler input
	input := ""
	for i := 0; i < 20; i++ {
		if i > 0 {
			input += "\n"
		}
		input += "10"
	}
	input += "\n"

	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 20 {
		t.Errorf("Expected 20 output lines, got %d", len(lines))
	}
}

// TestRunZeroValues tests with zero values
func TestRunZeroValues(t *testing.T) {
	input := "0\n0\n0\n0\n0\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 output lines, got %d", len(lines))
	}
}

// TestRunVolatileData tests with volatile data
func TestRunVolatileData(t *testing.T) {
	input := "10\n100\n20\n90\n30\n80\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 6 {
		t.Errorf("Expected 6 output lines, got %d", len(lines))
	}
}

// TestRunLargeValues tests with large values
func TestRunLargeValues(t *testing.T) {
	input := "1000000\n1000100\n1000200\n"
	in := strings.NewReader(input)

	var out bytes.Buffer

	run(in, &out)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("Expected 3 output lines, got %d", len(lines))
	}
}

// TestMainExists verifies main function exists
func TestMainExists(t *testing.T) {
	// This test ensures main() compiles and can be referenced
	// We don't call it directly as it uses os.Stdin/Stdout
	t.Log("Main function exists and compiles")
}
