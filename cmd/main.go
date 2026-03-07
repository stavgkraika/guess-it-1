// Package main is the entry point for the guess-it-1 application.
// This program implements a prediction system that reads input data,
// processes it through a predictor, and writes the results to output.
package main

import (
	"bufio" // Provides buffered I/O operations for efficient reading and writing
	"os"    // Provides platform-independent interface to operating system functionality

	"guess-it-1/internal" // Internal package containing the predictor implementation
)

// main is the entry point of the application.
// It sets up buffered I/O streams for efficient data processing and
// initializes the predictor to handle the guessing logic.
func main() {
	// Create a buffered reader from standard input with 1MB buffer (1<<20 bytes)
	// This improves performance when reading large amounts of data by reducing
	// the number of system calls required
	in := bufio.NewReaderSize(os.Stdin, 1<<20)
	
	// Create a buffered writer to standard output with 1MB buffer (1<<20 bytes)
	// This batches write operations to improve performance by reducing
	// the number of system calls required
	out := bufio.NewWriterSize(os.Stdout, 1<<20)
	
	// Ensure all buffered data is written to stdout before the program exits
	// This is critical to prevent data loss when the program terminates
	defer out.Flush()

	// Create a new predictor instance that will handle the guessing logic
	p := internal.NewPredictor()
	
	// Execute the predictor's main loop, passing the buffered input and output streams
	// The predictor will read data from 'in', process it, and write results to 'out'
	p.Run(in, out)
}
