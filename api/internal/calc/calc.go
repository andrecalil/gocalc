package calc

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Sentinel errors. The HTTP layer maps these to status codes and user-facing
// messages; keeping them as errors.Is-compatible sentinels avoids sharing
// HTTP-specific concerns with the calc package.
//
// Some errors carry a "recommended fix" — a friendlier message that suggests
// what the user should correct. These are specific sentinels (e.g.
// ErrUnclosedParenthesis) so the HTTP layer can map them to tailored
// messages instead of the generic "invalid expression". Adding another
// recommended-fix case is a matter of (1) adding a new sentinel here,
// (2) adding a detector in preprocess/validate, and (3) mapping it in
// httpapi/errors.go.
var (
	ErrEmptyExpression     = errors.New("expression is empty")
	ErrInvalidExpression   = errors.New("invalid expression")
	ErrUnclosedParenthesis = errors.New("unclosed parenthesis")
	ErrDivisionByZero      = errors.New("division by zero")
	ErrNonNumericResult    = errors.New("expression did not produce a numeric result")
)

// Calculate parses and evaluates a calculator expression and returns the
// result formatted for display.
//
// Pipeline:
//  1. Trim + reject empty input.
//  2. Preprocess (`^` → `**`; percent-of rewrite).
//  3. Force integer literals to floats so `a/0` returns +Inf instead of
//     panicking — this lets us detect division-by-zero uniformly via
//     IsInf/IsNaN on the result.
//  4. Compile + run via expr-lang/expr with `sqrt` exposed as a function.
//  5. Post-check for NaN/Inf to surface division-by-zero.
//  6. Format the result.
func Calculate(expression string) (string, error) {
	trimmed := strings.TrimSpace(expression)
	if trimmed == "" {
		return "", ErrEmptyExpression
	}

	// Recommended-fix detectors run BEFORE preprocessing so the diagnosis
	// reflects what the user actually typed. The percent-of rewrite adds
	// its own parens, which would otherwise hide a mismatch.
	if HasUnbalancedParens(trimmed) {
		return "", ErrUnclosedParenthesis
	}

	rewritten := Preprocess(trimmed)
	rewritten = floatifyIntegerLiterals(rewritten)

	env := map[string]any{
		"sqrt": math.Sqrt,
	}

	var program *vm.Program
	program, err := expr.Compile(rewritten, expr.Env(env), expr.AsFloat64())
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}

	out, err := expr.Run(program, env)
	if err != nil {
		// Runtime errors in expr are typically "runtime error: integer
		// divide by zero" (shouldn't happen now that we floatify literals)
		// or type-mismatch errors. Surface them as "invalid expression".
		return "", fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}

	val, ok := out.(float64)
	if !ok {
		return "", ErrNonNumericResult
	}

	if math.IsNaN(val) {
		// 0.0/0.0 → NaN; sqrt(-1) → NaN. Treat both as division-by-zero
		// from the user's perspective — a mathematically undefined result.
		return "", ErrDivisionByZero
	}
	if math.IsInf(val, 0) {
		return "", ErrDivisionByZero
	}

	return Format(val), nil
}

// floatifyIntegerLiterals rewrites integer literals as floats ("50" → "50.0")
// so that every operation in the expression happens in float64. This prevents
// `expr` from panicking on integer division by zero and gives us uniform
// IEEE-754 semantics (a/0 → +Inf, detectable via math.IsInf).
//
// The scanner is careful to:
//   - skip digits that follow an identifier character (e.g., `x2` is an
//     identifier, not a number);
//   - leave existing decimal literals untouched.
func floatifyIntegerLiterals(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	i := 0
	for i < len(s) {
		c := s[i]

		// Skip over identifiers entirely so digits inside them (like `x2`)
		// aren't treated as numeric literals.
		if isIdentStart(c) {
			j := i
			for j < len(s) && isIdentChar(s[j]) {
				j++
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}

		if isDigit(c) {
			j := i
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			// If a `.` follows and is itself followed by a digit, it's already
			// a float literal — keep scanning through the fractional part.
			if j < len(s) && s[j] == '.' && j+1 < len(s) && isDigit(s[j+1]) {
				j++
				for j < len(s) && isDigit(s[j]) {
					j++
				}
				b.WriteString(s[i:j])
			} else {
				// Bare integer — append ".0".
				b.WriteString(s[i:j])
				b.WriteString(".0")
			}
			i = j
			continue
		}

		b.WriteByte(c)
		i++
	}
	return b.String()
}
