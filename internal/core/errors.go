package core

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode classifies failures independently from transport and persistence adapters.
type ErrorCode string

const (
	CodeInvalidArgument       ErrorCode = "invalid_argument"
	CodeNotFound              ErrorCode = "not_found"
	CodeConflict              ErrorCode = "conflict"
	CodeStalePrecondition     ErrorCode = "stale_precondition"
	CodeProjectionUnavailable ErrorCode = "projection_unavailable"
	CodePolicyDenied          ErrorCode = "policy_denied"
	CodeInternal              ErrorCode = "internal"
)

// Error is the stable application error returned by the V2 core.
type Error struct {
	Code      ErrorCode
	Operation string
	Field     string
	Message   string
	Err       error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	parts := make([]string, 0, 3)
	if operation := strings.TrimSpace(e.Operation); operation != "" {
		parts = append(parts, operation)
	}
	if field := strings.TrimSpace(e.Field); field != "" {
		parts = append(parts, field)
	}

	message := strings.TrimSpace(e.Message)
	if message != "" {
		parts = append(parts, message)
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	if len(parts) == 0 {
		return string(e.Code)
	}
	return strings.Join(parts, ": ")
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewError creates an application error without an underlying cause.
func NewError(code ErrorCode, operation, message string) *Error {
	return &Error{Code: normalizeErrorCode(code), Operation: operation, Message: message}
}

// NewFieldError creates an application error associated with one input field.
func NewFieldError(code ErrorCode, operation, field, message string) *Error {
	return &Error{
		Code:      normalizeErrorCode(code),
		Operation: operation,
		Field:     field,
		Message:   message,
	}
}

// WrapError creates an application error while retaining the original cause.
func WrapError(code ErrorCode, operation, message string, err error) *Error {
	return &Error{
		Code:      normalizeErrorCode(code),
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// ErrorCodeOf returns the first application error code in an error chain.
func ErrorCodeOf(err error) (ErrorCode, bool) {
	var applicationError *Error
	if !errors.As(err, &applicationError) {
		return "", false
	}
	return applicationError.Code, true
}

// IsErrorCode reports whether an error chain contains an application error with code.
func IsErrorCode(err error, code ErrorCode) bool {
	actual, ok := ErrorCodeOf(err)
	return ok && actual == code
}

func normalizeErrorCode(code ErrorCode) ErrorCode {
	if code == "" {
		return CodeInternal
	}
	return code
}

func preserveApplicationError(operation, message string, err error) error {
	if err == nil {
		return nil
	}
	var applicationError *Error
	if errors.As(err, &applicationError) {
		return err
	}
	return WrapError(CodeInternal, operation, message, err)
}

func missingDependency(operation, field string) error {
	return NewFieldError(
		CodeInvalidArgument,
		operation,
		field,
		fmt.Sprintf("%s is required", field),
	)
}
