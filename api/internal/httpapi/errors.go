package httpapi

import (
	"errors"
	"net/http"

	"github.com/andrecalil/go-calc/api/internal/calc"
)

// errorResponse captures everything the handler needs to know about a failure
// in one place: what to log (err), what status to send, and what user-facing
// message to put in the JSON body.
type errorResponse struct {
	status  int
	message string
}

// mapError converts a calc-layer error into an HTTP response spec. Unknown
// errors become 500 with a generic message so we don't leak internals.
func mapError(err error) errorResponse {
	switch {
	case errors.Is(err, calc.ErrEmptyExpression):
		return errorResponse{http.StatusBadRequest, "expression is empty"}
	case errors.Is(err, calc.ErrUnclosedParenthesis):
		// Recommended-fix message — tailored guidance instead of the
		// generic "invalid expression" so the user knows what to correct.
		return errorResponse{
			http.StatusBadRequest,
			"Please review your equation, it seems there's a missing parenthesis",
		}
	case errors.Is(err, calc.ErrInvalidExpression):
		return errorResponse{http.StatusBadRequest, "invalid expression"}
	case errors.Is(err, calc.ErrDivisionByZero):
		return errorResponse{http.StatusUnprocessableEntity, "division by zero"}
	case errors.Is(err, calc.ErrNonNumericResult):
		return errorResponse{http.StatusUnprocessableEntity, "expression did not produce a numeric result"}
	default:
		return errorResponse{http.StatusInternalServerError, "internal server error"}
	}
}
