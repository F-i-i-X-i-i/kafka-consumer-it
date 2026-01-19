package customerrors

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Error codes
const (
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// AppError represents an application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// Is implements errors.Is
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAppError creates a new AppError
func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		HTTPStatus: e.HTTPStatus,
	}
}

// Common errors
var (
	ErrValidation         = NewAppError(ErrCodeValidation, "Validation failed", http.StatusBadRequest)
	ErrNotFound           = NewAppError(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrInternal           = NewAppError(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest         = NewAppError(ErrCodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrUnauthorized       = NewAppError(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden          = NewAppError(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrConflict           = NewAppError(ErrCodeConflict, "Conflict", http.StatusConflict)
	ErrServiceUnavailable = NewAppError(ErrCodeServiceUnavailable, "Service unavailable", http.StatusServiceUnavailable)
	ErrInvalidCommand     = NewAppError(ErrCodeBadRequest, "Invalid command", http.StatusBadRequest)
	ErrImageNotFound      = NewAppError(ErrCodeNotFound, "Image not found", http.StatusNotFound)
	ErrProcessingFailed   = NewAppError(ErrCodeInternal, "Image processing failed", http.StatusInternalServerError)
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error *AppError `json:"error"`
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = ErrInternal.WithDetails(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	json.NewEncoder(w).Encode(ErrorResponse{Error: appErr})
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, details string) {
	WriteError(w, ErrValidation.WithDetails(details))
}
