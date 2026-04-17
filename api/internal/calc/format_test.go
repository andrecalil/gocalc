package calc

import (
	"math"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want string
	}{
		{"zero", 0, "0"},
		{"integer", 8, "8"},
		{"negative integer", -42, "-42"},
		{"large integer", 1234567890, "1234567890"},
		{"simple float", 0.5, "0.5"},
		{"one third", 1.0 / 3.0, "0.3333333333"},
		{"two thirds", 2.0 / 3.0, "0.6666666667"},
		{"trim trailing zeros", 1.5, "1.5"},
		{"pi-ish", 3.14159265358979, "3.141592654"},
		{"integer-valued float", 8.0, "8"},
		{"near-integer stays integer when truly whole", 100.0 * 0.01, "1"},
		{"very small → scientific", 1e-9, "1e-09"},
		{"very large → scientific", 1e20, "1e+20"},
		{"one millionth boundary not scientific", 1e-6, "0.000001"},
		{"just below boundary scientific", 1e-7, "1e-07"},
		{"defensive NaN", math.NaN(), "NaN"},
		{"defensive +Inf", math.Inf(1), "+Inf"},
		{"defensive -Inf", math.Inf(-1), "-Inf"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Format(tc.in)
			if got != tc.want {
				t.Errorf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
