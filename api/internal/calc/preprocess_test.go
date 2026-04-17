package calc

import "testing"

func TestRewritePercent(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no percent", "1+2", "1+2"},
		{"simple numbers", "50 % 100", "(((50)/100)*(100))"},
		{"no spaces", "50%100", "(((50)/100)*(100))"},
		{"decimal operands", "12.5%80", "(((12.5)/100)*(80))"},
		{"left paren group", "(1+1)%100", "((((1+1))/100)*(100))"},
		{"right paren group", "50%(2*50)", "(((50)/100)*((2*50)))"},
		{"both paren groups", "(10+40)%(2+98)", "((((10+40))/100)*((2+98)))"},
		{"function left", "sqrt(25)%100", "(((sqrt(25))/100)*(100))"},
		{"function right", "10%sqrt(100)", "(((10)/100)*(sqrt(100)))"},
		{"nested parens", "((1+1))%((50))", "(((((1+1)))/100)*(((50))))"},
		{"multiple percents", "10%100+20%50", "(((10)/100)*(100))+(((20)/100)*(50))"},
		// Chained percent: the algorithm rewrites left-to-right. Result is
		// still numerically correct because the operators are commutative in
		// the composition; we assert shape so regressions stay visible.
		{"chained percent", "10%20%100", "((((((10)/100)*(20)))/100)*(100))"},
		{"trailing percent left alone", "50%", "50%"},
		{"leading percent left alone", "%100", "%100"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RewritePercent(tc.in)
			if got != tc.want {
				t.Errorf("RewritePercent(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPreprocess_CaretToPower(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"2^3", "2**3"},
		{"2^3+4^5", "2**3+4**5"},
		{"(1+1)^2", "(1+1)**2"},
		{"no caret", "no caret"},
	}
	for _, tc := range tests {
		if got := Preprocess(tc.in); got != tc.want {
			t.Errorf("Preprocess(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHasUnbalancedParens(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"1+1", false},
		{"(1+1)", false},
		{"((1+1))", false},
		{"sqrt(16)", false},
		{"", false},
		{"(1+1", true},       // missing close
		{"1+1)", true},       // extra close
		{"((1+1)", true},     // one unclosed
		{"(1+1))", true},     // one extra close
		{"sqrt(16", true},    // function call with no close
		{"()", false},        // syntactically bogus but balanced — falls through to expr
		{")(", false},        // balanced but out of order; caught later by expr
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := HasUnbalancedParens(tc.in); got != tc.want {
				t.Errorf("HasUnbalancedParens(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPreprocess_Combined(t *testing.T) {
	// Caret is rewritten first, then percent-of operates on the rewritten
	// expression. The outer parens on the percent rewrite stop `**` from
	// binding to the percent-of product.
	got := Preprocess("2^3 % 100")
	want := "(((3)/100)*(100))"
	// The left operand of `%` is only `3` (the regex operand is number-or-parens,
	// not exponentiation). Everything before stays as-is.
	wantFull := "2**" + want
	if got != wantFull {
		t.Errorf("Preprocess(%q) = %q, want %q", "2^3 % 100", got, wantFull)
	}
}
