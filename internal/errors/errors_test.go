// File: internal/errors/errors_test.go
package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	// Test all error codes are defined
	codes := []ErrorCode{
		ErrCodeClipboardInit,
		ErrCodeClipboardRead,
		ErrCodeClipboardWrite,
		ErrCodeClipboardUnavailable,
		ErrCodeStorageLoad,
		ErrCodeStorageSave,
		ErrCodeStorageEncrypt,
		ErrCodeStorageDecrypt,
		ErrCodeManagerAdd,
		ErrCodeManagerDelete,
		ErrCodeManagerNotFound,
		ErrCodeManagerInvalid,
		ErrCodeConfigLoad,
		ErrCodeConfigSave,
		ErrCodeConfigInvalid,
		ErrCodeBackupCreate,
		ErrCodeBackupRestore,
		ErrCodeBackupPassword,
		ErrCodeSnippetAdd,
		ErrCodeSnippetNotFound,
		ErrCodeExclusionInvalid,
		ErrCodeFileNotFound,
		ErrCodeFileOpen,
	}

	if len(codes) != 23 {
		t.Errorf("expected 23 error codes, got %d", len(codes))
	}
}

func TestClipboardError(t *testing.T) {
	// Test error without underlying error
	err := &ClipboardError{
		Code:    ErrCodeClipboardRead,
		Message: "failed to read clipboard",
		Err:     nil,
	}

	if err.Error() != "CLIPBOARD_READ: failed to read clipboard" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	// Test error with underlying error
	innerErr := errors.New("original error")
	errWithInner := &ClipboardError{
		Code:    ErrCodeClipboardWrite,
		Message: "failed to write",
		Err:     innerErr,
	}

	expected := "CLIPBOARD_WRITE: failed to write - original error"
	if errWithInner.Error() != expected {
		t.Errorf("expected %q, got %q", expected, errWithInner.Error())
	}

	// Test Unwrap
	if errWithInner.Unwrap() != innerErr {
		t.Error("Unwrap() should return the underlying error")
	}
}

func TestNewClipboardError(t *testing.T) {
	err := NewClipboardError(ErrCodeClipboardInit, "initialization failed")
	
	if err.Code != ErrCodeClipboardInit {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeClipboardInit)
	}
	if err.Message != "initialization failed" {
		t.Errorf("Message = %s, want 'initialization failed'", err.Message)
	}
	if err.Err != nil {
		t.Error("Err should be nil")
	}
}

func TestWrapError(t *testing.T) {
	// Test wrapping nil error
	err := WrapError(nil, ErrCodeConfigLoad, "config issue")
	if err.Code != ErrCodeConfigLoad {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeConfigLoad)
	}
	if err.Message != "config issue" {
		t.Errorf("Message = %s, want 'config issue'", err.Message)
	}

	// Test wrapping actual error
	original := fmt.Errorf("original error")
	wrapped := WrapError(original, ErrCodeStorageSave, "save failed")
	
	if wrapped.Code != ErrCodeStorageSave {
		t.Errorf("Code = %v, want %v", wrapped.Code, ErrCodeStorageSave)
	}
	if wrapped.Err != original {
		t.Error("wrapped error should contain original error")
	}
}

func TestClipboardErrorIs(t *testing.T) {
	err1 := &ClipboardError{Code: ErrCodeClipboardRead, Message: "test1"}
	err2 := &ClipboardError{Code: ErrCodeClipboardRead, Message: "test2"}
	err3 := &ClipboardError{Code: ErrCodeClipboardWrite, Message: "test3"}

	// Same codes should match
	if !err1.Is(err2) {
		t.Error("errors with same code should match")
	}

	// Different codes should not match
	if err1.Is(err3) {
		t.Error("errors with different codes should not match")
	}

	// Matching against non-ClipboardError
	if err1.Is(errors.New("regular error")) {
		t.Error("should not match regular errors")
	}
}

func TestClipboardErrorIsCode(t *testing.T) {
	err := &ClipboardError{Code: ErrCodeClipboardRead, Message: "test"}

	if !err.IsCode(ErrCodeClipboardRead) {
		t.Error("IsCode should return true for matching code")
	}

	if err.IsCode(ErrCodeClipboardWrite) {
		t.Error("IsCode should return false for non-matching code")
	}
}

func TestClipboardErrorWithError(t *testing.T) {
	original := &ClipboardError{
		Code:    ErrCodeClipboardRead,
		Message: "original message",
	}

	innerErr := fmt.Errorf("inner error")
	enhanced := original.WithError(innerErr)

	if enhanced.Code != original.Code {
		t.Errorf("Code = %v, want %v", enhanced.Code, original.Code)
	}
	if enhanced.Message != original.Message {
		t.Errorf("Message = %s, want %s", enhanced.Message, original.Message)
	}
	if enhanced.Err != innerErr {
		t.Error("WithError should add the inner error")
	}
}

func TestNew(t *testing.T) {
	err := New(ErrCodeManagerNotFound, "item not found")
	
	// Should return *ClipboardError
	ce, ok := err.(*ClipboardError)
	if !ok {
		t.Fatal("New should return *ClipboardError")
	}
	
	if ce.Code != ErrCodeManagerNotFound {
		t.Errorf("Code = %v, want %v", ce.Code, ErrCodeManagerNotFound)
	}
	if ce.Message != "item not found" {
		t.Errorf("Message = %s, want 'item not found'", ce.Message)
	}
}

func TestErrorf(t *testing.T) {
	err := Errorf(ErrCodeSnippetAdd, "failed to add snippet #%d", 42)
	
	// Should return *ClipboardError
	ce, ok := err.(*ClipboardError)
	if !ok {
		t.Fatal("Errorf should return *ClipboardError")
	}
	
	if ce.Code != ErrCodeSnippetAdd {
		t.Errorf("Code = %v, want %v", ce.Code, ErrCodeSnippetAdd)
	}
	
	expected := "failed to add snippet #42"
	if ce.Message != expected {
		t.Errorf("Message = %s, want %s", ce.Message, expected)
	}
}
