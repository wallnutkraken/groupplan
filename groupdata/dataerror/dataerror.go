// Package dataerror contains error types for specific data-related errors, the messages of which should be returned to the user
package dataerror

// BaseError is the base error struct for the dataerror package, to be embedded in every other error
type BaseError struct {
	Message string
}

// ErrBasic returns BaseError with the message provided
func ErrBasic(message string) error {
	return BaseError{
		Message: message,
	}
}

// Error returns the error message for this error
func (b BaseError) Error() string {
	return b.Message
}

// NotFound represents an error where data that was requested could not be found
type NotFound struct {
	BaseError
}

// ErrNotFound returns a NotFound error with the given message
func ErrNotFound(message string) error {
	return NotFound{
		BaseError: BaseError{
			Message: message,
		},
	}
}
