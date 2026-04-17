package calc

import (
	"math"
	"strconv"
	"strings"
)

// Format renders a float64 result as a user-facing string.
//
// Rules (from the project spec):
//   - Integer-valued results: no decimal point ("8" not "8.0").
//   - Float results: up to 10 significant digits, trailing zeros trimmed.
//   - Magnitudes outside [1e-6, 1e15]: scientific notation is acceptable.
//
// The goal is "what a calculator display would show" — never "8.0000000001"
// for a result that is mathematically 8, and never unnecessary trailing zeros
// for things like 1/2.
func Format(v float64) string {
	// Defensive — callers should already have translated NaN/Inf into a
	// division-by-zero error, but if a NaN/Inf slips through, produce a
	// stable string rather than panicking in strconv.
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 0) {
		if v > 0 {
			return "+Inf"
		}
		return "-Inf"
	}

	if v == 0 {
		return "0"
	}

	abs := math.Abs(v)

	// Outside the "reasonable" range — use scientific notation. Precision 10
	// gives up to 10 significant digits; 'g' format trims trailing zeros.
	if abs < 1e-6 || abs >= 1e15 {
		return strconv.FormatFloat(v, 'g', 10, 64)
	}

	// Integer-valued within range — no decimal point.
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}

	// Non-integer within range — render fixed with enough decimals to reach
	// 10 significant digits, then trim trailing zeros.
	//
	// exp is floor(log10(abs)), i.e. the zero-indexed position of the most
	// significant digit. We want (10 - 1 - exp) digits after the decimal
	// point so the total count of significant digits is 10.
	exp := int(math.Floor(math.Log10(abs)))
	decimals := 10 - 1 - exp
	if decimals < 0 {
		decimals = 0
	}
	if decimals > 15 {
		decimals = 15
	}
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
