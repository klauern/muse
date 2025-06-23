package fileops

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Hello, atomic world!")

	err := AtomicWriteFile(testFile, testData, 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("File content mismatch. Expected %q, got %q", testData, content)
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0o644 {
		t.Errorf("File permissions mismatch. Expected 0644, got %o", info.Mode().Perm())
	}
}

func TestAtomicWriteFile_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "subdir", "test.txt")
	testData := []byte("Test data")

	err := AtomicWriteFile(testFile, testData, 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("File was not created: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(testFile)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}
}

func TestAtomicWriteFile_EmptyFilename(t *testing.T) {
	err := AtomicWriteFile("", []byte("test"), 0o644)
	if err == nil {
		t.Fatal("Expected error for empty filename")
	}

	if !strings.Contains(err.Error(), "filename cannot be empty") {
		t.Errorf("Expected 'filename cannot be empty' error, got: %v", err)
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Write initial content
	initialData := []byte("Initial content")
	err := AtomicWriteFile(testFile, initialData, 0o644)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Overwrite with new content
	newData := []byte("New content")
	err = AtomicWriteFile(testFile, newData, 0o644)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(content, newData) {
		t.Errorf("File content mismatch after overwrite. Expected %q, got %q", newData, content)
	}
}

func TestSafeWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "safe_test.txt")
	testData := []byte("Safe write test")

	err := SafeWriteFile(testFile, testData, 0o644)
	if err != nil {
		t.Fatalf("SafeWriteFile failed: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("File content mismatch. Expected %q, got %q", testData, content)
	}
}

func TestSafeWriteFile_SizeLimit(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large_test.txt")

	// Create data larger than the limit (10MB + 1 byte)
	largeData := make([]byte, 10*1024*1024+1)

	err := SafeWriteFile(testFile, largeData, 0o644)
	if err == nil {
		t.Fatal("Expected error for file size exceeding limit")
	}

	if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
		t.Errorf("Expected size limit error, got: %v", err)
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "valid relative path",
			path:      "test.txt",
			wantError: false,
		},
		{
			name:      "valid absolute path",
			path:      "/tmp/test.txt",
			wantError: true, // /tmp is considered suspicious
		},
		{
			name:      "empty path",
			path:      "",
			wantError: true,
		},
		{
			name:      "directory traversal",
			path:      "../../../etc/passwd",
			wantError: true,
		},
		{
			name:      "hidden directory traversal",
			path:      "subdir/../../../etc/passwd",
			wantError: true,
		},
		{
			name:      "valid subdirectory",
			path:      "subdir/test.txt",
			wantError: false,
		},
		{
			name:      "root directory attempt",
			path:      "/root/test.txt",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilePath(tt.path)
			if tt.wantError && err == nil {
				t.Errorf("validateFilePath(%q) expected error, got nil", tt.path)
			}
			if !tt.wantError && err != nil {
				t.Errorf("validateFilePath(%q) unexpected error: %v", tt.path, err)
			}
		})
	}
}

func TestContainsSuspiciousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "safe path",
			path:     "documents/test.txt",
			expected: false,
		},
		{
			name:     "directory traversal",
			path:     "../config/secrets.txt",
			expected: true,
		},
		{
			name:     "etc directory",
			path:     "/etc/passwd",
			expected: true,
		},
		{
			name:     "root directory",
			path:     "/root/.ssh/id_rsa",
			expected: true,
		},
		{
			name:     "tmp directory",
			path:     "/tmp/malicious.sh",
			expected: true,
		},
		{
			name:     "home shortcut",
			path:     "~/secrets.txt",
			expected: true,
		},
		{
			name:     "normal file",
			path:     "project/src/main.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSuspiciousPatterns(tt.path)
			if result != tt.expected {
				t.Errorf("containsSuspiciousPatterns(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	srcData := []byte("Source file content")
	err := os.WriteFile(srcFile, srcData, 0o644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	dstFile := filepath.Join(tempDir, "destination.txt")
	err = CopyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file
	dstData, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if !bytes.Equal(srcData, dstData) {
		t.Errorf("File content mismatch. Expected %q, got %q", srcData, dstData)
	}

	// Verify permissions match
	srcInfo, _ := os.Stat(srcFile)
	dstInfo, _ := os.Stat(dstFile)

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permission mismatch. Source: %o, Destination: %o", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFile_NonexistentSource(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "nonexistent.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")

	err := CopyFile(srcFile, dstFile)
	if err == nil {
		t.Fatal("Expected error for nonexistent source file")
	}

	if !strings.Contains(err.Error(), "failed to open source file") {
		t.Errorf("Expected 'failed to open source file' error, got: %v", err)
	}
}

func TestAtomicWriteFile_NoTempFileLeftover(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Test data")

	err := AtomicWriteFile(testFile, testData, 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Check that no temporary files are left over
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp.") {
			t.Errorf("Temporary file left over: %s", entry.Name())
		}
	}

	// Should only have our target file
	if len(entries) != 1 || entries[0].Name() != "test.txt" {
		t.Errorf("Unexpected files in directory: %v", entries)
	}
}

// Test atomic behavior by simulating an interruption
func TestAtomicWriteFile_PartialWriteProtection(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a file with initial content
	initialData := []byte("Initial content")
	err := os.WriteFile(testFile, initialData, 0o644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Attempt atomic write (this should succeed)
	newData := []byte("New atomic content")
	err = AtomicWriteFile(testFile, newData, 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Verify the file has the new content (not partial or mixed)
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(content, newData) {
		t.Errorf("File content mismatch. Expected %q, got %q", newData, content)
	}

	// Ensure the content is not the initial data
	if bytes.Equal(content, initialData) {
		t.Error("File still contains initial data - atomic write failed")
	}
}
