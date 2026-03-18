// File: internal/clipboard/storage_test.go
package clipboard

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestStorageNew tests creating a new storage instance
func TestStorageNew(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	if storage == nil {
		t.Fatal("NewStorage returned nil")
	}
	
	// Verify storage path is set
	path := storage.GetPath()
	if path == "" {
		t.Error("Storage path should not be empty")
	}
	
	// Verify directory was created
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Storage directory should exist")
	}
}

// TestStorageNewEmptyPath tests creating storage with empty path (uses home directory)
func TestStorageNewEmptyPath(t *testing.T) {
	storage, err := NewStorage("")
	if err != nil {
		t.Fatalf("NewStorage with empty path failed: %v", err)
	}
	
	if storage == nil {
		t.Fatal("NewStorage returned nil")
	}
	
	path := storage.GetPath()
	if path == "" {
		t.Error("Storage path should not be empty")
	}
}

// TestStorageLoad tests loading empty history
func TestStorageLoad(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Load from non-existent file
	items, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(items) != 0 {
		t.Errorf("Expected 0 items from new storage, got %d", len(items))
	}
}

// TestStorageSaveAndLoad tests saving and loading items
func TestStorageSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Create test items
	items := []Item{
		{
			ID:        "1",
			Type:      TypeText,
			Content:   "Hello World",
			Timestamp: time.Now(),
			Pinned:    false,
			CopyCount: 1,
		},
		{
			ID:        "2",
			Type:      TypeImage,
			ImageData: "base64encodedimage",
			ImageType: "png",
			Timestamp: time.Now(),
			Pinned:    true,
			CopyCount: 0,
		},
	}
	
	// Save items
	err = storage.Save(items)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	// Load items
	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	// Verify count
	if len(loadedItems) != 2 {
		t.Errorf("Expected 2 items, got %d", len(loadedItems))
	}
	
	// Verify first item
	if loadedItems[0].ID != "1" {
		t.Errorf("First item ID = %v; want 1", loadedItems[0].ID)
	}
	if loadedItems[0].Content != "Hello World" {
		t.Errorf("First item Content = %v; want Hello World", loadedItems[0].Content)
	}
	if loadedItems[0].Pinned != false {
		t.Error("First item should not be pinned")
	}
	
	// Verify second item
	if loadedItems[1].ID != "2" {
		t.Errorf("Second item ID = %v; want 2", loadedItems[1].ID)
	}
	if loadedItems[1].Type != TypeImage {
		t.Errorf("Second item Type = %v; want TypeImage", loadedItems[1].Type)
	}
	if loadedItems[1].Pinned != true {
		t.Error("Second item should be pinned")
	}
}

// TestStorageSaveEmpty tests saving empty items
func TestStorageSaveEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Save empty items
	err = storage.Save([]Item{})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	// Load should return empty
	items, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

// TestStorageClear tests clearing storage
func TestStorageClear(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Save some items first
	items := []Item{
		{ID: "1", Type: TypeText, Content: "Test"},
	}
	err = storage.Save(items)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	// Clear storage
	err = storage.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	
	// Load should return empty
	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(loadedItems) != 0 {
		t.Errorf("Expected 0 items after clear, got %d", len(loadedItems))
	}
}

// TestStorageConcurrentSaveAndLoad tests concurrent save and load operations
func TestStorageConcurrentSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Run concurrent saves and loads
	done := make(chan bool)
	
	// Concurrent save
	go func() {
		for i := 0; i < 10; i++ {
			items := []Item{
				{ID: "test", Type: TypeText, Content: "concurrent test"},
			}
			storage.Save(items)
		}
		done <- true
	}()
	
	// Concurrent load
	go func() {
		for i := 0; i < 10; i++ {
			storage.Load()
		}
		done <- true
	}()
	
	<-done
	<-done
	
	// Test passes if no race conditions
}

