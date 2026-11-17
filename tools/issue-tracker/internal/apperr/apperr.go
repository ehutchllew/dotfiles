// Package apperr defines errors with stable machine-readable codes.
package apperr

import "fmt"

// Error is a failure reported to the caller as JSON.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Usage   bool           `json:"-"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// New returns an error with a code and message.
func New(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Usage returns an error for bad arguments or input.
func Usage(code, format string, args ...any) *Error {
	e := New(code, format, args...)
	e.Usage = true
	return e
}

// With adds a detail and returns the error.
func (e *Error) With(key string, value any) *Error {
	if e.Details == nil {
		e.Details = map[string]any{}
	}
	e.Details[key] = value
	return e
}
