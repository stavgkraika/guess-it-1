package internal

import (
	"bufio"
	"fmt"
	"math"
)

// Predictor encapsulates state for streaming prediction intervals.
// It maintains a rolling window of recent values, tracks differences between
// consecutive values, monitors hit rates, and dynamically adjusts the prediction
// interval width based on performance feedback.
type Predictor struct {
	// win stores the rolling window of recent observed values
	win Ring
	
	// diffs stores absolute differences between consecutive values
	diffs DiffRing
	
	// hits tracks whether recent predictions were successful (1) or not (0)
	hits HitRing

	// prevY is the most recently observed value
	prevY int64
	
	// hasPrevY indicates whether we have seen at least one value
	hasPrevY bool
	
	// prevL and prevU are the lower and upper bounds of the previous prediction interval
	prevL, prevU int64
	
	// hasPrevIV indicates whether we have made at least one prediction interval
	hasPrevIV bool

	// k is the dynamic multiplier for the interval width, adjusted based on hit rate
	k float64
	
	// kMin and kMax define the allowed range for the k multiplier
	kMin, kMax float64

	// tmpVals is a reusable slice for extracting window values (avoids allocations)
	tmpVals []int64
	
	// tmpDiffs is a reusable slice for extracting diff values (avoids allocations)
	tmpDiffs []int64
}

// NewPredictor creates and initializes a new Predictor instance.
// It sets initial values for the k multiplier and its bounds, and
// pre-allocates temporary slices to avoid repeated allocations during processing.
func NewPredictor() *Predictor {
	return &Predictor{
		k:        1.30, // Initial multiplier for interval width
		kMin:     1.05, // Minimum allowed multiplier (tighter intervals)
		kMax:     2.00, // Maximum allowed multiplier (wider intervals)
		tmpVals:  make([]int64, 0, WindowN), // Pre-allocate for window values
		tmpDiffs: make([]int64, 0, WindowN), // Pre-allocate for diff values
	}
}

// Run is the main processing loop that reads values from input, generates
// prediction intervals, and writes them to output. It continues until EOF.
func (p *Predictor) Run(in *bufio.Reader, out *bufio.Writer) {
	for {
		// Read the next integer value from input
		var y int64
		_, err := fmt.Fscan(in, &y)
		if err != nil {
			// EOF or read error, terminate the loop
			return
		}

		// Evaluate previous interval against current y
		// This feedback loop allows us to adjust the k multiplier dynamically
		if p.hasPrevIV {
			// Check if the current value falls within the previous prediction interval
			hit := 0
			if y >= p.prevL && y <= p.prevU {
				hit = 1 // Prediction was successful
			}
			p.hits.Push(hit)

			// Calculate the hit rate over recent predictions
			r := p.hits.Rate()
			
			// If hit rate is too low (<80%), widen intervals by increasing k
			if r < 0.80 {
				p.k = math.Min(p.k+0.10, p.kMax)
			} else if r > 0.92 {
				// If hit rate is too high (>92%), narrow intervals by decreasing k
				// This allows us to be more precise without sacrificing coverage
				p.k = math.Max(p.k-0.03, p.kMin)
			}
		}

		// Update diffs and window with the new value
		if p.hasPrevY {
			// Store the absolute difference between consecutive values
			// This helps us understand the typical step size in the data
			p.diffs.Push(abs64(y - p.prevY))
		}
		
		// Add the current value to the rolling window
		p.win.Push(y)

		// Build time-ordered slice from the ring buffer
		// This gives us the values in chronological order for statistical analysis
		p.tmpVals = p.win.ToSlice(p.tmpVals)

		// Startup: very early wide interval
		// When we have insufficient data, use a conservative wide interval
		if len(p.tmpVals) < 3 {
			// Start with a base width of 500
			w := int64(500)
			ay := abs64(y)
			
			// If the current value is large, scale the width proportionally
			if ay > 0 {
				prop := ay / 2
				if prop > w {
					w = prop
				}
			}
			
			// Create symmetric interval around current value
			L := y - w
			U := y + w
			fmt.Fprintf(out, "%d %d\n", L, U)

			// Store this interval and value for next iteration
			p.prevL, p.prevU, p.hasPrevIV = L, U, true
			p.prevY, p.hasPrevY = y, true
			continue
		}

		// Stats: compute mean, standard deviation, and robust statistics
		// mu is the mean of recent values (center of distribution)
		// sigma is the standard deviation (measure of spread)
		mu, sigma := MeanStd(p.tmpVals)
		
		// MAD (Median Absolute Deviation) is a robust measure of spread
		// that is less sensitive to outliers than standard deviation
		_, mad := MedianMAD(p.tmpVals)
		
		// Convert MAD to a scale comparable to standard deviation
		// The factor 1.4826 makes MAD consistent with std dev for normal distributions
		sigmaRobust := 1.4826 * mad
		
		// Use the larger of the two spread measures for robustness
		sigmaStar := math.Max(sigma, sigmaRobust)

		// Typical step s: the median of recent absolute differences
		// This tells us how much the values typically change between observations
		p.tmpDiffs = p.diffs.Values(p.tmpDiffs)
		s := TypicalStep(p.tmpDiffs)
		if s <= 0 {
			// Fallback if we don't have enough diff data
			s = math.Max(1.0, sigmaStar)
		}

		// Trend d = y_t - y_{t-1}
		// This captures the direction and magnitude of the most recent change
		var d int64
		if p.hasPrevY {
			d = y - p.prevY
		}

		// Forecast: trend + mean reversion
		// beta controls how much we trust the recent trend to continue
		// gamma controls how much we pull the forecast back toward the mean
		beta := 0.60
		gamma := 0.20
		
		// Early on, disable mean reversion and use wider intervals
		if len(p.tmpVals) < 10 {
			gamma = 0.0 // Don't pull toward mean when we have little data
			if p.k < 1.40 {
				p.k = 1.40 // Ensure wider intervals during startup
			}
		}
		
		// Forecast formula: current value + trend component + mean reversion component
		yhat := float64(y) + beta*float64(d) + gamma*(mu-float64(y))

		// Width: base width on spread and the dynamic k multiplier
		b := 2.0 // Base offset to ensure minimum width
		W := p.k*sigmaStar + b

		// Clamp width to reasonable bounds based on typical step size
		Wmin := math.Max(3.0, 1.5*s)  // At least 1.5x the typical step
		Wmax := math.Max(20.0, 10.0*s) // At most 10x the typical step

		// During startup, enforce a higher minimum width for safety
		if len(p.tmpVals) < 10 && Wmin < 10.0 {
			Wmin = 10.0
		}

		// Apply the width bounds
		W = clampFloat(W, Wmin, Wmax)

		// Compute final interval bounds
		// Floor and Ceil ensure we get integer bounds that fully cover the forecast
		L := int64(math.Floor(yhat - W))
		U := int64(math.Ceil(yhat + W))
		
		// Ensure L <= U (should always be true, but defensive check)
		if L > U {
			L, U = U, L
		}

		// Write the prediction interval to output
		fmt.Fprintf(out, "%d %d\n", L, U)

		// Store this interval and value for the next iteration
		p.prevL, p.prevU, p.hasPrevIV = L, U, true
		p.prevY, p.hasPrevY = y, true
	}
}
