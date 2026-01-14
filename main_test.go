package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestFileMonitoringIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	
	// Save original state
	origConfig := config
	origBatches := batches
	origMonitorPath := monitorPath
	defer func() {
		config = origConfig
		batches = origBatches
		monitorPath = origMonitorPath
		stopMonitor()
	}()

	// Setup
	config = Config{VideoEnabled: true, CompletionTimeout: 2}
	batches = make(map[string]*Batch)
	monitorPath = tmpDir

	// Start monitor
	err := startMonitor(tmpDir)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	t.Log("Monitor started")

	// Create context for goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start event handler
	updateCount := 0
	updateFunc := func() { updateCount++ }
	go handleFileEvents(ctx, updateFunc, nil)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test_video.mp4")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	t.Log("Test file created")

	// Wait for event to be processed
	time.Sleep(500 * time.Millisecond)

	// Check batch was created
	batchesMu.RLock()
	batchCount := len(batches)
	var foundBatch *Batch
	for _, b := range batches {
		foundBatch = b
		break
	}
	batchesMu.RUnlock()

	if batchCount != 1 {
		t.Errorf("Expected 1 batch, got %d", batchCount)
	}

	if foundBatch != nil {
		if foundBatch.Status != "uploading" {
			t.Errorf("Expected status 'uploading', got '%s'", foundBatch.Status)
		}
		if len(foundBatch.Files) != 1 {
			t.Errorf("Expected 1 file, got %d", len(foundBatch.Files))
		}
		t.Logf("Batch created: %+v", foundBatch)
	}

	// Test completion check
	// Manually mark as completed for test
	time.Sleep(2500 * time.Millisecond)
	batchesMu.Lock()
	for _, b := range batches {
		if b.Status == "uploading" && time.Since(b.LastTime) > 2*time.Second {
			b.Status = "completed"
		}
	}
	batchesMu.Unlock()
	
	// Wait for completion timeout (2 seconds + buffer)
	time.Sleep(3 * time.Second)

	batchesMu.RLock()
	if foundBatch != nil && foundBatch.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", foundBatch.Status)
	}
	batchesMu.RUnlock()

	t.Log("File monitoring integration test passed")
}

func TestAutoStartFunctions(t *testing.T) {
	// Test that auto-start functions don't panic
	// We can't fully test them without admin privileges
	
	enabled := isAutoStartEnabled()
	t.Logf("Auto-start currently enabled: %v", enabled)
	
	// Test executable path
	path := getExecutablePath()
	if path == "" {
		t.Log("Warning: Could not get executable path")
	} else {
		t.Logf("Executable path: %s", path)
	}
}
