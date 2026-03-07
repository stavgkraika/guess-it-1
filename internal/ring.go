package internal

import "math"

const (
	// WindowN is the size of the rolling window for values and diffs
	// This determines how much history we keep for statistical analysis
	WindowN = 50
	
	// HitM is the size of the hit-rate history
	// This determines how many recent predictions we use to calculate hit rate
	HitM = 20
)

// Ring is a fixed-size ring buffer for int64 values.
// It efficiently stores the most recent WindowN values in a circular buffer,
// automatically overwriting the oldest value when full.
type Ring struct {
	// buf is the underlying fixed-size array storing the values
	buf [WindowN]int64
	
	// count tracks how many values have been added (up to WindowN)
	count int
	
	// idx is the position where the next value will be written
	idx int
}

// Push adds a new value to the ring buffer.
// If the buffer is full, it overwrites the oldest value.
func (r *Ring) Push(x int64) {
	// Write the value at the current index position
	r.buf[r.idx] = x
	
	// Move to the next position, wrapping around if necessary
	r.idx = (r.idx + 1) % WindowN
	
	// Increment count until we reach capacity
	if r.count < WindowN {
		r.count++
	}
}

// Count returns the number of values currently stored in the ring buffer.
// This will be less than WindowN during the initial fill phase.
func (r *Ring) Count() int { return r.count }

// ToSlice returns window values in time order (oldest -> newest) into dst.
// It reuses the provided slice to avoid allocations, resizing it as needed.
// The returned slice contains values in chronological order, which is important
// for time-series analysis and statistical calculations.
func (r *Ring) ToSlice(dst []int64) []int64 {
	// Resize the destination slice to match the number of stored values
	dst = dst[:r.count]
	
	if r.count == 0 {
		return dst
	}
	
	// Calculate the starting position (oldest value)
	// This is where we started overwriting, or 0 if not yet full
	start := r.idx - r.count
	for start < 0 {
		start += WindowN // Handle negative wrap-around
	}
	
	// Copy values in chronological order
	for i := 0; i < r.count; i++ {
		dst[i] = r.buf[(start+i)%WindowN]
	}
	
	return dst
}

// DiffRing stores recent absolute diffs.
// It's structurally identical to Ring but semantically represents
// differences between consecutive values rather than the values themselves.
type DiffRing struct {
	// buf is the underlying fixed-size array storing the diff values
	buf [WindowN]int64
	
	// count tracks how many diffs have been added (up to WindowN)
	count int
	
	// idx is the position where the next diff will be written
	idx int
}

// Push adds a new diff value to the ring buffer.
// If the buffer is full, it overwrites the oldest diff.
func (d *DiffRing) Push(x int64) {
	// Write the diff at the current index position
	d.buf[d.idx] = x
	
	// Move to the next position, wrapping around if necessary
	d.idx = (d.idx + 1) % WindowN
	
	// Increment count until we reach capacity
	if d.count < WindowN {
		d.count++
	}
}

// Values returns all stored diffs in time order (oldest -> newest) into dst.
// It reuses the provided slice to avoid allocations.
func (d *DiffRing) Values(dst []int64) []int64 {
	// Resize the destination slice to match the number of stored diffs
	dst = dst[:d.count]
	
	if d.count == 0 {
		return dst
	}
	
	// Calculate the starting position (oldest diff)
	start := d.idx - d.count
	for start < 0 {
		start += WindowN // Handle negative wrap-around
	}
	
	// Copy diffs in chronological order
	for i := 0; i < d.count; i++ {
		dst[i] = d.buf[(start+i)%WindowN]
	}
	
	return dst
}

// HitRing stores recent hit/miss outcomes.
// Each entry is 1 (hit) or 0 (miss), indicating whether a prediction
// interval successfully contained the actual value.
type HitRing struct {
	// buf is the underlying fixed-size array storing hit/miss indicators
	buf [HitM]int
	
	// count tracks how many outcomes have been recorded (up to HitM)
	count int
	
	// idx is the position where the next outcome will be written
	idx int
	
	// sum is the running total of hits, used to efficiently calculate hit rate
	sum int
}

// Push adds a new hit/miss outcome to the ring buffer.
// It maintains a running sum for efficient rate calculation.
func (h *HitRing) Push(hit int) {
	if h.count < HitM {
		// Buffer not yet full, just add the new value
		h.buf[h.idx] = hit
		h.sum += hit
		h.idx = (h.idx + 1) % HitM
		h.count++
		return
	}
	
	// Buffer is full, replace the oldest value
	old := h.buf[h.idx]
	h.sum -= old        // Remove old value from sum
	h.buf[h.idx] = hit  // Write new value
	h.sum += hit        // Add new value to sum
	h.idx = (h.idx + 1) % HitM
}

// Rate returns the hit rate as a fraction between 0 and 1.
// This represents the proportion of recent predictions that were successful.
func (h *HitRing) Rate() float64 {
	if h.count == 0 {
		return 0
	}
	return float64(h.sum) / float64(h.count)
}

// Helpers

// abs64 returns the absolute value of an int64.
// This is used to compute absolute differences between values.
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// clampFloat constrains a float64 value to the range [lo, hi].
// This ensures that computed widths stay within reasonable bounds.
func clampFloat(x, lo, hi float64) float64 {
	return math.Min(math.Max(x, lo), hi)
}
