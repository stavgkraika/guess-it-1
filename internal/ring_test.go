package internal

import (
	"testing"
)

// TestRingPush tests the Push method of Ring
func TestRingPush(t *testing.T) {
	r := &Ring{}

	// Test pushing values
	for i := int64(1); i <= 10; i++ {
		r.Push(i)
		if r.Count() != int(i) {
			t.Errorf("Expected count %d, got %d", i, r.Count())
		}
	}

	// Test that count stops at WindowN
	for i := int64(11); i <= WindowN+10; i++ {
		r.Push(i)
	}
	if r.Count() != WindowN {
		t.Errorf("Expected count %d, got %d", WindowN, r.Count())
	}
}

// TestRingToSlice tests the ToSlice method
func TestRingToSlice(t *testing.T) {
	r := &Ring{}
	dst := make([]int64, 0, WindowN)

	// Test empty ring
	result := r.ToSlice(dst)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}

	// Test with some values
	for i := int64(1); i <= 5; i++ {
		r.Push(i)
	}
	result = r.ToSlice(dst)
	if len(result) != 5 {
		t.Errorf("Expected length 5, got %d", len(result))
	}
	for i := 0; i < 5; i++ {
		if result[i] != int64(i+1) {
			t.Errorf("Expected value %d at index %d, got %d", i+1, i, result[i])
		}
	}

	// Test wraparound by filling beyond WindowN
	r2 := &Ring{}
	for i := int64(1); i <= WindowN+5; i++ {
		r2.Push(i)
	}
	result = r2.ToSlice(dst)
	if len(result) != WindowN {
		t.Errorf("Expected length %d, got %d", WindowN, len(result))
	}
	// Should contain values from 6 to WindowN+5
	for i := 0; i < WindowN; i++ {
		expected := int64(i + 6)
		if result[i] != expected {
			t.Errorf("Expected value %d at index %d, got %d", expected, i, result[i])
		}
	}
}

// TestDiffRingPush tests the Push method of DiffRing
func TestDiffRingPush(t *testing.T) {
	d := &DiffRing{}

	// Test pushing values
	for i := int64(1); i <= 10; i++ {
		d.Push(i)
		if d.count != int(i) {
			t.Errorf("Expected count %d, got %d", i, d.count)
		}
	}

	// Test that count stops at WindowN
	for i := int64(11); i <= WindowN+10; i++ {
		d.Push(i)
	}
	if d.count != WindowN {
		t.Errorf("Expected count %d, got %d", WindowN, d.count)
	}
}

// TestDiffRingValues tests the Values method
func TestDiffRingValues(t *testing.T) {
	d := &DiffRing{}
	dst := make([]int64, 0, WindowN)

	// Test empty ring
	result := d.Values(dst)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}

	// Test with some values
	for i := int64(10); i <= 50; i += 10 {
		d.Push(i)
	}
	result = d.Values(dst)
	if len(result) != 5 {
		t.Errorf("Expected length 5, got %d", len(result))
	}
	expected := []int64{10, 20, 30, 40, 50}
	for i := 0; i < 5; i++ {
		if result[i] != expected[i] {
			t.Errorf("Expected value %d at index %d, got %d", expected[i], i, result[i])
		}
	}
}

// TestHitRingPush tests the Push method of HitRing
func TestHitRingPush(t *testing.T) {
	h := &HitRing{}

	// Test pushing hits
	for i := 0; i < 10; i++ {
		h.Push(1)
		if h.count != i+1 {
			t.Errorf("Expected count %d, got %d", i+1, h.count)
		}
		if h.sum != i+1 {
			t.Errorf("Expected sum %d, got %d", i+1, h.sum)
		}
	}

	// Test pushing misses
	h2 := &HitRing{}
	for i := 0; i < 5; i++ {
		h2.Push(0)
	}
	if h2.sum != 0 {
		t.Errorf("Expected sum 0, got %d", h2.sum)
	}

	// Test wraparound
	h3 := &HitRing{}
	for i := 0; i < HitM; i++ {
		h3.Push(1)
	}
	// Now push a 0, should replace oldest 1
	h3.Push(0)
	if h3.sum != HitM-1 {
		t.Errorf("Expected sum %d, got %d", HitM-1, h3.sum)
	}
	if h3.count != HitM {
		t.Errorf("Expected count %d, got %d", HitM, h3.count)
	}
}

// TestHitRingRate tests the Rate method
func TestHitRingRate(t *testing.T) {
	h := &HitRing{}

	// Test empty ring
	if h.Rate() != 0 {
		t.Errorf("Expected rate 0 for empty ring, got %f", h.Rate())
	}

	// Test 100% hit rate
	for i := 0; i < 10; i++ {
		h.Push(1)
	}
	if h.Rate() != 1.0 {
		t.Errorf("Expected rate 1.0, got %f", h.Rate())
	}

	// Test 50% hit rate
	h2 := &HitRing{}
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			h2.Push(1)
		} else {
			h2.Push(0)
		}
	}
	if h2.Rate() != 0.5 {
		t.Errorf("Expected rate 0.5, got %f", h2.Rate())
	}

	// Test 0% hit rate
	h3 := &HitRing{}
	for i := 0; i < 10; i++ {
		h3.Push(0)
	}
	if h3.Rate() != 0.0 {
		t.Errorf("Expected rate 0.0, got %f", h3.Rate())
	}
}

// TestAbs64 tests the abs64 helper function
func TestAbs64(t *testing.T) {
	tests := []struct {
		input    int64
		expected int64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{100, 100},
		{-100, 100},
	}

	for _, test := range tests {
		result := abs64(test.input)
		if result != test.expected {
			t.Errorf("abs64(%d) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

// TestClampFloat tests the clampFloat helper function
func TestClampFloat(t *testing.T) {
	tests := []struct {
		x        float64
		lo       float64
		hi       float64
		expected float64
	}{
		{5.0, 0.0, 10.0, 5.0},
		{-5.0, 0.0, 10.0, 0.0},
		{15.0, 0.0, 10.0, 10.0},
		{7.5, 5.0, 10.0, 7.5},
		{3.0, 5.0, 10.0, 5.0},
		{12.0, 5.0, 10.0, 10.0},
	}

	for _, test := range tests {
		result := clampFloat(test.x, test.lo, test.hi)
		if result != test.expected {
			t.Errorf("clampFloat(%f, %f, %f) = %f, expected %f",
				test.x, test.lo, test.hi, result, test.expected)
		}
	}
}

// TestRingCount tests the Count method
func TestRingCount(t *testing.T) {
	r := &Ring{}

	if r.Count() != 0 {
		t.Errorf("Expected count 0 for new ring, got %d", r.Count())
	}

	r.Push(1)
	if r.Count() != 1 {
		t.Errorf("Expected count 1, got %d", r.Count())
	}

	for i := 2; i <= 100; i++ {
		r.Push(int64(i))
	}
	if r.Count() != 100 {
		t.Errorf("Expected count 100, got %d", r.Count())
	}
}
