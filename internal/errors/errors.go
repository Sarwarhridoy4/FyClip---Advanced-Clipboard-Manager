// File: internal/errors/errors.go
package errors

import (
	"fmt"
)

// ErrorCode represents error categories
type ErrorCode string

const (
	// Clipboard errors
	ErrCodeClipboardInit      ErrorCode = "CLIPBOARD_INIT"
	ErrCodeClipboardRead     ErrorCode = "CLIPBOARD_READ"
	ErrCodeClipboardWrite    ErrorCode = "CLIPBOARD_WRITE"
	ErrCodeClipboardUnavailable ErrorCode = "CLIPBOARD_UNAVAILABLE"

	// Storage errors
	ErrCodeStorageLoad     ErrorCode = "STORAGE_LOAD"
	ErrCodeStorageSave    ErrorCode = "STORAGE_SAVE"
	ErrCodeStorageEncrypt ErrorCode = "STORAGE_ENCRYPT"
	ErrCodeStorageDecrypt ErrorCode = "STORAGE_DECRYPT"

	// Manager errors
	ErrCodeManagerAdd      ErrorCode = "MANAGER_ADD"
	ErrCodeManagerDelete   ErrorCode = "MANAGER_DELETE"
	ErrCodeManagerNotFound ErrorCode = "MANAGER_NOT_FOUND"
	ErrCodeManagerInvalid ErrorCode = "MANAGER_INVALID"

	// Config errors
	ErrCodeConfigLoad     ErrorCode = "CONFIG_LOAD"
	ErrCodeConfigSave    ErrorCode = "CONFIG_SAVE"
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// Backup errors
	ErrCodeBackupCreate   ErrorCode = "BACKUP_CREATE"
	ErrCodeBackupRestore  ErrorCode = "BACKUP_RESTORE"
	ErrCodeBackupPassword ErrorCode = "BACKUP_PASSWORD"

	// Snippet errors
	ErrCodeSnippetAdd    ErrorCode = "SNIPPET_ADD"
	ErrCodeSnippetNotFound ErrorCode = "SNIPPET_NOT_FOUND"

	// Exclusion errors
	ErrCodeExclusionInvalid ErrorCode = "EXCLUSION_INVALID"

	// File errors
	ErrCodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileOpen    ErrorCode = "FILE_OPEN"
)

// ClipboardError represents a clipboard-related error
type ClipboardError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface
func (e *ClipboardError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *ClipboardError) Unwrap() error {
	return e.Err
}

// WithError wraps an existing error with additional context
func (e *ClipboardError) WithError(err error) *ClipboardError {
	return &ClipboardError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// NewClipboardError creates a new ClipboardError
func NewClipboardError(code ErrorCode, message string) *ClipboardError {
	return &ClipboardError{
		Code:    code,
		Message: message,
		Err:     nil,
	}
}

// WrapError wraps an error with a message and code
func WrapError(err error, code ErrorCode, message string) *ClipboardError {
	if err == nil {
		return NewClipboardError(code, message)
	}
	return &ClipboardError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Is checks if the error is of the same code
func (e *ClipboardError) Is(target error) bool {
	if t, ok := target.(*ClipboardError); ok {
		return e.Code == t.Code
	}
	return false
}

// IsCode checks if the error has a specific code
func (e *ClipboardError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// New creates a new error with the given code and message
func New(code ErrorCode, message string) error {
	return &ClipboardError{
		Code:    code,
		Message: message,
	}
}

// Errorf creates a new formatted error
func Errorf(code ErrorCode, format string, args ...interface{}) error {
	return &ClipboardError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
