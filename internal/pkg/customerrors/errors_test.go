package customerrors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := NewAppError("TEST_ERROR", "test message", http.StatusBadRequest)
	if err.Error() != "test message" {
		t.Errorf("Expected 'test message', got '%s'", err.Error())
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := NewAppError("TEST_ERROR", "test message", http.StatusBadRequest)
	err2 := NewAppError("TEST_ERROR", "different message", http.StatusInternalServerError)
	err3 := NewAppError("OTHER_ERROR", "test message", http.StatusBadRequest)

	if !errors.Is(err1, err2) {
		t.Error("Expected errors with same code to match")
	}

	if errors.Is(err1, err3) {
		t.Error("Expected errors with different codes to not match")
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := ErrValidation.WithDetails("field is required")
	if err.Details != "field is required" {
		t.Errorf("Expected details 'field is required', got '%s'", err.Details)
	}
	if err.Code != ErrCodeValidation {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeValidation, err.Code)
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteError(recorder, ErrNotFound)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	body := recorder.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

func TestWriteError_RegularError(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteError(recorder, errors.New("some regular error"))

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestWriteValidationError(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteValidationError(recorder, "field 'id' is required")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
