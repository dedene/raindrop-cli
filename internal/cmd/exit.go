package cmd

import "errors"

const (
	ExitSuccess   = 0
	ExitGeneric   = 1
	ExitUsage     = 2
	ExitAuth      = 3
	ExitNotFound  = 4
	ExitRateLimit = 5
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// ExitCoder interface for errors that know their exit code.
type ExitCoder interface {
	ExitCode() int
}

func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var ee *ExitError
	if errors.As(err, &ee) && ee != nil {
		if ee.Code < 0 {
			return ExitGeneric
		}

		return ee.Code
	}

	// Check if error has ExitCode method (like errfmt.ExitError)
	var ec ExitCoder
	if errors.As(err, &ec) {
		code := ec.ExitCode()
		if code < 0 {
			return ExitGeneric
		}

		return code
	}

	return ExitGeneric
}
