// Package shtypes contains shared JSON request and response types for all endpoint packages
package shtypes

import (
	"github.com/google/uuid"
)

// UserError is a response type for user errors (i.e. not server errors)
type UserError struct {
	Error string `json:"error"`
}

// ServerError is a response type for system errors, containing a UUID which should be mirrored in logs
type ServerError struct {
	Reference string `json:"error_reference"`
}

// NewServerError creates a new ServerError instance with a UUID provided
func NewServerError() ServerError {
	return ServerError{
		Reference: uuid.New().String(),
	}
}

// NewUserError creates a UserError with a given message
func NewUserError(message string) UserError {
	return UserError{
		Error: message,
	}
}
