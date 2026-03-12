// Package main is the entry point for the guess-it-1 application.
// This program implements a prediction system that reads input data,
// processes it through a predictor, and writes the results to output.
package main

import (
	"bufio" // Provides buffered I/O operations for efficient reading and writing
	"io"    // Provides basic I/O interfaces
	"os"    // Provides platform-independent interface to operating system functionality

	"guess-it-1/internal" // Internal package containing the predictor implementation
)

// run executes the main logic with provided input and output streams.
// This function is extracted to make the code testable.
func run(in io.Reader, out io.Writer) {
	// Create a buffered reader from input with 1MB buffer (1<<20 bytes)
	// This improves performance when reading large amounts of data by reducing
	// the number of system calls required
	bufIn := bufio.NewReaderSize(in, 1<<20)
	
	// Create a buffered writer to output with 1MB buffer (1<<20 bytes)
	// This batches write operations to improve performance by reducing
	// the number of system calls required
	bufOut := bufio.NewWriterSize(out, 1<<20)
	
	// Ensure all buffered data is written before returning
	// This is critical to prevent data loss
	defer bufOut.Flush()

	// Create a new predictor instance that will handle the guessing logic
	p := internal.NewPredictor()
	
	// Execute the predictor's main loop, passing the buffered input and output streams
	// The predictor will read data from 'bufIn', process it, and write results to 'bufOut'
	p.Run(bufIn, bufOut)
}

// main is the entry point of the application.
// It sets up buffered I/O streams for efficient data processing and
// initializes the predictor to handle the guessing logic.
func main() {
	run(os.Stdin, os.Stdout)
}
