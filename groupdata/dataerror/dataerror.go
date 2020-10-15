// Package dataerror contains error types for specific data-related errors, the messages of which should be returned to the user
package dataerror

import "net/http"

// BaseError is the base error struct for the dataerror package, to be embedded in every other error
type BaseError struct {
	Message string
	Status  int
}

// ErrBasic returns BaseError with the message provided
func ErrBasic(message string) error {
	return BaseError{
		Message: message,
		Status:  http.StatusUnprocessableEntity,
	}
}

// Error returns the error message for this error
func (b BaseError) Error() string {
	return b.Message
}

// StatusCode returns the intended status code for this error
func (b BaseError) StatusCode() int {
	return b.Status
}

// NotFound represents an error where data that was requested could not be found
type NotFound struct {
	BaseError
}

// Unauthorized is the error for unauthorized actions
type Unauthorized struct {
	BaseError
}

// ErrUnauthorized creates a new Unauthorized error
func ErrUnauthorized(message string) error {
	return Unauthorized{
		BaseError: BaseError{
			Message: message,
			Status:  http.StatusUnauthorized,
		},
	}
}

// ErrNotFound returns a NotFound error with the given message
func ErrNotFound(message string) error {
	return NotFound{
		BaseError: BaseError{
			Message: message,
			Status:  http.StatusNotFound,
		},
	}
}
