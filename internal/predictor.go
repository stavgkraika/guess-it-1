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

	// recentTrend tracks the exponential moving average of recent changes
	recentTrend float64

	// volatility tracks the recent volatility for adaptive width adjustment
	volatility float64
}

// NewPredictor creates and initializes a new Predictor instance.
// It sets initial values for the k multiplier and its bounds, and
// pre-allocates temporary slices to avoid repeated allocations during processing.
func NewPredictor() *Predictor {
	return &Predictor{
		k:        1.35,                      // Lower initial multiplier for tighter intervals
		kMin:     1.10,                      // Tighter minimum for precise predictions
		kMax:     2.20,                      // Lower maximum to keep intervals smaller
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

			// Aggressive thresholds for tight but accurate intervals
			// Target hit rate: 90-94%
			if r < 0.90 {
				// Hit rate too low, widen intervals
				p.k = math.Min(p.k+0.10, p.kMax)
			} else if r > 0.94 {
				// Hit rate too high, narrow intervals aggressively
				p.k = math.Max(p.k-0.025, p.kMin)
			}
		}

		// Update diffs and window with the new value
		if p.hasPrevY {
			// Store the absolute difference between consecutive values
			// This helps us understand the typical step size in the data
			diff := abs64(y - p.prevY)
			p.diffs.Push(diff)

			// Update exponential moving average of trend (alpha = 0.35 for faster response)
			change := float64(y - p.prevY)
			p.recentTrend = 0.35*change + 0.65*p.recentTrend

			// Update volatility estimate using exponential moving average
			p.volatility = 0.25*float64(diff) + 0.75*p.volatility
		}

		// Add the current value to the rolling window
		p.win.Push(y)

		// Build time-ordered slice from the ring buffer
		// This gives us the values in chronological order for statistical analysis
		p.tmpVals = p.win.ToSlice(p.tmpVals)

		// Startup: early wide interval (but smaller than before)
		// When we have insufficient data, use a conservative interval
		if len(p.tmpVals) < 8 {
			// Start with a base width of 600 (reduced from 1000)
			w := int64(600)
			ay := abs64(y)

			// If the current value is large, scale the width proportionally
			if ay > 0 {
				prop := ay / 2 // 50% of absolute value (reduced from 60%)
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

		// Use weighted combination favoring the smaller spread for tighter intervals
		// Give more weight to the smaller value for precision
		sigmaStar := 0.4*math.Max(sigma, sigmaRobust) + 0.6*math.Min(sigma, sigmaRobust)

		// Incorporate recent volatility but with less weight for tighter bounds
		if p.volatility > 0 {
			sigmaStar = math.Max(sigmaStar, p.volatility*0.6)
		}

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

		// Forecast: enhanced trend + mean reversion with smoothing
		// beta controls how much we trust the recent trend to continue
		// gamma controls how much we pull the forecast back toward the mean
		// delta controls influence of the smoothed trend
		beta := 0.50  // Balanced trend following
		gamma := 0.40 // Strong mean reversion for stability
		delta := 0.25 // Smoothed trend component

		// Adjust parameters based on data availability and volatility
		if len(p.tmpVals) < 20 {
			// Early phase: be more conservative
			gamma = 0.20
			delta = 0.15
			if p.k < 1.50 {
				p.k = 1.50 // Ensure adequate intervals during early startup
			}
		} else if len(p.tmpVals) >= 50 {
			// Mature phase: use all components with better tuning
			if p.volatility > sigmaStar*1.5 {
				// High volatility: reduce trend following, increase mean reversion
				beta = 0.40
				gamma = 0.50
			}
		}

		// Forecast formula: current value + immediate trend + smoothed trend + mean reversion
		yhat := float64(y) + beta*float64(d) + delta*p.recentTrend + gamma*(mu-float64(y))

		// Width: base width on spread and the dynamic k multiplier
		b := 2.5 // Reduced base offset for tighter intervals
		W := p.k*sigmaStar + b

		// Add volatility-based adjustment with reduced weight
		if p.volatility > 0 {
			W += p.volatility * 0.3 // Reduced from 0.5
		}

		// Clamp width to tighter bounds based on typical step size
		Wmin := math.Max(4.0, 1.8*s)  // Tighter minimum
		Wmax := math.Max(25.0, 10.0*s) // Lower maximum

		// During startup, enforce a reasonable minimum width for safety
		if len(p.tmpVals) < 20 && Wmin < 12.0 {
			Wmin = 12.0 // Reduced from 20
		}

		// In mature phase with low volatility, allow even tighter bounds
		if len(p.tmpVals) >= 80 && p.volatility < sigmaStar*0.5 {
			Wmin = math.Max(3.0, 1.5*s) // Very tight for stable data
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
