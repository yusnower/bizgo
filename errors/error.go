package errors

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
)

// BizError represents a business error with a specific error code key,
// an underlying error that provides the detailed message, and location information.
type BizError struct {
	key   string // Unique identifier for the error type
	err   error  // Underlying error with detailed message
	uuid  string //
	stack *stack
}

// Error implements the error interface and returns the error message
// including location information.
func (r *BizError) Error() string {

	return r.key
}

func (r *BizError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, r.key)
			r.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, r.key)
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", r.key)
	}
}

// Is implements error interface Is method for errors.Is support.
// It returns true if the target is also a *BizError with the same key.
func (r *BizError) Is(target error) bool {
	var t *BizError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}

	return r.key == t.key
}

// Unwrap returns the underlying error.
// This enables compatibility with the errors.Is and errors.As functions.
func (r *BizError) Unwrap() error {
	return r.err
}

// captureLocation returns a formatted string with the caller's location information.
// The skip parameter determines how many stack frames to skip.
func captureLocation(skip int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = filepath.Base(fn.Name())
	}

	return fmt.Sprintf("%s/%s:%d", filepath.Base(file), funcName, line)
}
