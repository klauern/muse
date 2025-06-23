package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes data to a file atomically by writing to a temporary file
// and then renaming it to the target file. This prevents partial writes and race conditions.
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	return AtomicWriteFileWithDir(filename, data, perm, "")
}

// AtomicWriteFileWithDir writes data to a file atomically with a custom temporary directory
func AtomicWriteFileWithDir(filename string, data []byte, perm os.FileMode, tempDir string) error {
	// Validate inputs
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Ensure the target directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Determine temporary directory
	if tempDir == "" {
		tempDir = dir
	}

	// Create temporary file in the same directory as the target
	// This ensures the rename operation is atomic (same filesystem)
	tempFile, err := os.CreateTemp(tempDir, filepath.Base(filename)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tempPath := tempFile.Name()

	// Ensure cleanup of temporary file on error
	defer func() {
		if tempFile != nil {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	// Write data to temporary file
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	tempFile = nil // Prevent cleanup defer from trying to close again

	// Set correct permissions
	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %w", err)
	}

	// Atomically move temporary file to target location
	if err := os.Rename(tempPath, filename); err != nil {
		return fmt.Errorf("failed to rename temporary file to target: %w", err)
	}

	return nil
}

// SafeWriteFile provides additional safety checks before writing
func SafeWriteFile(filename string, data []byte, perm os.FileMode) error {
	// Validate file path for security
	if err := validateFilePath(filename); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if data is reasonable size (prevent memory exhaustion)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", len(data), maxFileSize)
	}

	return AtomicWriteFile(filename, data, perm)
}

// validateFilePath performs basic security validation on file paths
func validateFilePath(filename string) error {
	// Check for empty path
	if filename == "" {
		return fmt.Errorf("empty file path")
	}

	// Clean the path
	cleaned := filepath.Clean(filename)

	// Check for directory traversal attempts
	if filepath.IsAbs(cleaned) {
		// Allow absolute paths but check for suspicious patterns
		if containsSuspiciousPatterns(cleaned) {
			return fmt.Errorf("suspicious path pattern detected")
		}
	} else {
		// For relative paths, ensure they don't escape current directory
		if filepath.IsAbs(cleaned) || containsSuspiciousPatterns(cleaned) {
			return fmt.Errorf("suspicious relative path pattern detected")
		}
	}

	return nil
}

// containsSuspiciousPatterns checks for common path traversal patterns
func containsSuspiciousPatterns(path string) bool {
	suspicious := []string{
		"../", "..\\", // Directory traversal
		"/etc/", "\\etc\\", // System directories
		"/root/", "\\root\\",
		// Removed /tmp/ from suspicious patterns as Git hooks legitimately use temp files
		"~", // Home directory shortcut
	}

	for _, pattern := range suspicious {
		if filepath.ToSlash(path) != path {
			// Check both slash styles
			if containsPattern(path, pattern) || containsPattern(filepath.ToSlash(path), pattern) {
				return true
			}
		} else {
			if containsPattern(path, pattern) {
				return true
			}
		}
	}

	return false
}

// containsPattern checks if a path contains a suspicious pattern
func containsPattern(path, pattern string) bool {
	// Simple substring check for now
	// Could be enhanced with more sophisticated pattern matching
	return len(path) >= len(pattern) &&
		(path[:len(pattern)] == pattern ||
			path[len(path)-len(pattern):] == pattern ||
			containsSubstring(path, pattern))
}

// containsSubstring checks for substring occurrence
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CopyFile copies a file atomically from src to dst
func CopyFile(src, dst string) error {
	// Validate paths
	if err := validateFilePath(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := validateFilePath(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create temporary file for atomic copy
	tempFile, err := os.CreateTemp(filepath.Dir(dst), filepath.Base(dst)+".copy.*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for copy: %w", err)
	}

	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	// Copy data
	_, err = io.Copy(tempFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Sync and close
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync copied file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Set permissions to match source
	if err := os.Chmod(tempPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, dst); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
