package errors

import (
	"fmt"
)

type ErrorType string

const (
	ErrTypeGit        ErrorType = "git"
	ErrTypeConfig     ErrorType = "config"
	ErrTypeAnalysis   ErrorType = "analysis"
	ErrTypeValidation ErrorType = "validation"
	ErrTypeIO         ErrorType = "io"
)

type CadenceError struct {
	Type    ErrorType
	Message string
	Details string
	Wrapped error
}

func NewError(errType ErrorType, message string) *CadenceError {
	return &CadenceError{
		Type:    errType,
		Message: message,
	}
}

func (e *CadenceError) WithDetails(details string) *CadenceError {
	e.Details = details
	return e
}

func (e *CadenceError) Wrap(err error) *CadenceError {
	e.Wrapped = err
	return e
}

func (e *CadenceError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Details)
	}
	if e.Wrapped != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *CadenceError) Is(target error) bool {
	t, ok := target.(*CadenceError)
	return ok && e.Type == t.Type
}

func (e *CadenceError) Unwrap() error {
	return e.Wrapped
}

// Convenience constructors for common error types.

func GitError(message string) *CadenceError {
	return NewError(ErrTypeGit, message)
}

func ConfigError(message string) *CadenceError {
	return NewError(ErrTypeConfig, message)
}

func AnalysisError(message string) *CadenceError {
	return NewError(ErrTypeAnalysis, message)
}

func ValidationError(message string) *CadenceError {
	return NewError(ErrTypeValidation, message)
}

func IOError(message string) *CadenceError {
	return NewError(ErrTypeIO, message)
}