// TestStorageGetPath tests getting storage path
func TestStorageGetPath(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	path := storage.GetPath()
	if path == "" {
		t.Error("GetPath should not return empty string")
	}
	
	// Verify path contains expected directory
	expectedDir := ".fyclip"
	if filepath.Dir(path) != tmpDir && !contains(path, expectedDir) {
		t.Errorf("Path should contain %s, got %s", expectedDir, path)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestStorageMultipleSaves tests multiple save operations
func TestStorageMultipleSaves(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Save multiple times with different data
	for i := 0; i < 5; i++ {
		items := []Item{
			{ID: string(rune('a' + i)), Type: TypeText, Content: string(rune('A' + i))},
		}
		err := storage.Save(items)
		if err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}
	}
	
	// Load should have the last saved item
	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(loadedItems) != 1 {
		t.Errorf("Expected 1 item, got %d", len(loadedItems))
	}
	
	// Should be the last item saved (id = 'e')
	if loadedItems[0].ID != "e" {
		t.Errorf("Expected last item with ID 'e', got %s", loadedItems[0].ID)
	}
}

// TestStorageLargeContent tests storing large content
func TestStorageLargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Create large content
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "abcdefghij"
	}
	
	items := []Item{
		{ID: "large", Type: TypeText, Content: largeContent},
	}
	
	err = storage.Save(items)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(loadedItems) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(loadedItems))
	}
	
	if loadedItems[0].Content != largeContent {
		t.Error("Large content mismatch")
	}
}

// TestStorageSpecialCharacters tests storing special characters
func TestStorageSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Test various special characters
	specialContent := "Hello 世界 🌍 🎉\n\t\r\"'\\"
	
	items := []Item{
		{ID: "special", Type: TypeText, Content: specialContent},
	}
	
	err = storage.Save(items)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if loadedItems[0].Content != specialContent {
		t.Errorf("Special content mismatch:\nExpected: %q\nGot: %q", specialContent, loadedItems[0].Content)
	}
}

// TestEncryptDecrypt tests the encryption functions directly
func TestEncryptDecrypt(t *testing.T) {
	// Create a test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	
	plaintext := []byte("Hello, World! This is a test message.")
	
	// Encrypt
	ciphertext, err := encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Decrypt
	decrypted, err := decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	
	// Verify
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted content = %q; want %q", string(decrypted), string(plaintext))
	}
}

// TestEncryptDecryptEmpty tests encryption with empty data
func TestEncryptDecryptEmpty(t *testing.T) {
	key := make([]byte, 32)
	
	// Encrypt empty
	ciphertext, err := encrypt([]byte(""), key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	decrypted, err := decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	
	if string(decrypted) != "" {
		t.Error("Empty encryption should produce empty result")
	}
}

// TestEncryptInvalidKey tests encryption with invalid key
func TestEncryptInvalidKey(t *testing.T) {
	plaintext := []byte("test")
	
	// Short key (should fail)
	shortKey := []byte("short")
	_, err := encrypt(plaintext, shortKey)
	if err == nil {
		t.Error("Should fail with short key")
	}
}

// TestDecryptInvalidCiphertext tests decryption with invalid ciphertext
func TestDecryptInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	
	// Empty ciphertext
	_, err := decrypt([]byte(""), key)
	if err == nil {
		t.Error("Should fail with empty ciphertext")
	}
	
	// Too short ciphertext
	_, err = decrypt([]byte("short"), key)
	if err == nil {
		t.Error("Should fail with too short ciphertext")
	}
	
	// Random invalid data
	_, err = decrypt([]byte("this is definitely not encrypted data at all!!!"), key)
	if err == nil {
		t.Error("Should fail with random data")
	}
}

// TestStorageOverwrite tests overwriting storage
func TestStorageOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	
	// Save first item
	err = storage.Save([]Item{{ID: "1", Type: TypeText, Content: "First"}})
	if err != nil {
		t.Fatalf("First save failed: %v", err)
	}
	
	// Overwrite with second item
	err = storage.Save([]Item{{ID: "2", Type: TypeText, Content: "Second"}})
	if err != nil {
		t.Fatalf("Second save failed: %v", err)
	}
	
	// Load should have only second item
	items, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if len(items) != 1 || items[0].ID != "2" {
		t.Errorf("Expected only second item, got %d items", len(items))
	}
}
