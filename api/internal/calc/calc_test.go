package calc

import (
	"errors"
	"testing"
)

func TestCalculate_Success(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		result string
	}{
		{"addition", "1+1", "2"},
		{"subtraction", "10-3", "7"},
		{"multiplication", "6*7", "42"},
		{"division integer result", "10/2", "5"},
		{"division producing float", "1/3", "0.3333333333"},
		{"precedence", "1+2*3", "7"},
		{"parens override precedence", "(1+2)*3", "9"},
		{"nested parens", "((1+1)*4)", "8"},
		{"exponent", "2^10", "1024"},
		{"exponent float", "2^0.5", "1.414213562"},
		{"sqrt", "sqrt(16)", "4"},
		// sqrt(2)*sqrt(2) is 2.0000000000000004 in IEEE-754, but the formatter
		// rounds to 10 significant digits and trims trailing zeros → "2".
		{"sqrt composed", "sqrt(2)*sqrt(2)", "2"},
		{"percent-of basic", "50%100", "50"},
		{"percent-of reversed", "100%50", "50"},
		{"percent-of with spaces", "20 % 200", "40"},
		{"percent inside expr", "100+10%100", "110"},
		{"percent parens", "(10+40)%200", "100"},
		{"negative", "-5+3", "-2"},
		{"negative in parens", "(-5)*2", "-10"},
		{"whitespace", "  1 + 1  ", "2"},
		{"decimal literal", "0.1+0.2", "0.3"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Calculate(tc.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.result {
				t.Errorf("Calculate(%q) = %q, want %q", tc.expr, got, tc.result)
			}
		})
	}
}

func TestCalculate_Errors(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want error
	}{
		{"empty", "", ErrEmptyExpression},
		{"whitespace only", "   ", ErrEmptyExpression},
		{"parse error", "1+", ErrInvalidExpression},
		{"unclosed paren", "(1+1", ErrUnclosedParenthesis},
		{"extra close paren", "1+1)", ErrUnclosedParenthesis},
		{"function call unclosed", "sqrt(16", ErrUnclosedParenthesis},
		{"unknown function", "foo(1)", ErrInvalidExpression},
		{"division by zero int", "1/0", ErrDivisionByZero},
		{"division by zero float", "1.0/0.0", ErrDivisionByZero},
		{"division by zero in expression", "5 + 10/0", ErrDivisionByZero},
		{"lone percent", "50%", ErrInvalidExpression},
		{"zero over zero", "0/0", ErrDivisionByZero},
		{"sqrt of negative", "sqrt(-1)", ErrDivisionByZero},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Calculate(tc.expr)
			if !errors.Is(err, tc.want) {
				t.Errorf("Calculate(%q) err = %v, want %v", tc.expr, err, tc.want)
			}
		})
	}
}

func TestFloatifyIntegerLiterals(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"1+1", "1.0+1.0"},
		{"1.5+2", "1.5+2.0"},
		{"sqrt(16)", "sqrt(16.0)"},
		{"(1+2)*3", "(1.0+2.0)*3.0"},
		{"0.5", "0.5"},
		{"x2 + 3", "x2 + 3.0"}, // x2 is identifier, not number
	}
	for _, tc := range tests {
		if got := floatifyIntegerLiterals(tc.in); got != tc.want {
			t.Errorf("floatifyIntegerLiterals(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
