// File: internal/clipboard/backup.go
package clipboard

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Backup represents a complete backup of clipboard history
type Backup struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Items     []Item    `json:"items"`
	Checksum  string    `json:"checksum"`
}

// BackupManager handles backup and restore operations
type BackupManager struct {
	storage *Storage
}

// NewBackupManager creates a new backup manager
func NewBackupManager(storage *Storage) *BackupManager {
	return &BackupManager{
		storage: storage,
	}
}

// CreateBackup creates an encrypted backup of clipboard history
func (bm *BackupManager) CreateBackup(password string) ([]byte, error) {
	// Load current items
	items, err := bm.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	// Create backup structure
	backup := Backup{
		Version:   "1.0",
		Timestamp: time.Now(),
		Items:     items,
	}

	// Calculate checksum
	data, err := json.Marshal(backup.Items)
	if err != nil {
		return nil, err
	}
	checksum := sha256.Sum256(data)
	backup.Checksum = hex.EncodeToString(checksum[:])

	// Marshal backup
	backupData, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup: %w", err)
	}

	// If password provided, encrypt with password
	if password != "" {
		return bm.encryptWithPassword(backupData, password)
	}

	// Otherwise, return plain JSON
	return backupData, nil
}

// ExportBackup writes backup to a file
func (bm *BackupManager) ExportBackup(path string, password string) error {
	backupData, err := bm.CreateBackup(password)
	if err != nil {
		return err
	}

	return os.WriteFile(path, backupData, 0644)
}

// ImportBackup restores clipboard history from a backup file
func (bm *BackupManager) ImportBackup(path string, password string, merge bool) error {
	// Read backup file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var backupData []byte

	// If password provided, try to decrypt
	if password != "" {
		backupData, err = bm.decryptWithPassword(data, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt backup (wrong password?): %w", err)
		}
	} else {
		// Try plain JSON first
		backupData = data
	}

	// Parse backup
	var backup Backup
	if err := json.Unmarshal(backupData, &backup); err != nil {
		return fmt.Errorf("failed to parse backup: %w", err)
	}

	// Verify checksum
	dataToCheck, err := json.Marshal(backup.Items)
	if err != nil {
		return err
	}
	checksum := sha256.Sum256(dataToCheck)
	calculatedChecksum := hex.EncodeToString(checksum[:])

	if calculatedChecksum != backup.Checksum {
		return fmt.Errorf("backup checksum mismatch - file may be corrupted")
	}

	if merge {
		// Load existing items and merge
		existing, err := bm.storage.Load()
		if err != nil {
			return err
		}

		// Create a map of existing items by ID
		existingMap := make(map[string]Item)
		for _, item := range existing {
			existingMap[item.ID] = item
		}

		// Add items from backup that don't exist
		for _, item := range backup.Items {
			if _, exists := existingMap[item.ID]; !exists {
				existing = append(existing, item)
			}
		}

		return bm.storage.Save(existing)
	}

	// Replace mode - just save backup items
	return bm.storage.Save(backup.Items)
}

// encryptWithPassword encrypts data using AES-256-GCM with password-derived key
func (bm *BackupManager) encryptWithPassword(data []byte, password string) ([]byte, error) {
	// Derive key from password using simple hash (in production, use PBKDF2)
	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decryptWithPassword decrypts data using AES-256-GCM with password-derived key
func (bm *BackupManager) decryptWithPassword(data []byte, password string) ([]byte, error) {
	// Derive key from password
	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetBackupInfo returns information about a backup file without decrypting
func (bm *BackupManager) GetBackupInfo(path string) (Backup, error) {
	var backup Backup

	data, err := os.ReadFile(path)
	if err != nil {
		return backup, err
	}

	// Try to parse as JSON (without password)
	if err := json.Unmarshal(data, &backup); err != nil {
		// File might be encrypted, return minimal info
		return Backup{
			Version:   "encrypted",
			Timestamp: time.Now(),
		}, nil
	}

	return backup, nil
}
