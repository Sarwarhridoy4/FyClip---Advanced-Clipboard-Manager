// File: internal/clipboard/storage.go
package clipboard

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const historyFileName = "clipboard_history.json"
const encryptionKeyFileName = "encryption.key" // New constant

// Storage handles persistence of clipboard history
type Storage struct {
	mu       sync.RWMutex
	filePath string
	key      []byte // New field for encryption key
}

// NewStorage creates a new storage instance
func NewStorage(basePath string) (*Storage, error) {
	if basePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		basePath = homeDir
	}

	// Create directory if it doesn't exist
	configDir := filepath.Join(basePath, ".fyclip")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	s := &Storage{
		filePath: filepath.Join(configDir, historyFileName),
	}

	// --- Encryption Key Handling ---
	keyPath := filepath.Join(configDir, encryptionKeyFileName)
	key, err := loadOrCreateKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load or create encryption key: %w", err)
	}
	s.key = key
	// --- End Encryption Key Handling ---

	return s, nil
}

// loadOrCreateKey loads the encryption key from path or creates a new one
func loadOrCreateKey(keyPath string) ([]byte, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Key file does not exist, create a new one
		key := make([]byte, 32) // AES-256 key
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil { // Permissions 0600 for owner only
			return nil, fmt.Errorf("failed to save encryption key: %w", err)
		}
		return key, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat encryption key file: %w", err)
	}

	// Key file exists, load it
	encodedKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}
	key, err := hex.DecodeString(string(encodedKey))
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key length: %d bytes, expected 32", len(key))
	}
	return key, nil
}

// Load reads clipboard history from disk
func (s *Storage) Load() ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encryptedData, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	// Decrypt data
	decryptedData, err := decrypt(encryptedData, s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt history: %w", err)
	}

	// Decompress data if compressed
	decompressedData, err := decompress(decryptedData)
	if err != nil {
		// If decompression fails, try using the data as-is (for backward compatibility)
		decompressedData = decryptedData
	}

	var items []Item
	if err := json.Unmarshal(decompressedData, &items); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	// Generate thumbnails for items that don't have them (backward compatibility)
	for i := range items {
		items[i].EnsureThumbnail()
	}

	return items, nil
}

// Save writes clipboard history to disk
func (s *Storage) Save(items []Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Compress data before encryption
	compressedData := compress(jsonData)

	// Encrypt data
	encryptedData, err := encrypt(compressedData, s.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt history: %w", err)
	}

	// Write to temp file first
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, encryptedData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Cleanup
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetPath returns the storage file path
func (s *Storage) GetPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePath
}

// GetDir returns the directory path for storage
func (s *Storage) GetDir() string {
	return filepath.Dir(s.filePath)
}

// Clear removes all stored history
func (s *Storage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove history file: %w", err)
	}

	return nil
}

// encrypt encrypts data using AES-256 GCM.
// The nonce is prepended to the ciphertext.
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256 GCM.
// The nonce is read from the beginning of the ciphertext.
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// compress compresses data using gzip
// Returns the original data if compression fails (for safety)
func compress(data []byte) []byte {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return data // Return original if compression fails
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return data
	}

	err = writer.Close()
	if err != nil {
		return data
	}

	return buf.Bytes()
}

// decompress decompresses gzip data
// Returns original data if decompression fails (for backward compatibility)
func decompress(data []byte) ([]byte, error) {
	// Check if data is gzip compressed (magic number)
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, fmt.Errorf("not gzip compressed")
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return data, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
