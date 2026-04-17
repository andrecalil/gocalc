// Package calc performs expression preprocessing, evaluation, and result
// formatting. This file handles the two preprocessing steps we need before
// handing the expression to expr-lang/expr:
//
//  1. Rewrite `^` as `**` (expr-lang/expr uses `**` for exponentiation; `^`
//     is bitwise XOR, which is not what a calculator user expects).
//  2. Rewrite the binary percent-of operator. In our calculator `a % b` means
//     "a percent of b", i.e. `((a)/100)*(b)`. In expr, `%` is modulo, so we
//     must rewrite before evaluation.
//
// The percent-of rewrite supports three operand shapes on either side of `%`:
//
//   - a numeric literal (integer or decimal)
//   - a parenthesized sub-expression with balanced parens
//   - a function call like `sqrt(x)` (identifier + balanced parens)
//
// Anything else (e.g., a bare `%` with a missing operand) is left as-is; the
// downstream `expr` parser will reject it and the user will see
// "invalid expression".
package calc

import (
	"regexp"
	"strings"
)

// Precompiled regexes for the parenthesis detector. Using a regex here
// (rather than a plain byte loop) keeps the "find all of these characters"
// intent explicit and makes it trivial to add more shape checks later.
var (
	openParenRegexp  = regexp.MustCompile(`\(`)
	closeParenRegexp = regexp.MustCompile(`\)`)
)

// HasUnbalancedParens reports whether the expression has a different number
// of `(` and `)` characters — either a missing `)` (more opens than closes)
// or a stray `)` (more closes than opens). Either way the user has a
// parenthesis mismatch and we surface a tailored recommended-fix message.
//
// Note: this only catches count mismatches, not ordering errors like
// `)1+1(`. Ordering errors still reach `expr` and come back as the generic
// "invalid expression". That's an acceptable tradeoff — count mismatches
// are the common typo and this check is O(n) and allocation-free.
func HasUnbalancedParens(expression string) bool {
	opens := len(openParenRegexp.FindAllStringIndex(expression, -1))
	closes := len(closeParenRegexp.FindAllStringIndex(expression, -1))
	return opens != closes
}

// Preprocess runs the full preprocessing pipeline on the raw expression.
func Preprocess(expression string) string {
	s := rewriteCaretAsPower(expression)
	s = RewritePercent(s)
	return s
}

// rewriteCaretAsPower replaces every `^` with `**`. `^` does not appear in any
// other role in the expressions we accept, so a blind substitution is safe.
func rewriteCaretAsPower(s string) string {
	return strings.ReplaceAll(s, "^", "**")
}

// RewritePercent rewrites every binary percent-of occurrence in s.
// `a % b` becomes `((a)/100)*(b)`. The function iterates left-to-right,
// rewriting one occurrence per pass; the loop terminates when there are no
// more rewritable `%` operators (either no `%` remains, or remaining ones
// don't have both operands and can't be rewritten).
func RewritePercent(s string) string {
	// We track a starting index so that, once we've decided a particular `%`
	// isn't rewritable, we skip past it instead of spinning forever.
	searchFrom := 0
	for {
		idx := strings.IndexByte(s[searchFrom:], '%')
		if idx < 0 {
			return s
		}
		idx += searchFrom

		lStart, lEnd, okL := findLeftOperand(s, idx)
		rStart, rEnd, okR := findRightOperand(s, idx)
		if !okL || !okR {
			// Move past this `%` and keep looking — maybe a later one is valid.
			searchFrom = idx + 1
			continue
		}

		left := s[lStart:lEnd]
		right := s[rStart:rEnd]
		// Outer parens around the whole rewrite are essential: `%` binds
		// tighter than `*`/`/`, so in `2**3 % 100` the left operand is
		// just `3`. Without the outer parens the rewritten `*` binds to
		// the `**` in surprising ways. The outer parens confine the
		// rewrite to a single sub-expression with `%`'s original precedence.
		replacement := "(((" + left + ")/100)*(" + right + "))"
		s = s[:lStart] + replacement + s[rEnd:]
		// Reset the search so nested rewrites can be picked up.
		searchFrom = 0
	}
}

// findLeftOperand scans backward from pos to find the operand immediately to
// the left of the `%` at s[pos]. Returns [start, end) and ok=true on success.
func findLeftOperand(s string, pos int) (int, int, bool) {
	end := pos
	for end > 0 && isSpace(s[end-1]) {
		end--
	}
	if end == 0 {
		return 0, 0, false
	}

	last := s[end-1]

	// Parenthesized group: scan back for matching `(`, and include any
	// preceding identifier so function calls like `sqrt(x)` stay intact.
	if last == ')' {
		openIdx, ok := matchOpenParen(s, end-1)
		if !ok {
			return 0, 0, false
		}
		start := openIdx
		for start > 0 && isIdentChar(s[start-1]) {
			start--
		}
		return start, end, true
	}

	// Numeric literal: digits and at most one decimal point.
	if isDigit(last) || last == '.' {
		start := end - 1
		for start > 0 && (isDigit(s[start-1]) || s[start-1] == '.') {
			start--
		}
		return start, end, true
	}

	return 0, 0, false
}

// findRightOperand scans forward from pos to find the operand immediately to
// the right of the `%` at s[pos].
func findRightOperand(s string, pos int) (int, int, bool) {
	start := pos + 1
	for start < len(s) && isSpace(s[start]) {
		start++
	}
	if start >= len(s) {
		return 0, 0, false
	}

	first := s[start]

	// Identifier followed by `(` — a function call.
	if isIdentStart(first) {
		i := start
		for i < len(s) && isIdentChar(s[i]) {
			i++
		}
		if i < len(s) && s[i] == '(' {
			closeIdx, ok := matchCloseParen(s, i)
			if !ok {
				return 0, 0, false
			}
			return start, closeIdx + 1, true
		}
		// bare identifier isn't a valid percent operand in our grammar
		return 0, 0, false
	}

	// Parenthesized group.
	if first == '(' {
		closeIdx, ok := matchCloseParen(s, start)
		if !ok {
			return 0, 0, false
		}
		return start, closeIdx + 1, true
	}

	// Numeric literal.
	if isDigit(first) || first == '.' {
		i := start
		for i < len(s) && (isDigit(s[i]) || s[i] == '.') {
			i++
		}
		return start, i, true
	}

	return 0, 0, false
}

// matchOpenParen finds the `(` that matches the `)` at closeIdx.
func matchOpenParen(s string, closeIdx int) (int, bool) {
	depth := 1
	for i := closeIdx - 1; i >= 0; i-- {
		switch s[i] {
		case ')':
			depth++
		case '(':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

// matchCloseParen finds the `)` that matches the `(` at openIdx.
func matchCloseParen(s string, openIdx int) (int, bool) {
	depth := 1
	for i := openIdx + 1; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }
func isDigit(b byte) bool { return b >= '0' && b <= '9' }
func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}
func isIdentChar(b byte) bool { return isIdentStart(b) || isDigit(b) }
