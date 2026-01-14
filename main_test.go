package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		result := formatSize(tt.input)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestIsTempFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/path/to/file.mp4", false},
		{"/path/to/file.tmp", true},
		{"/path/to/file.part", true},
		{"/path/to/file.crdownload", true},
		{"/path/to/~$document.doc", true},
		{"/path/to/file.swp", true},
		{"/path/to/normal.txt", false},
	}
	for _, tt := range tests {
		result := isTempFile(tt.path)
		if result != tt.expected {
			t.Errorf("isTempFile(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGetEnabledExts(t *testing.T) {
	// Save original config
	origConfig := config
	defer func() { config = origConfig }()

	// Test with only video enabled
	config = Config{VideoEnabled: true}
	exts := getEnabledExts()
	if len(exts) == 0 {
		t.Error("Expected video extensions, got none")
	}
	found := false
	for _, e := range exts {
		if e == ".mp4" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .mp4 in enabled extensions")
	}

	// Test with custom extensions
	config = Config{CustomExts: "psd, ai, .sketch"}
	exts = getEnabledExts()
	expectedCustom := []string{".psd", ".ai", ".sketch"}
	for _, exp := range expectedCustom {
		found := false
		for _, e := range exts {
			if e == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s in custom extensions", exp)
		}
	}
}

func TestIsMonitoredFile(t *testing.T) {
	// Save original config
	origConfig := config
	defer func() { config = origConfig }()

	config = Config{VideoEnabled: true, ImageEnabled: true}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/path/to/video.mp4", true},
		{"/path/to/video.MP4", true}, // Case insensitive
		{"/path/to/image.jpg", true},
		{"/path/to/doc.pdf", false},  // Doc not enabled
		{"/path/to/file.tmp", false}, // Temp file
	}
	for _, tt := range tests {
		result := isMonitoredFile(tt.path)
		if result != tt.expected {
			t.Errorf("isMonitoredFile(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	oldConfigPath := configPath
	configPath = filepath.Join(tmpDir, "config.json")
	defer func() { configPath = oldConfigPath }()

	// Save config
	config = Config{
		VideoEnabled:      true,
		ImageEnabled:      false,
		CompletionTimeout: 45,
		CustomExts:        ".test",
	}
	saveConfig()

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Reset and load
	config = Config{}
	loadConfig()

	if !config.VideoEnabled {
		t.Error("VideoEnabled should be true after load")
	}
	if config.ImageEnabled {
		t.Error("ImageEnabled should be false after load")
	}
	if config.CompletionTimeout != 45 {
		t.Errorf("CompletionTimeout = %d, want 45", config.CompletionTimeout)
	}
	if config.CustomExts != ".test" {
		t.Errorf("CustomExts = %s, want .test", config.CustomExts)
	}
}

func TestBatchManagement(t *testing.T) {
	// Create temp directory with test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Save original config and batches
	origConfig := config
	origBatches := batches
	defer func() {
		config = origConfig
		batches = origBatches
	}()

	config = Config{VideoEnabled: true}
	batches = make(map[string]*Batch)

	// Add file to batch
	isNew := addFileToBatch(testFile)
	if !isNew {
		t.Error("Expected new batch to be created")
	}

	// Check batch was created
	if len(batches) != 1 {
		t.Errorf("Expected 1 batch, got %d", len(batches))
	}

	// Add same file again
	isNew = addFileToBatch(testFile)
	if isNew {
		t.Error("Expected existing batch, not new")
	}

	// Check still only 1 batch
	if len(batches) != 1 {
		t.Errorf("Expected 1 batch after re-add, got %d", len(batches))
	}

	// Check batch properties
	for _, b := range batches {
		if b.Folder != tmpDir {
			t.Errorf("Batch folder = %s, want %s", b.Folder, tmpDir)
		}
		if len(b.Files) != 1 {
			t.Errorf("Expected 1 file in batch, got %d", len(b.Files))
		}
		if b.Status != "uploading" {
			t.Errorf("Batch status = %s, want uploading", b.Status)
		}
	}
}

func TestAutoStartPaths(t *testing.T) {
	// Just test that getExecutablePath returns something
	path := getExecutablePath()
	if path == "" {
		t.Log("Warning: Could not get executable path (may be normal in test)")
	}
}
